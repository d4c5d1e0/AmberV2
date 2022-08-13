package joiner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/bytixo/AmberV2/internal/logger"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

func NewSession(token string) *Session {
	return &Session{
		connection: &Connection{
			closeChan: make(chan struct{}),
		},
		Token: token,
		mutex: new(sync.Mutex),
	}
}

func (s *Session) Open() error {
	ws, _, err := websocket.DefaultDialer.Dial("wss://gateway.discord.gg/?v=9&encoding=json", nil)
	if err != nil {
		return err
	}
	s.connection.Conn = ws
	err = s.ReadHello()
	if err != nil {
		return err
	}
	go s.receiveIncomingMessages()
	return nil
}

func (s *Session) ReadHello() error {
	_, message, err := s.connection.Conn.ReadMessage()
	if err != nil {
		return err
	}
	opCode := gjson.GetBytes(message, "op").Int()

	if opCode != 10 {
		return fmt.Errorf("Expected op 10 but got %v", string(message))
	}
	interval := gjson.GetBytes(message, "d.heartbeat_interval").Float()
	err = s.identify()
	if err != nil {
		return fmt.Errorf("error identifying: %v", err)
	}
	go s.setupHeartbeat(interval)
	return nil
}
func (s *Session) SendLazyRequest(guildID, channelID string) error {

	payload := []byte(fmt.Sprintf(`{"op":14,"d":{"guild_id":"%s","typing":true,"threads":true,"activities":true,"members":[],"channels":{"%s":[[0,99]]},"thread_member_lists":[]}}`, guildID, channelID))
	return s.SendMessage(payload)
}
func (s *Session) SendMessage(message []byte) error {
	err := s.connection.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return err
	}
	return nil
}
func (s *Session) receiveIncomingMessages() {

	for {
		_, message, err := s.connection.Conn.ReadMessage()
		if err != nil {
			switch {
			case errors.Is(err, websocket.ErrCloseSent):
				s.connection.Conn.Close()
				return
			case strings.Contains(err.Error(), "unexpected EOF"):
				logger.Debug(io.ErrUnexpectedEOF.Error(), s.Token)
				go s.reconnect()
				return
			case strings.Contains(err.Error(), "1000"):
				logger.Debug(err.Error(), s.Token)
				go s.reconnect()
				return
			case strings.Contains(err.Error(), "wsarecv"):
				logger.Debug("Error wsarecv", s.Token)
				go s.reconnect()
				return
			default:
				logger.Error("Error reading message:", err, s.Token)
				s.connection.Conn.Close()
				return
			}

		}
		go s.parseMessage(message)
	}
}
func (s *Session) parseMessage(message []byte) {

	opCode := gjson.GetBytes(message, "op").Int()
	eventName := gjson.GetBytes(message, "t").String()

	switch {

	case opCode == 7:
		logger.Debug("OP 7 Reconnecting")
		go s.reconnect()
	case eventName == "GUILD_MEMBER_LIST_UPDATE":
		go parseGuildData(string(message))
	case eventName == "READY":
		go getGuildsData(string(message))
	case eventName == "READY_SUPPLEMENTAL":
	}
}
func (s *Session) identify() error {

	b := fmt.Sprintf("{\"op\":2,\"d\":{\"token\":\"%s\",\"capabilities\":253,\"properties\":{\"os\":\"Mac OS X\",\"browser\":\"Chrome\",\"device\":\"\",\"system_locale\":\"fr-FR\",\"browser_user_agent\":\"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36\",\"browser_version\":\"98.0.4758.80\",\"os_version\":\"10.12.6\",\"referrer\":\"https://www.google.com/\",\"referring_domain\":\"www.google.com\",\"search_engine\":\"google\",\"referrer_current\":\"\",\"referring_domain_current\":\"\",\"release_channel\":\"stable\",\"client_build_number\":113584,\"client_event_source\":null},\"presence\":{\"status\":\"online\",\"since\":0,\"activities\":[],\"afk\":false},\"compress\":false,\"client_state\":{\"guild_hashes\":{},\"highest_last_message_id\":\"0\",\"read_state_version\":0,\"user_guild_settings_version\":-1,\"user_settings_version\":-1}}}", s.Token)
	err := s.connection.Conn.WriteMessage(websocket.TextMessage, []byte(b))
	if err != nil {
		return err
	}

	return nil
}
func (s *Session) setupHeartbeat(interval float64) {
	go func() {
		t := time.NewTicker((time.Duration(interval) * time.Millisecond))
		defer t.Stop()
		for {
			select {
			case <-s.connection.closeChan:
				return
			case <-t.C:
				b, err := json.Marshal(GatewayPayload{1, nil, ""})
				if err != nil {
					logger.Error(err)
				}

				err = s.connection.Conn.WriteMessage(websocket.TextMessage, b)
				if err != nil {
					logger.Error("Error writing heartbeat: ", err, s.Token)
					go s.reconnect()
					return
				}
			}
		}
	}()
}
func (s *Session) reconnect() {
	logger.Info(fmt.Sprintf("Reconnecting (%v)", s.Token))
	logger.Debug(fmt.Sprintf("Reconnecting (%v)", s.Token))
	s.connection.Close()
	err := s.Open()
	if err != nil {
		logger.Error((fmt.Sprintf(" Error Reconnecting (%v): %v", s.Token, err)))
	}
	logger.Info(fmt.Sprintf("Reconnected token (%v)", s.Token))
	logger.Debug(fmt.Sprintf("Reconnected token (%v)", s.Token))
	err = s.subscribeToRanges()
	if err != nil {
		logger.Error("error subscribing for", s.Token, err)
		return
	}
}
func (s *Session) Connection() *Connection {
	return s.connection
}
func (connection *Connection) Close() error {
	connection.closeChan <- struct{}{}
	err := connection.Conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "going away"), time.Now().Add(time.Second*10))
	if err != nil {
		if connection.Conn != nil {
			connection.Conn.Close()
		}
	}
	return nil
}
