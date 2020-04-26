package mlog

import (
	"github.com/Ankr-network/kit/app"
	"io"
	"log"
	"sync"
)

var (
	std       = New()
	loggerMap sync.Map
)

func init() {
	app.SubSync(app.ExitTopic, func(_ app.Event) {
		Sync()
		log.Println("all loggers synced")
	})
}

func Logger(name string) *MLog {
	if name == "" {
		return std
	}

	logger, ok := loggerMap.Load(name)
	if !ok {
		logger = std.Clone(name)
		loggerMap.Store(name, logger)
	}

	return logger.(*MLog)
}

func Sync() {
	std.Sync()
	loggerMap.Range(func(key, value interface{}) bool {
		l := value.(*MLog)
		l.Sync()
		return true
	})
}

func Ignore(err error) {
	std.Ignore(err)
}

func Close(c io.Closer) {
	std.Close(c)
}
