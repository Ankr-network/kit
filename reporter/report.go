package reporter

import (
	"context"
	"errors"
	"github.com/dukex/mixpanel"
	"sync"
	"time"
)

const (
	_apiURL = "https://api.mixpanel.com"

	_defaultPoolNumber = 1000
	_defaultTimeout    = 5 * time.Second
)

type Reporter struct {
	err    error
	client mixpanel.Mixpanel

	timeout time.Duration
	pool    chan struct{}

	lock sync.RWMutex
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
	reporter *Reporter
	once     sync.Once
)

var ErrTooManyCoroutines = errors.New("too many threads are created")

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

func Init(cfg *Config, opts ...Opt) {
	once.Do(func() {
		client := mixpanel.New(cfg.MixpanelToken, _apiURL)
		reporter = &Reporter{
			err:     nil,
			client:  client,
			timeout: _defaultTimeout,
			pool:    make(chan struct{}, _defaultPoolNumber),
		}

		for i := 0; i < _defaultPoolNumber; i++ {
			reporter.pool <- struct{}{}
		}

		Opts(opts).apply(reporter)
	})
}

func hookEvent(handler func(ctx context.Context) (chan *Result, error)) (chan *Result, error) {
	if reporter.err != nil {
		return nil, reporter.err
	}

	if err := acquireLock(); err != nil { // 1
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
			freeLock() // 1
			close(out)
		}()
		for {
			select {
			case <-ctx.Done():
				//fmt.Println("timeout")
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

func HookEvent(userID string, eventName string, properties map[string]interface{}) (chan *Result, error) {
	return hookEvent(trace(userID, eventName, properties))
}

func UpdateUser(userID string, properties map[string]interface{}) (chan *Result, error) {
	return hookEvent(updateUser(userID, properties))
}

func trace(userID string, eventName string, properties map[string]interface{}) func(ctx context.Context) (chan *Result, error) {
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

			err := reporter.client.Track(userID, eventName, &mixpanel.Event{
				Properties: properties,
			})

			if err != nil {
				r.err = err
				r.msg = "timeout"
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

			err := reporter.client.Update(userID, &mixpanel.Update{
				Operation:  "$set",
				Properties: properties,
			})

			if err != nil {
				r.err = err
				r.msg = "timeout"
			} else {
				r.msg = "success"
			}
			c <- r
		}(ctx)
		return c, nil
	}
}
