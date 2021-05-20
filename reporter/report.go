package reporter

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	_defaultPoolNumber = 1000
	_defaultTimeout    = 5 * time.Second
)


var (
	reporter *Reporter
	once     sync.Once
)

type communicant interface {
	loadConfig() error
	sendEvent(ctx context.Context, userID string, eventName string, properties map[string]interface{}, eventFilter ...*EventFilter) error
}

// 没有传默认是所有注册的渠道都会通知, 传了需要每个字段都check
type EventFilter struct {
	SendMixpanel bool
	SendSlackChannelNames []slackChannelName
}

func Register(c ...communicant) []communicant {
	return c
}

type Reporter struct {
	err    error

	timeout time.Duration
	pool    chan struct{}

	lock sync.RWMutex
	communicants []communicant
}

type Opt func(t *Reporter)
type Opts []Opt

type Result struct {
	err error
	msg string

	userID     string
	eventName  string
	properties map[string]interface{}
}


var (
	ErrTooManyCoroutines = errors.New("too many threads are created")
)

func init() {
	reporter = &Reporter{
		err: errors.New("reporter is not initialized"),
	}
}

func WithTimeout(duration time.Duration) Opt {
	return func(t *Reporter) {
		t.timeout = duration
	}
}

func WithMaxGoroutine(cnt int64) Opt {
	return func(t *Reporter) {
		t.pool = make(chan struct{}, cnt)
		for i := 0; i < int(cnt); i++ {
			reporter.pool <- struct{}{}
		}
	}
}

func (o Opts) apply(r *Reporter) {
	for _, opt := range o {
		opt(r)
	}
}

func Init(c []communicant, opts ...Opt) {
	once.Do(func() {
		reporter = &Reporter{
			err:     nil,
			timeout: _defaultTimeout,
			pool:    make(chan struct{}, _defaultPoolNumber),
			communicants: c,
		}

		for i := 0; i < _defaultPoolNumber; i++ {
			reporter.pool <- struct{}{}
		}

		Opts(opts).apply(reporter)

		for _, v := range c {
			err := v.loadConfig()
			if err != nil {
				reporter.err = err
			}
		}
	})
}

func hookEvent(handler func(ctx context.Context) (chan *Result, error)) (chan *Result, error) {
	if reporter.err != nil {
		return nil, reporter.err
	}

	if err := acquireLock(); err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), reporter.timeout)
	res, err := handler(ctx)
	if err != nil {
		return nil, err
	}

	out := make(chan *Result, 1)
	go func() {
		defer func() {
			freeLock()
			close(out)
		}()
		for {
			select {
			case <-ctx.Done():
				out <- &Result{
					err: errors.New("report timeout"),
				}
				return
			case r := <-res:
				out <- r
				return
			}
		}
	}()
	return out, nil
}

func freeLock() {
	reporter.pool <- struct{}{}
}

func acquireLock() error {
	reporter.lock.Lock()
	defer reporter.lock.Unlock()

	if len(reporter.pool) == 0 {
		return ErrTooManyCoroutines
	}

	<-reporter.pool
	return nil
}

func HookEvent(userID string, eventName string, properties map[string]interface{}, eventFilter ...*EventFilter) (chan *Result, error) {
	return hookEvent(trace(userID, eventName, properties, eventFilter...))
}

func UpdateUser(userID string, properties map[string]interface{}) (chan *Result, error) {
	return hookEvent(updateUser(userID, properties))
}

func trace(userID string, eventName string, properties map[string]interface{}, eventFilter ...*EventFilter) func(ctx context.Context) (chan *Result, error) {
	return func(ctx context.Context) (chan *Result, error) {
		if err := acquireLock(); err != nil {
			return nil, ErrTooManyCoroutines
		}
		defer freeLock()

		c := make(chan *Result, 1)
		var r *Result
		go func(ctx context.Context) {
			r = &Result{
				err:        nil,
				msg:        "",
				userID:     userID,
				eventName:  eventName,
				properties: properties,
			}

			err := sendEvents(ctx, userID, eventName, properties, eventFilter...)

			if err != nil {
				r.err = err
				r.msg = err.Error()
			} else {
				r.msg = "success"
			}
			c <- r
		}(ctx)
		return c, nil
	}
}


func updateUser(userID string, properties map[string]interface{}) func(ctx context.Context) (chan *Result, error) {
	return func(ctx context.Context) (chan *Result, error) {
		if err := acquireLock(); err != nil {
			return nil, ErrTooManyCoroutines
		}
		defer freeLock()

		c := make(chan *Result, 1)
		var r *Result
		go func(ctx context.Context) {
			r = &Result{
				err:        nil,
				msg:        "",
				userID:     userID,
				properties: properties,
			}

			var err error
			for idx, c := range reporter.communicants {
				if mix, ok:= c.(*mixpanel); ok {
					err = mix.updateUser(ctx, userID, properties)
					break
				}

				if idx == len(reporter.communicants)-1 {
					err = errors.New("error: mixpannel may not be registered")
				}
			}

			if err != nil {
				r.err = err
				r.msg = err.Error()
			} else {
				r.msg = "success"
			}
			c <- r
		}(ctx)
		return c, nil
	}
}

func sendEvents(ctx context.Context, userID string, eventName string, properties map[string]interface{}, eventFilter ...*EventFilter) error {
	for _, c := range reporter.communicants {
		err := c.sendEvent(ctx, userID, eventName, properties, eventFilter...)
		if err != nil {
			return err
		}
	}
	return nil
}
