package discord

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bytixo/AmberV2/internal/captcha"
	"github.com/bytixo/AmberV2/internal/config"
	"github.com/bytixo/AmberV2/internal/handler"
	"github.com/bytixo/AmberV2/internal/logger"
	"github.com/bytixo/AmberV2/internal/proxies"
	"github.com/tidwall/gjson"
)

const (
	XSuperProperties string = `eyJvcyI6Ik1hYyBPUyBYIiwiYnJvd3NlciI6IkNocm9tZSIsImRldmljZSI6IiIsInN5c3RlbV9sb2NhbGUiOiJmci1GUiIsImJyb3dzZXJfdXNlcl9hZ2VudCI6Ik1vemlsbGEvNS4wIChNYWNpbnRvc2g7IEludGVsIE1hYyBPUyBYIDEwXzEyXzYpIEFwcGxlV2ViS2l0LzUzNy4zNiAoS0hUTUwsIGxpa2UgR2Vja28pIENocm9tZS85OS4wLjQ4NDQuNzQgU2FmYXJpLzUzNy4zNiIsImJyb3dzZXJfdmVyc2lvbiI6Ijk5LjAuNDg0NC43NCIsIm9zX3ZlcnNpb24iOiIxMC4xMi42IiwicmVmZXJyZXIiOiJodHRwczovL3d3dy5nb29nbGUuY29tLyIsInJlZmVycmluZ19kb21haW4iOiJ3d3cuZ29vZ2xlLmNvbSIsInNlYXJjaF9lbmdpbmUiOiJnb29nbGUiLCJyZWZlcnJlcl9jdXJyZW50IjoiIiwicmVmZXJyaW5nX2RvbWFpbl9jdXJyZW50IjoiIiwicmVsZWFzZV9jaGFubmVsIjoic3RhYmxlIiwiY2xpZW50X2J1aWxkX251bWJlciI6MTE5NTk0LCJjbGllbnRfZXZlbnRfc291cmNlIjpudWxsfQ==`
	UserAgent        string = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.74 Safari/537.36`
)

var (
	//errors
	ErrCannotDM        = errors.New("can't message user")
	ErrNotAuthorized   = errors.New("invalid token")
	ErrRateLimited     = errors.New("rate limited")
	ErrTokenRateLimit  = errors.New("the token is rate limited")
	ErrTokenLocked     = errors.New("token is locked or invalid")
	ErrMemberScreening = errors.New("member screening bypass")
	ErrCaptcha         = errors.New("captcha detected")
)

func NewMessager(event *handler.AmberEvent) *Messager {
	return &Messager{
		Event: event,
	}
}

func (m *Messager) SendDM() error {
	err := m.initClient()
	if err != nil {
		return err
	}
	err = m.openChannel()
	if err != nil {
		return err
	}
	err = m.sendMessage()
	if err != nil {
		return err
	}

	return nil
}

func (m *Messager) openChannel() error {
	data := fmt.Sprintf(`{"recipients":["%s"]}`, m.Event.TargetID)

	req, err := http.NewRequest("POST", "https://discord.com/api/v9/users/@me/channels", strings.NewReader(data))
	if err != nil {
		return err
	}

	req.Header = http.Header{
		"X-Super-Properties":   []string{XSuperProperties},
		"X-Context-Properties": []string{`e30=`},
		"X-Debug-Options":      []string{`bugReporterEnabled`},
		"Accept-Language":      []string{`en-US,fr-FR;q=0.9`},
		"Authorization":        []string{m.Token},
		"Content-Type":         []string{`application/json`},
		"User-Agent":           []string{UserAgent},
		"X-Discord-Locale":     []string{`en`},
		"X-Fingerprint":        []string{m.Fingerprint},
		"Accept":               []string{`*/*`},
		"Origin":               []string{`https://discord.com`},
		"Sec-Fetch-Site":       []string{`same-origin`},
		"Sec-Fetch-Mode":       []string{`cors`},
		"Sec-Fetch-Dest":       []string{`empty`},
		"Referer":              []string{fmt.Sprintf(`https://discord.com/channels/%s/%s`, m.Event.GuildID, m.Event.ChannelID)},
	}

	res, err := m.client.Do(req)
	AddTotalRequests()
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	switch res.StatusCode {
	case 401:
		return ErrNotAuthorized
	case 403:
		return ErrTokenLocked
	case 429:
		return fmt.Errorf("rate limited")
	case 200:
		m.SnowFlake = gjson.Get(string(body), "id").String()
		return nil
	}
	return fmt.Errorf("error opening channel: does not contain id")
}
func (m *Messager) sendMessage() error {
	content := func() string {
		var mess string
		if strings.Contains(config.MessageToSend(), "<user>") {
			mess = strings.ReplaceAll(config.MessageToSend(), "<user>", fmt.Sprintf("<@%s>", m.Event.TargetID))
		} else {
			mess = config.MessageToSend()
		}
		return mess
	}()

	var data string
	if m.Task == nil {
		data = fmt.Sprintf(`{"content":"%s","nonce":"%d","tts":false}`, content, Snowflake())
	} else {
		data = fmt.Sprintf(`{"content":"%s","nonce":"%d","tts":false,"captcha_key":"%s","captcha_rqtoken":"%s"}`, content, Snowflake(), m.Task.Response, m.Task.Token)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://discord.com/api/v9/channels/%s/messages", m.SnowFlake), strings.NewReader(data))
	if err != nil {
		return err
	}

	req.Header = http.Header{
		"X-Super-Properties":   []string{XSuperProperties},
		"X-Context-Properties": []string{`e30=`},
		"X-Debug-Options":      []string{`bugReporterEnabled`},
		"Accept-Language":      []string{`en-US,fr-FR;q=0.9`},
		"Authorization":        []string{m.Token},
		"Content-Type":         []string{`application/json`},
		"User-Agent":           []string{UserAgent},
		"X-Discord-Locale":     []string{`en`},
		"X-Fingerprint":        []string{m.Fingerprint},
		"Accept":               []string{`*/*`},
		"Origin":               []string{`https://discord.com`},
		"Sec-Fetch-Site":       []string{`same-origin`},
		"Sec-Fetch-Mode":       []string{`cors`},
		"Sec-Fetch-Dest":       []string{`empty`},
		"Referer":              []string{fmt.Sprintf(`https://discord.com/channels/@me/%s`, m.SnowFlake)},
	}

	res, err := m.client.Do(req)
	AddTotalRequests()
	if err != nil {
		return err
	}

	err, m.Task = handleMessageError(res)
	if err != nil && errors.Is(err, ErrCaptcha) {
		logger.Info(fmt.Sprintf("Handling captcha (%s)", m.Token))
		err = m.handleCaptcha()
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func (m *Messager) handleCaptcha() error {
	client := captcha.NewClient(UserAgent, m.Task.Sitekey, "https://discord.com", config.CaptchaKey(), m.Task.Data)
	token, err := client.GetCaptcha()
	if err != nil {
		return err
	}
	m.Task.Response = token

	err = m.sendMessage()
	if err != nil {
		return err
	}
	return nil
}
func handleMessageError(res *http.Response) (error, *CaptchaTask) {
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err, nil
	}

	message := gjson.GetBytes(body, "message").String()
	code := gjson.GetBytes(body, "code").Int()

	logger.Debug(res.StatusCode, string(body))

	switch {
	case res.StatusCode == 200:
		return nil, nil
	case res.StatusCode == 401:
		return ErrNotAuthorized, nil
	case res.StatusCode == 400:
		task := new(CaptchaTask)
		task.Sitekey = gjson.GetBytes(body, "captcha_sitekey").String()
		task.Data = gjson.GetBytes(body, "captcha_rqdata").String()
		task.Token = gjson.GetBytes(body, "captcha_rqtoken").String()
		return ErrCaptcha, task
	case res.StatusCode == 403 && code == 50007:
		return ErrCannotDM, nil
	case res.StatusCode == 403 && code == 40003:
		return ErrTokenRateLimit, nil
	case res.StatusCode == 403 && code == 40002, res.StatusCode == 401, res.StatusCode == 405:
		return ErrTokenLocked, nil
	case res.StatusCode == 403 && code == 50009:
		return ErrMemberScreening, nil
	default:
		return fmt.Errorf("unknown response: %d, %s code %d", res.StatusCode, message, code), nil
	}
}
func (m *Messager) initClient() error {
	p := proxies.GetProxy()
	proxy, err := url.Parse(p)
	if err != nil {
		return fmt.Errorf("wrong proxy format %s should use, username:password@host:port", p)
	}

	//never return an error
	jar, _ := cookiejar.New(nil)

	//our transport
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MaxVersion: tls.VersionTLS13,
			CipherSuites: []uint16{
				0x1302, 0x1303, 0x1301, 0xC02C, 0xC030, 0xC02B, 0xC02F, 0xCCA9,
				0xCCA8, 0x009F, 0x009E, 0xCCAA, 0xC0AF, 0xC0AD, 0xC0AE, 0xC0AC,
				0xC024, 0xC028, 0xC023, 0xC027, 0xC00A, 0xC014, 0xC009, 0xC013,
				0xC0A3, 0xC09F, 0xC0A2, 0xC09E, 0x006B, 0x0067, 0x0039, 0x0033,
				0x009D, 0x009C, 0xC0A1, 0xC09D, 0xC0A0, 0xC09C, 0x003D, 0x003C,
				0x0035, 0x002F, 0x00FF,
			},
			InsecureSkipVerify: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveID(0x001D),
				tls.CurveID(0x0017),
				tls.CurveID(0x0018),
				tls.CurveID(0x0019),
				tls.CurveID(0x0100),
				tls.CurveID(0x0101),
			},
		},
		Proxy: http.ProxyURL(proxy),
	}
	client := &http.Client{
		Jar:       jar,
		Transport: tr,
		Timeout:   40 * time.Second,
	}

	m.client = client
	m.Token = GetRandomToken()
	m.Task = nil
	return nil
}

