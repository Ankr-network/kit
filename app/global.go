package app

var (
	appSyncEventBus = NewSyncEventBus()
)

func SubSync(topic string, h SubHandler) {
	appSyncEventBus.Sub(topic, h)
}

func PubSync(topic string, data interface{}) {
	appSyncEventBus.Pub(topic, data)
}
