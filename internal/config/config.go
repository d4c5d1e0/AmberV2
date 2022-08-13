package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var (
	config *Config
)

func LoadConfig() error {
	config = new(Config)

	content, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return err
	}
	return nil
}

//GetReactWatcher return the config ReactWatcher
func GetReactWatcher() *ReactWatcher {
	return &config.ReactWatcher
}

//GetMessageWatcher return the config MessageWatcher
func GetMessageWatcher() *MessageWatcher {
	return &config.MessageWatcher
}

//GetWebhook return the config Webhook
func GetWebhook() *Webhook {
	return &config.Webhook
}

//GetJoinWatcher return the config JoinWatcher
func GetJoinWatcher() *JoinWatcher {
	return &config.JoinWatcher
}

//GetLargeGuild return the config LargeGuild
func GetLargeGuild() *LargeGuild {
	return &config.LargeGuild
}

func OnlineTokens() bool     { return config.OnlineTokens }
func CaptchaKey() string     { return config.CaptchaKey }
func Master() string         { return config.Master }
func MessageToSend() string  { return config.MessageToSend }
func CaptchaService() string { return config.CaptchaService }
