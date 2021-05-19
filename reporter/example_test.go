package reporter

import (
	"fmt"
	"os"
	"testing"
)

func TestMy(t *testing.T) {
	_ = os.Setenv("SLACK_CHANNEL_ADDR", "https://hooks.slack.com/services/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")
	_ = os.Setenv("MIXPANEL_TOKEN", "947bca6172a824b18b3ffae793196f44")

	slack := NewSlack("testchannel1")
	mixpanel := NewMixpanel()

	Init(Register(slack, mixpanel))

	fmt.Println(reporter.err)

}
//
//// 用户注册
//func TestUserSignUp(t *testing.T) {
//	_ = os.Setenv("MIXPANEL_TOKEN", "947bca6172a824b18b3ffae793196f44")
//	cfg, _ := LoadConfig()
//	Init(cfg, WithTimeout(time.Second*5), WithMaxGoroutine(1000))
//
//	// 用户注册的属性
//	properties := map[string]interface{}{
//		"$email":       "fyx@email.com",
//		"$last_login":  time.Now(),
//		"$created":     time.Now().String(),
//		//"any key":    "any value",
//	}
//
//	if _, err := UpdateUser("fyxemm", properties); err != nil {  // async
//		t.Fatal(err)
//	}
//
//	time.Sleep(time.Second * 5)
//}
//
//
//// 触发自定义事件
//func TestSendEventExample1(t *testing.T) {
//	_ = os.Setenv("MIXPANEL_TOKEN", "947bca6172a824b18b3ffae793196f44")
//	cfg, _ := LoadConfig()
//	Init(cfg, WithTimeout(time.Second*5), WithMaxGoroutine(1000))
//
//	// logic here
//	userID := "feixiang1209"
//	eventName := "deploy app"
//	properties := map[string]interface{}{
//		"deploy_way": "daily billing2",
//	}
//
//	if _, err := HookEvent(userID, eventName, properties); err != nil { // async
//		t.Fatal(err)
//	}
//
//	time.Sleep(time.Second * 5)
//}
//
//func TestBenchmark(t *testing.T) {
//	_ = os.Setenv("MIXPANEL_TOKEN", "947bca6172a824b18b3ffae793196f44")
//	cfg, _ := LoadConfig()
//	Init(cfg, WithTimeout(time.Second*7), WithMaxGoroutine(20))
//
//	for {
//		for i := 0; i < 50; i ++ {
//			_, err := testEvent()
//			if err != nil {
//				log.Println(err)
//			}
//		}
//		time.Sleep(time.Second * 3)
//	}
//}
//
//
//func TestTimeoutEvent(t *testing.T) {
//	_ = os.Setenv("MIXPANEL_TOKEN", "947bca6172a824b18b3ffae793196f44")
//	cfg, _ := LoadConfig()
//	Init(cfg, WithTimeout(time.Second*6), WithMaxGoroutine(20))
//
//	res, err := testEvent()
//	if err != nil {
//		panic(err)
//	}
//
//	if r, ok := <- res; ok {
//		if r.err != nil {
//			fmt.Println(r.err)
//		}else {
//			fmt.Println(r.msg)
//		}
//	}
//
//	fmt.Println("ok")
//	select {
//
//	}
//}
//
//func testEvent() (chan *Result, error) {
//	return hookEvent(func(ctx context.Context) (chan *Result, error) {
//		if err := acquireLock(); err != nil {
//			return nil, ErrTooManyCoroutines
//		}
//		defer freeLock()
//		c := make(chan *Result, 1)
//		r := &Result{}
//		go func(ctx context.Context) {
//			time.Sleep(time.Second * 5)
//			//fmt.Println("example test")
//			r.msg = "example test success"
//			c <- r
//		}(ctx)
//		return c, nil
//	})
//}
//
