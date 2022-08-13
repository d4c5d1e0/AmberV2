package discord

import (
	"net/http"

	"github.com/bytixo/AmberV2/internal/handler"
)

type Messager struct {
	Token       string
	Fingerprint string

	SnowFlake string
	Event     *handler.AmberEvent

	client *http.Client
	Task   *CaptchaTask
}
type GatewayFetchMembers struct {
	Opcode    int              `json:"op"`
	EventData FetchMembersData `json:"d"`
}

type FetchMembersData struct {
	GuildID  string      `json:"guild_id"`
	Channels interface{} `json:"channels"`
}

type CaptchaTask struct {
	Sitekey string
	Data    string
	Token   string

	Response string
}
