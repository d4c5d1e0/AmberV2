package main

import (
	"os"

	"github.com/bytixo/AmberV2/internal/config"
	"github.com/bytixo/AmberV2/internal/logger"
	"github.com/bytixo/AmberV2/modules/joiner"
	"github.com/bytixo/AmberV2/modules/monitor"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		logger.Error("Error opening cfg", err)
		os.Exit(1)
	}

	joinWatcher := config.GetJoinWatcher()
	reactWatcher := config.GetReactWatcher()
	messageWatcher := config.GetMessageWatcher()

	if reactWatcher.Enabled {
	}
	if messageWatcher.Enabled {
	}
	if joinWatcher.Enabled {
		cfg := &joiner.Config{
			GuildID:   joinWatcher.GuildID,
			ChannelID: joinWatcher.ChannelID,
			Master:    config.Master(),
		}

		go cfg.Monitor()
	}

	go monitor.Monitor()
	select {}
}
