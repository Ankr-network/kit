package app

import (
	"fmt"
	"go.uber.org/atomic"
	"os"
	"os/signal"
	"syscall"
)

const (
	ExitTopic = "app.exit"
)

var (
	exited   = make(chan struct{}, 1)
	autoExit = atomic.NewBool(true)
)

func init() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go gracefulShutdown(sigCh)
}

func gracefulShutdown(sigCh <-chan os.Signal) {
	sig := <-sigCh
	Existing(fmt.Errorf("receive signal %v", sig))
	exited <- struct{}{}
	go doAutoExit()
}

func Existing(err error) {
	PubSync(ExitTopic, err)
}

func doAutoExit() {
	if autoExit.Load() {
		os.Exit(1)
	}
}

func Exit() {
	autoExit.Store(false)
	<-exited
	os.Exit(1)
}
