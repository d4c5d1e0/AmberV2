package joiner

import (
	"bufio"
	"log"
	"os"
	"time"

	"github.com/bytixo/AmberV2/internal/logger"
)

func (config *Config) Monitor() {
	tokens, err := loadWatcherTokens()
	if err != nil {
		log.Fatal(err)
	}

	masterSession := NewSession(config.Master)
	masterSession.ChannelID = config.ChannelID
	masterSession.GuildID = config.GuildID

	err = masterSession.Open()
	if err != nil {
		logger.Error("Error opening session on Master:", err)
		return
	}
	err = masterSession.SendLazyRequest(config.GuildID, config.ChannelID)
	if err != nil {
		logger.Error("Error opening session on Master:", err)
		return
	}

	time.Sleep(250 * time.Millisecond)
	logger.Info("Starting Watchers")
	startFetching(tokens, config.GuildID, config.ChannelID)

}

func loadWatcherTokens() ([]string, error) {
	file, err := os.Open("data/w_tokens.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
