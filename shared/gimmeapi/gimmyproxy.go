package gimmeapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

//https://gimmeproxy.com/api/getProxy?api_key=bf24b012-8eaf-4039-a5ab-377f02e0f6f2

func GetProxy(api string) (*Gimme, error) {
	for i := 0; i < 100; i++ {
		gp, err := getProxy(api)
		if err != nil {
			continue
		}
		return gp, nil
	}
	return nil, errors.New("fail load proxy-list")
}

func getProxy(api string) (*Gimme, error) {
	url := fmt.Sprintf("https://gimmeproxy.com/api/getProxy?api_key=%s", api)
	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if resp.StatusCode != 200 {
		time.Sleep(time.Second)
		return nil, fmt.Errorf("CODE:%d", resp.StatusCode)
	}
	var proxyResponse Gimme
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &proxyResponse)
	if err != nil {
		return nil, err
	}
	return &proxyResponse, nil
}

type Gimme struct {
	SupportsHttps  bool   `json:"supportsHttps"`
	Protocol       string `json:"protocol"`
	Ip             string `json:"ip"`
	Port           string `json:"port"`
	Get            bool   `json:"get"`
	Post           bool   `json:"post"`
	Cookies        bool   `json:"cookies"`
	Referer        bool   `json:"referer"`
	UserAgent      bool   `json:"user-agent"`
	AnonymityLevel int    `json:"anonymityLevel"`
	Websites       struct {
		Example    bool `json:"example"`
		Google     bool `json:"google"`
		Amazon     bool `json:"amazon"`
		Yelp       bool `json:"yelp"`
		GoogleMaps bool `json:"google_maps"`
	} `json:"websites"`
	Country         interface{} `json:"country"`
	UnixTimestampMs int64       `json:"unixTimestampMs"`
	TsChecked       int         `json:"tsChecked"`
	UnixTimestamp   int         `json:"unixTimestamp"`
	Curl            string      `json:"curl"`
	IpPort          string      `json:"ipPort"`
	Type            string      `json:"type"`
	Speed           float64     `json:"speed"`
	OtherProtocols  struct {
	} `json:"otherProtocols"`
	VerifiedSecondsAgo int `json:"verifiedSecondsAgo"`
}
