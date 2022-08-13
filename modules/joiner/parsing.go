package joiner

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/bytixo/AmberV2/internal/config"
	"github.com/bytixo/AmberV2/internal/event"
	"github.com/bytixo/AmberV2/internal/handler"
	"github.com/bytixo/AmberV2/internal/logger"
	"github.com/tidwall/gjson"
)

var (
	FetchedUsers []string
	ListMutex    sync.Mutex
)

func parseGuildData(data string) {
	go updateGuildData(data)

	d := gjson.Get(data, "d")
	ops := d.Get("ops").Array()

	if len(ops) == 2 && ops[1].Get("op").String() == "INSERT" {
		var m UserData
		member := ops[1].Get("item.member").String()
		if member == "" {
			return
		}
		err := json.Unmarshal([]byte(member), &m)
		if err != nil {
			logger.Error(err, data)
		}

		joined := m.JoinedAt.Local()

		if time.Since(joined) < time.Second {
			if isInSlice(FetchedUsers, m.User.Id) {
				return
			}
			addToList(m.User.Id)
			logger.Info(fmt.Sprintf("%s#%s joined %v ago (%v) | Member Count: %v\n", m.User.Username, m.User.Discriminator, time.Since(joined), joined.Format("2006/01/02 15:04:05.00000"), d.Get("member_count").Int()))

			e := handler.NewEvent(config.GetJoinWatcher().GuildID, config.GetJoinWatcher().ChannelID, m.User.Id, handler.NewJoin, handler.User{
				Username:      m.User.Username,
				Discriminator: m.User.Discriminator,
			})

			event.Handle(e)
		}
	}
}

func updateGuildData(data string) {
	d := gjson.Get(data, "d")
	guildID := d.Get("guild_id").String()

	GuildMutex.Lock()
	ScrapedGuilds[guildID] = d.Get("online_count").Int()
	GuildMutex.Unlock()
}

func addToList(user string) {
	ListMutex.Lock()
	defer ListMutex.Unlock()
	FetchedUsers = append(FetchedUsers, user)
}

func isInSlice(slice []string, t string) bool {
	for _, j := range slice {
		if j == t {
			return true
		}
	}
	return false
}

type UserData struct {
	User struct {
		Username      string `json:"username"`
		PublicFlags   int    `json:"public_flags"`
		Id            string `json:"id"`
		Discriminator string `json:"discriminator"`
		Avatar        string `json:"avatar"`
	} `json:"user"`
	JoinedAt time.Time `json:"joined_at"`
}
