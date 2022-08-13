package captcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bytixo/AmberV2/internal/config"
	"github.com/bytixo/AmberV2/internal/logger"
	"github.com/tidwall/gjson"
)

var (

	// errors
	ErrMaxRetries = errors.New("max retries reached")
)

func NewClient(userAgent, siteKey, targetUrl, apiKey, data string) *CaptchaClient {

	client := &http.Client{
		Timeout: time.Second * 20,
	}

	return &CaptchaClient{
		Service:   parseService(config.CaptchaService()),
		UserAgent: userAgent,
		client:    client,
		SiteKey:   siteKey,
		TargetURL: targetUrl,
		Data:      data,
		APIKey:    apiKey,
	}
}

func parseService(s string) string {
	switch s {
	case "anti-captcha":
		return "api.anti-captcha.com"
	case "capmonster":
		return "api.capmonster.cloud"
	default:
		panic("invalid captcha service")
	}
}

// return valid hcaptcha token or an error
func (c *CaptchaClient) GetCaptcha() (string, error) {
	err := c.createTask()
	if err != nil {
		return "", err
	}

	time.Sleep(time.Second * 2)
	solution, err := c.getTaskResult()
	if err != nil {
		return "", err
	}

	return solution, nil
}

func (c *CaptchaClient) createTask() error {
	var URL = fmt.Sprintf("https://%s/createTask", c.Service)

	task := &HCaptchaProxyTask{
		ClientKey: c.APIKey,
		Task: Task{
			Type:       "HCaptchaTaskProxyless",
			WebsiteURL: c.TargetURL,
			WebsiteKey: c.SiteKey,
			IsInvisble: true,
			UserAgent:  c.UserAgent,
			Data:       c.Data,
		},
	}

	payload, err := json.Marshal(task)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if gjson.Get(string(body), "errorId").Int() != 0 {
		return fmt.Errorf("failed to create task %s, %s", gjson.Get(string(body), "errorCode"), gjson.Get(string(body), "errorDescription"))
	}

	c.TaskID = gjson.Get(string(body), "taskId").Int()
	return nil
}
func (c *CaptchaClient) getTaskResult() (string, error) {

	var tries int

	type SlaveTask struct {
		Status   string
		Solution string
	}

	slaveTask := func() (*SlaveTask, error) {
		URL := fmt.Sprintf("https://%s/createTask", c.Service)

		p := map[string]interface{}{
			"clientKey": c.APIKey,
			"taskId":    c.TaskID,
		}
		payload, err := json.Marshal(p)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", URL, bytes.NewBuffer(payload))
		if err != nil {
			return nil, err
		}

		res, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		if gjson.Get(string(body), "errorId").Int() != 0 {
			return nil, fmt.Errorf("failed to create task %s, %s", gjson.Get(string(body), "errorCode"), gjson.Get(string(body), "errorDescription"))
		}

		if gjson.Get(string(body), "status").String() == "ready" {
			return &SlaveTask{Status: "ready", Solution: gjson.Get(string(body), "solution.gRecaptchaResponse").String()}, nil
		}

		return &SlaveTask{Status: "processing", Solution: ""}, nil
	}

	for {

		if tries == 15 {
			return "", ErrMaxRetries
		}
		t, err := slaveTask()
		if err != nil {
			logger.Error("error on slavetask :", err)
		}
		if t.Status == "processing" {
			tries++
			time.Sleep(3 * time.Second)
			continue
		}

		return t.Solution, nil
	}
}

type HCaptchaProxyTask struct {
	ClientKey string `json:"clientKey"`
	Task      Task   `json:"task"`
}

type Task struct {
	Type       string `json:"type"`
	WebsiteURL string `json:"websiteURL"`
	WebsiteKey string `json:"websiteKey"`
	IsInvisble bool   `json:"isInvisible"`
	UserAgent  string `json:"userAgent"`
	// hcaptcha rqdata
	Data string `json:"data"`
}

//Captcha client
type CaptchaClient struct {
	UserAgent string
	TaskID    int64
	SiteKey   string
	TargetURL string
	Data      string
	client    *http.Client

	APIKey  string
	Service string
}