func (m *Messager) getFingerprint() error {
	req, err := http.NewRequest("POST", "https://discord.com/api/v9/auth/fingerprint", nil)
	if err != nil {
		return fmt.Errorf("error getting fingerprint: %s", err)
	}
	req.Header = http.Header{
		"Sec-Ch-Ua":          []string{`" Not A;Brand";v="99", "Chromium";v="99", "Google Chrome";v="99"`},
		"X-Super-Properties": []string{XSuperProperties},
		"X-Debug-Options":    []string{`bugReporterEnabled`},
		"Sec-Ch-Ua-Mobile":   []string{`?0`},
		"Authorization":      []string{`undefined`},
		"User-Agent":         []string{UserAgent},
		"X-Discord-Locale":   []string{`en`},
		"Sec-Ch-Ua-Platform": []string{`"macOS"`},
		"Accept":             []string{`*/*`},
		"Sec-Fetch-Site":     []string{`same-origin`},
		"Sec-Fetch-Mode":     []string{`cors`},
		"Sec-Fetch-Dest":     []string{`empty`},
		"Accept-Language":    []string{`en-US,en;q=0.9`},
	}
	res, err := m.client.Do(req)
	AddTotalRequests()
	if err != nil {
		return fmt.Errorf("error getting fingerprint: %s", err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error getting fingerprint: %s", err)
	}

	if !strings.Contains(string(body), "fingerprint") {
		return fmt.Errorf("error getting fingerprint: body not containing any fingerprint")
	}

	m.Fingerprint = gjson.Get(string(body), "fingerprint").String()
	return nil
}

//ty v4nshaj
func Snowflake() int64 {
	snowflake := strconv.FormatInt((time.Now().UTC().UnixNano()/1000000)-1420070400000, 2) + "0000000000000000000000"
	nonce, _ := strconv.ParseInt(snowflake, 2, 64)
	return nonce
}
