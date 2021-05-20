package reporter

import (
	"context"
	"github.com/caarlos0/env/v6"
	mixpanelapi "github.com/dukex/mixpanel"
)

const (
	_apiURL = "https://api.mixpanel.com"
)

// 需要环境变量: MIXPANEL_TOKEN
func NewMixpanel() *mixpanel {
	return &mixpanel{}
}

type mixpanel struct {
	client mixpanelapi.Mixpanel
	Token  string `env:"MIXPANEL_TOKEN,required"`
}

func (s *mixpanel) loadConfig() error {
	if err := env.Parse(s); err != nil {
		panic("load mixpanel config error")
	}

	s.initClient(s.Token)
	return nil
}


func (s *mixpanel) sendEvent(ctx context.Context, userID string, eventName string, properties map[string]interface{}, eventFilter ...*EventFilter) error {
	if len(eventFilter) > 0 {
		if shouldSend := eventFilter[0].SendMixpanel; !shouldSend{
			return nil
		}
	}

	return s.client.Track(userID, eventName, &mixpanelapi.Event{
		Properties: properties,
	})
}

func (s *mixpanel) updateUser(ctx context.Context, userID string,  properties map[string]interface{}) error {
	return s.client.Update(userID, &mixpanelapi.Update{
		Operation:  "$set",
		Properties: properties,
	})
}


func (s *mixpanel) initClient(token string) {
	s.client = mixpanelapi.New(token, _apiURL)
}



