package config

type Config struct {
	Master         string         `yaml:"master"`
	CaptchaKey     string         `yaml:"captcha_key"`
	CaptchaService string         `yaml:"captcha_service"`
	MessageToSend  string         `yaml:"message_to_send"`
	OnlineTokens   bool           `yaml:"online_tokens"`
	Webhook        Webhook        `yaml:"webhook"`
	MessageWatcher MessageWatcher `yaml:"message_watcher"`
	ReactWatcher   ReactWatcher   `yaml:"react_watcher"`
	JoinWatcher    JoinWatcher    `yaml:"join_watcher"`
	LargeGuild     LargeGuild     `yaml:"large_guild"`
}
type Webhook struct {
	Link     string `yaml:"link"`
	Campaign string `yaml:"campaign"`
}
type MessageWatcher struct {
	Enabled bool   `yaml:"enabled"`
	Mode    string `yaml:"mode"`
	Author  string `yaml:"author"`
	Channel string `yaml:"channel"`
}
type Emoji struct {
	Name string `yaml:"name"`
	ID   bool   `yaml:"id"`
}
type ReactWatcher struct {
	Enabled   bool    `yaml:"enabled"`
	MessageID string  `yaml:"message_id"`
	Emoji     []Emoji `yaml:"emoji"`
}
type JoinWatcher struct {
	Enabled   bool   `yaml:"enabled"`
	ChannelID string `yaml:"channel_id"`
	GuildID   string `yaml:"guild_id"`
}
type LargeGuild struct {
	IsLarge   bool   `yaml:"is_large"`
	ChannelID string `yaml:"channel_id"`
	GuildID   string `yaml:"guild_id"`
}
