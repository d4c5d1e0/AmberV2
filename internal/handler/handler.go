package handler

type Event string

var (
	NewJoin    Event = "newJoin"
	NewReact   Event = "newReact"
	NewMessage Event = "newMessage"
)

func NewEvent(guildID, channelID string, target string, event Event, user User) *AmberEvent {
	return &AmberEvent{
		Event:     event,
		GuildID:   guildID,
		ChannelID: channelID,
		TargetID:  target,
		User:      user,
	}
}

type AmberEvent struct {
	Event Event
	//used to build referer
	GuildID   string
	ChannelID string
	TargetID  string
	User      User
	//additionnal data
	Data interface{}
}

type User struct {
	Username      string
	Discriminator string
}
