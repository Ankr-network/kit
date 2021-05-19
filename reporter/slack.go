package reporter

import (
	"context"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v6"
)

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

	s.urls = make(map[string]string)
	for i := 0; i < lenAddr; i ++ {
		s.urls[s.ChannelNameArray[i]] = s.ChannelAddrArray[i]
	}

	return nil
}


func (s *slack) sendEvent(ctx context.Context, userID string, eventName string, properties map[string]interface{}) error {
	fmt.Println("in slack events")

	return nil
}

