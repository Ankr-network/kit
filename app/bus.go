package app

import "sync"

type Event struct {
	Topic string
	Data  interface{}
}

type SubHandler func(e Event)

type SyncEventBus struct {
	subs  map[string][]SubHandler
	mutex sync.RWMutex
}

func NewSyncEventBus() *SyncEventBus {
	return &SyncEventBus{
		subs: map[string][]SubHandler{},
	}
}

func (s *SyncEventBus) Sub(topic string, handler SubHandler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	handlers, ok := s.subs[topic]
	if !ok {
		s.subs[topic] = []SubHandler{handler}
	} else {
		handlers = append(handlers, handler)
	}
}

func (s *SyncEventBus) Pub(topic string, data interface{}) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	e := Event{Topic: topic, Data: data}
	handlers, ok := s.subs[topic]
	if ok {
		for _, h := range handlers {
			h(e)
		}
	}
}
