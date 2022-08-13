package joiner

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/bytixo/AmberV2/internal/discord"
	"github.com/bytixo/AmberV2/internal/logger"
	"github.com/tidwall/gjson"
)

var (
	ScrapedGuilds = map[string]int64{}
	GuildMutex    sync.Mutex
)

func startFetching(tokens []string, guildID, channelID string) {
	// Fetch the online count
	var memberCount int64
	for {
		memberCount = getMemberCount(guildID)
		if memberCount < 100 {
			continue
		}
		break
	}
	logger.Info("Online Count", memberCount)

	tasks := calculateIndexes(tokens, memberCount)

	var wg sync.WaitGroup
	for token, v := range tasks {
		wg.Add(1)
		go func(token string, v [][][]int) {
			defer wg.Done()
			session := NewSession(token)
			//set the target guild
			session.GuildID = guildID
			//set target channel
			session.ChannelID = channelID
			//set indexes
			session.Indexes = v
			//open session
			err := session.Open()
			if err != nil {
				logger.Error("error opening session for", token, err)
				return
			}
			//send all the ranges
			err = session.subscribeToRanges()
			if err != nil {
				logger.Error("error subscribing for", token, err)
				return
			}
			logger.Info(fmt.Sprintf("Started session (%v)", token))
		}(token, v)
	}
	wg.Wait()
	logger.Info("Started Listening")
}
func (s *Session) subscribeToRanges() error {
	time.Sleep(650 * time.Millisecond)
	err := s.SendLazyRequest(s.GuildID, s.ChannelID)
	if err != nil {
		logger.Error("Error: ", err, s.Token)
		return err
	}
	time.Sleep(100 * time.Millisecond)
	err = s.sendSubscribtions()
	if err != nil {
		logger.Error("Error: ", err, s.Token)
		return err
	}
	return nil
}
func (s *Session) sendSubscribtions() error {
	for _, index := range s.Indexes {
		payload, err := json.Marshal(discord.GatewayFetchMembers{
			Opcode: 14,
			EventData: discord.FetchMembersData{
				GuildID: s.GuildID,
				Channels: map[string][][]int{
					s.ChannelID: index,
				},
			},
		})
		if err != nil {
			return err
		}
		time.Sleep(250 * time.Millisecond)
		err = s.SendMessage(payload)
		if err != nil {
			return err
		}
	}
	return nil
}
func calculateIndexes(tokens []string, memberCount int64) map[string][][][]int {
	indexPerToken := map[string][][][]int{}
	n := memberCount / int64(len(tokens))

	// r == number of member per token
	r := int(100 * math.Round(float64(n)/100.0))
	if r > 500 {
		logger.Error("Warning exceeding 500 member per token might reduce efficiency")
	}

	totalRanges := (r) * (len(tokens))
	var counter int
	for i := 0; i < (totalRanges / 100); i++ {

		if i != 0 && i%((r)/100) == 0 {
			counter++
		}

		tokenRange := getRanges(i, 100, int(memberCount))
		if len(tokenRange) != 3 {
			continue
		}
		indexPerToken[tokens[counter]] = append(indexPerToken[tokens[counter]], tokenRange)
	}

	return indexPerToken

}
func getGuildsData(data string) {
	guilds := gjson.Get(data, "d.guilds")
	for _, guild := range guilds.Array() {
		guildID := guild.Get("id").String()
		GuildMutex.Lock()
		ScrapedGuilds[guildID] = 0
		GuildMutex.Unlock()
	}
}

func getMemberCount(guildID string) int64 {
	GuildMutex.Lock()
	defer GuildMutex.Unlock()
	return ScrapedGuilds[guildID]
}
