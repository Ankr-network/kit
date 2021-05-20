package reporter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
	"io/ioutil"
	"net/http"
	"time"
)

var eventTemplate = `{
  "blocks": [
		{
			"type": "header",
			"text": {
				"type": "plain_text",
				"text": "EventHook Coming"
			}
		},
		{
			"type": "section",
			"fields": [
				{
					"type": "mrkdwn",
					"text": "*EventName:*\n%s"
				},
				{
					"type": "mrkdwn",
					"text": "*UserID:*\n%s"
				}
			]
		},
		{
			"type": "section",
			"fields": [
				{
					"type": "mrkdwn",
					"text": "*EventProperties:*\n%s"
				}
			]
		}
	]
}`


type (
	slackChannelName = string
	slackAddr = string
)

// 需要环境变量: SLACK_CHANNEL_ADDR  可以有多个
// SLACK_CHANNEL_ADDR与slack的channel名顺序必须对应
func NewSlack(slackChannelNames ...slackChannelName) *slack {
	return &slack{
		urls:             nil,
		ChannelAddrArray: nil,
		ChannelNameArray: slackChannelNames,
	}
}

type slack struct {
	urls map[slackChannelName]slackAddr

	ChannelAddrArray []string `env:"SLACK_CHANNEL_ADDR,required"`
	ChannelNameArray []string
}


func (s *slack) loadConfig() error {
	if err := env.Parse(s); err != nil {
		return errors.New("load slack config error")
	}

	lenAddr := len(s.ChannelAddrArray)
	lenName := len(s.ChannelNameArray)
	if lenAddr != lenName{
		panic("slack config error, the length of slack channel name is not match length of container env:SLACK_CHANNEL_ADDR")
	}

	if lenAddr == 0 || lenName == 0{
		panic("slack config error,make sure that env:SLACK_CHANNEL_ADDR is already set")
	}

	s.urls = make(map[string]string)
	for i := 0; i < lenAddr; i ++ {
		s.urls[s.ChannelNameArray[i]] = s.ChannelAddrArray[i]
	}

	return nil
}


func (s *slack) sendEvent(ctx context.Context, userID string, eventName string, properties map[string]interface{}, eventFilter ...*EventFilter) error {
	var propertyStr string
	for k, v := range properties {
		propertyStr += fmt.Sprintf("%s: %s\n", k, v)
	}

	msg := fmt.Sprintf(eventTemplate, eventName, userID, propertyStr)

	if len(eventFilter) > 0 {
		channelNames := eventFilter[0].SendSlackChannelNames
		if  channelNames == nil || len(channelNames) == 0 {
			return nil
		}

		for _, cn := range channelNames {
			if url, ok := s.urls[cn]; ok {
				_, _, err := postJSON(ctx, url, msg)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	for _, url := range s.urls {
		_, _, err := postJSON(ctx, url, msg)
		if err != nil {
			return err
		}
	}

	return nil
}


func postJSON(ctx context.Context, url string, body string) (int, string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return -1, "", err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, "", err
	}
	return resp.StatusCode, string(result), nil
}
