package joiner

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Config struct {
	//GuildID is the ID of the guild you want to monitor
	GuildID string
	//ChannelID should be a channel visible once you join and visible
	//by the watchers tokens too
	ChannelID string
	//Master is the token used to send the first events
	Master string
}

type Connection struct {
	Conn      *websocket.Conn
	closeChan chan struct{}
}

type Session struct {
	connection *Connection
	Token      string
	Indexes    [][][]int
	mutex      *sync.Mutex

	//GuildID is the ID of the guild you want to monitor
	GuildID string
	//ChannelID should be a channel visible once you join and visible
	//by the watchers tokens too
	ChannelID string
}
type GatewayPayload struct {
	Opcode    int         `json:"op"`
	EventData interface{} `json:"d"`
	EventName string      `json:"t,omitempty"`
}
