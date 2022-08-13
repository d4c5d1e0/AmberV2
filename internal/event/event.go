package event

import (
	"fmt"
	"time"

	"github.com/bytixo/AmberV2/internal/database"
	"github.com/bytixo/AmberV2/internal/discord"
	"github.com/bytixo/AmberV2/internal/handler"
	"github.com/bytixo/AmberV2/internal/logger"
)

var (
	MAX_RETRIES = 5
)

func Handle(event *handler.AmberEvent) {
	switch event.Event {
	case handler.NewJoin:
		//handle new join
		logger.Debug(fmt.Sprintf("%+v", event))
		go sendMessage(event)
	case handler.NewReact:
		//handle new reaction
	case handler.NewMessage:
		//handles new message
	}
}

func sendMessage(event *handler.AmberEvent) {
	var retries int
	if database.Proxy.IsBlackListed(event.TargetID) {
		logger.Error(fmt.Sprintf("%s#%s has already been dm'ed, skipping", event.User.Username, event.User.Discriminator))
		return
	}
	start := time.Now()

	logger.Info(fmt.Sprintf("Sending DM to %s%s(%s)", event.User.Username, event.User.Discriminator, event.TargetID))
DMIN:
	for {
		if retries >= MAX_RETRIES {
			logger.Error("Couldn't DM User, too much retries")
			discord.AddFailedDM()
			break
		}
		messager := discord.NewMessager(event)
		err := messager.SendDM()
		if err != nil {
			//TODO: Better error handling
			switch err {
			case discord.ErrCannotDM:
				logger.Error("Couldn't DM User, maybe his dm are closed or the token is not present on the server")
				discord.AddFailedDM()
				break DMIN
			case discord.ErrNotAuthorized:
				logger.Error("Couldn't DM User, invalid token, removing it")
				discord.RemoveToken(messager.Token)
				continue DMIN
			case discord.ErrTokenLocked:
				logger.Error("Couldn't DM User, locked token, removing it")
				discord.RemoveToken(messager.Token)
				continue DMIN
			case discord.ErrCaptcha:
				logger.Error("Couldn't DM User, captcha detected, retrying")
				discord.RemoveToken(messager.Token)
				retries++
				continue DMIN
			case discord.ErrRateLimited:
				//shouldn't happen
				logger.Error("Couldn't DM User, rate limited")
				discord.AddFailedDM()
				break DMIN
			case discord.ErrMemberScreening:
				logger.Error("Couldn't DM User, make sure your tokens have bypassed server verification, that they are on the right level of verification, or have waited the required time for new members")
				discord.AddFailedDM()
				break DMIN
			default:
				logger.Error("Couldn't DM User, unknown error:", err)
				discord.AddFailedDM()
				break DMIN
			}
		}
		logger.Info(fmt.Sprintf("Successfully sent DM to %s%s(%s) in %v", event.User.Username, event.User.Discriminator, event.TargetID, time.Now().Sub(start)))
		discord.AddSuccessDM()
		break
	}
	err := database.Proxy.BlacklistID(event.TargetID)
	if err != nil {
		logger.Error("error blacklisting", event.TargetID, ":", err)
	}
}
