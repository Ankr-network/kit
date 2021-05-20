package reporter

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)


// 用户注册
func TestUserSignUp(t *testing.T) {
	_ = os.Setenv("MIXPANEL_TOKEN", "947bca6172a824b18b3ffae793196f44")
	mixpanel := NewMixpanel()
	Init(Register(mixpanel), WithTimeout(time.Second*5), WithMaxGoroutine(1000))

	// 用户注册的属性
	properties := map[string]interface{}{
		"$email":       "fyx123@email.com",
		"$last_login":  time.Now(),
		"$created":     time.Now().String(),
	}

	if _, err := UpdateUser("fyxemm", properties); err != nil {  // async
		t.Fatal(err)
	}

	time.Sleep(time.Second * 5)
}


// 触发自定义事件
func TestSendEventExample1(t *testing.T) {
	_ = os.Setenv("SLACK_CHANNEL_ADDR", "https://hooks.slack.com/services/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")
	_ = os.Setenv("MIXPANEL_TOKEN", "947bca6172a824b18b3ffae793196f44")
	const channelName = "testchannel1"

	//slack := NewSlack(channelName)
	mixpanel := NewMixpanel()

	Init(Register(mixpanel), WithTimeout(time.Second*5), WithMaxGoroutine(1000))

	// logic here
	userID := "feixiang1209"
	eventName := "deploy app3"
	properties := map[string]interface{}{
		"deploy_way": "daily billing3",
	}

	if c, err := HookEvent(userID, eventName, properties); err == nil {
		if res, ok := <- c; ok { // 阻塞此处获得事件返回结果
			if res.err != nil {
				log.Println(res.msg, res.err)  // error handle
			}else {
				// 正常流程
			}
		}
	}


	time.Sleep(time.Second * 5)
}


func TestExample2(t *testing.T) {
	_ = os.Setenv("SLACK_CHANNEL_ADDR", "https://hooks.slack.com/services/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")
	_ = os.Setenv("MIXPANEL_TOKEN", "947bca6172a824b18b3ffae793196f44")
	const channelName = "testchannel1"

	slack := NewSlack(channelName)
	mixpanel := NewMixpanel()

	Init(Register(slack, mixpanel))  // 注册

	filters := &EventFilter{
		SendMixpanel:          true,                   //  发送到mixpanel
		SendSlackChannelNames: []string{channelName},  //  发送到这个slack的这个频道中, 可能有多个频道，选择需要发送的
	}
	_, err := HookEvent("feixiang1209", "deploy app3", map[string]interface{}{  // async call
		"deploy_way": "daily billing3",
	}, filters)

	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 10)
}

func TestBenchmark(t *testing.T) {
	_ = os.Setenv("SLACK_CHANNEL_ADDR", "https://hooks.slack.com/services/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")
	_ = os.Setenv("MIXPANEL_TOKEN", "947bca6172a824b18b3ffae793196f44")
	const channelName = "testchannel1"

	slack := NewSlack(channelName)
	mixpanel := NewMixpanel()

	Init(Register(slack, mixpanel), WithTimeout(time.Second*5), WithMaxGoroutine(1000))

	for {
		for i := 0; i < 50; i ++ {
			_, err := testEvent()
			if err != nil {
				log.Println(err)
			}
		}
		time.Sleep(time.Second * 3)
	}
}


func TestTimeoutEvent(t *testing.T) {
	_ = os.Setenv("SLACK_CHANNEL_ADDR", "https://hooks.slack.com/services/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")
	_ = os.Setenv("MIXPANEL_TOKEN", "947bca6172a824b18b3ffae793196f44")
	const channelName = "testchannel1"

	slack := NewSlack(channelName)
	mixpanel := NewMixpanel()

	Init(Register(slack, mixpanel), WithTimeout(time.Second*5), WithMaxGoroutine(1000))

	res, err := testEvent()
	if err != nil {
		panic(err)
	}

	if r, ok := <- res; ok {
		if r.err != nil {
			fmt.Println(r.err)
		}else {
			fmt.Println(r.msg)
		}
	}

	fmt.Println("ok")
	select {

	}
}

func testEvent() (chan *Result, error) {
	return hookEvent(func(ctx context.Context) (chan *Result, error) {
		if err := acquireLock(); err != nil {
			return nil, ErrTooManyCoroutines
		}
		defer freeLock()
		c := make(chan *Result, 1)
		r := &Result{}
		go func(ctx context.Context) {
			time.Sleep(time.Second * 5)
			r.msg = "example test success"
			c <- r
		}(ctx)
		return c, nil
	})
}

