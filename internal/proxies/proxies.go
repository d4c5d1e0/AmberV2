package proxies

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bytixo/AmberV2/internal/logger"
)

var (
	proxies []string
)

type Proxy struct {
	Scheme   string
	IP       string
	Port     int64
	Username string
	Password string
}

func init() {
	populateProxies()
}

func ParseProxy(proxy string) *Proxy {

	proxy = strings.Replace(proxy, "http://", "", 1)
	parsed := strings.Split(proxy, "@")

	u := strings.Split(parsed[0], ":")
	i := strings.Split(parsed[1], ":")
	port, _ := strconv.Atoi(i[1])

	return &Proxy{
		Scheme:   "http",
		IP:       i[0],
		Port:     int64(port),
		Username: u[0],
		Password: u[1],
	}

}
func GetProxy() string {
	for {
		proxy := proxies[rand.Intn(len(proxies))]

		err := CheckProxy(proxy)
		if err != nil {
			continue
		} else {
			return proxy
		}
	}
}
func populateProxies() {
	file, err := os.Open("data/proxies.txt")
	if err != nil {
		logger.Error(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxies = append(proxies, "http://"+scanner.Text())
	}

}

func CheckProxy(proxy string) error {
	purl, _ := url.Parse(proxy)
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy: http.ProxyURL(purl),
		},
	}

	res, err := client.Get("https://discord.com/api/v9/auth/location-metadata")
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		// handle error ...
		return err
	}

	if !strings.Contains(string(body), "consent_required") {
		return fmt.Errorf("bad content")
	}
	return nil
}
