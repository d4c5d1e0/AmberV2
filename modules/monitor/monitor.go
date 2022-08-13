package monitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/bytixo/AmberV2/internal/config"
	"github.com/bytixo/AmberV2/internal/discord"
	"github.com/bytixo/AmberV2/internal/logger"
)

var (
	backticks = "```"
)

func Monitor() {
	go func() {
		t := time.NewTicker(10 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				logger.Info("Sending webhook")
				stats := getDMStats()
				webhook := &Webhook{
					Content: "",
					Embeds: []Embeds{
						{
							Description: fmt.Sprintf(":robot: **Auto-DM Recap (%s)**", config.GetWebhook().Campaign),
							Color:       14391705,
							Timestamp:   time.Now(),
							Author: Author{
								Name: "Amber",
							},
							Fields: []Fields{
								{
									Name:   ":white_check_mark: Successful DM's ( last 10 minutes )",
									Value:  fmt.Sprintf("%s%v%s", backticks, stats.SuccessDM, backticks),
									Inline: true,
								},
								{
									Name:   "❌ Failed DM's ( last 10 minutes )",
									Value:  fmt.Sprintf("%s%v%s", backticks, stats.FailedDM, backticks),
									Inline: true,
								},
								{
									Name:   "<:bughunter_2:909102089324593182> Requests Total ( last 10 minutes )",
									Value:  fmt.Sprintf("%s%v%s", backticks, stats.TotalRequests, backticks),
									Inline: false,
								},
								{
									Name:   ":white_check_mark: Alive Tokens",
									Value:  fmt.Sprintf("%s%v%s", backticks, stats.AliveTokens, backticks),
									Inline: true,
								},
								{
									Name:   "❌ Dead tokens",
									Value:  fmt.Sprintf("%s%v%s", backticks, stats.DeadTokens, backticks),
									Inline: true,
								},
							},
						},
					},
					Username: "Amber",
				}
				payload, err := json.Marshal(webhook)
				if err != nil {
					logger.Error(err)
					continue
				}
				err = sendWebhook(payload)
				if err != nil {
					logger.Error(err)
					continue
				}
				flushStats()
			}
		}
	}()
	go func() {
		t := time.NewTicker(24 * time.Hour)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				logger.Info("Sending webhook")
				stats := getPdmStats()
				webhook := &Webhook{
					Content: "",
					Embeds: []Embeds{
						{
							Description: fmt.Sprintf(":robot: **Auto-DM Recap (%s)**", config.GetWebhook().Campaign),
							Color:       14391705,
							Timestamp:   time.Now(),
							Author: Author{
								Name: "Amber",
							},
							Fields: []Fields{
								{
									Name:   ":white_check_mark: Successful DM's ( last 24 hours )",
									Value:  fmt.Sprintf("%s%v%s", backticks, stats.SuccessDM, backticks),
									Inline: true,
								},
								{
									Name:   "❌ Failed DM's ( last 24 hours )",
									Value:  fmt.Sprintf("%s%v%s", backticks, stats.FailedDM, backticks),
									Inline: true,
								},
								{
									Name:   "<:bughunter_2:909102089324593182> Requests Total ( last 24 hours )",
									Value:  fmt.Sprintf("%s%v%s", backticks, stats.TotalRequests, backticks),
									Inline: false,
								},
							},
						},
					},
					Username: "Amber",
				}

				payload, err := json.Marshal(webhook)
				if err != nil {
					logger.Error(err)
					continue
				}
				err = sendWebhook(payload)
				if err != nil {
					logger.Error(err)
					continue
				}
				pflushStats()
			}
		}
	}()
}

func sendWebhook(payload []byte) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", config.GetWebhook().Link, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	//fmt.Println(string(body))
	_ = body
	return nil
}
func pflushStats() {
	atomic.SwapInt32(&discord.PsuccessDM, 0)
	atomic.SwapInt32(&discord.PfailedDM, 0)
	atomic.SwapInt32(&discord.PtotalRequests, 0)

}
func flushStats() {
	atomic.SwapInt32(&discord.SuccessDM, 0)
	atomic.SwapInt32(&discord.FailedDM, 0)
	atomic.SwapInt32(&discord.TotalRequests, 0)

}
func getDMStats() *dmMonitor {
	a := int32(len(discord.GetTokens()))
	d := discord.DeadTokens
	s := discord.SuccessDM
	f := discord.FailedDM

	return &dmMonitor{
		AliveTokens:   a,
		DeadTokens:    d,
		SuccessDM:     s,
		FailedDM:      f,
		TotalRequests: discord.TotalRequests,
	}
}
func getPdmStats() *dmMonitor {
	a := int32(len(discord.GetTokens()))
	d := discord.DeadTokens
	s := discord.PsuccessDM
	f := discord.PfailedDM

	return &dmMonitor{
		AliveTokens:   a,
		DeadTokens:    d,
		SuccessDM:     s,
		FailedDM:      f,
		TotalRequests: discord.PtotalRequests,
	}
}

type dmMonitor struct {
	AliveTokens, DeadTokens, SuccessDM, FailedDM, TotalRequests int32
}
type Webhook struct {
	Content  interface{} `json:"content"`
	Embeds   []Embeds    `json:"embeds"`
	Username string      `json:"username"`
}
type Fields struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}
type Author struct {
	Name string `json:"name"`
}
type Footer struct {
	Text string `json:"text"`
}
type Embeds struct {
	Description string    `json:"description"`
	Color       int       `json:"color"`
	Fields      []Fields  `json:"fields"`
	Author      Author    `json:"author"`
	Footer      Footer    `json:"footer"`
	Timestamp   time.Time `json:"timestamp"`
}
