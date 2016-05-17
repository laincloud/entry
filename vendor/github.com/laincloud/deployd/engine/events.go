package engine

import "sync"

type Listener interface {
	ListenerId() string
	HandleEvent(payload interface{})
}

type Publisher interface {
	EmitEvent(payload interface{})
	AddListener(subscriber Listener)
	RemoveListener(subscriber Listener)
}

type _BasePublisher struct {
	sync.RWMutex
	goRoutine bool
	listeners map[string]Listener
}

func NewPublisher(goRoutine bool) Publisher {
	return &_BasePublisher{
		goRoutine: goRoutine,
		listeners: make(map[string]Listener),
	}
}

func (pub *_BasePublisher) EmitEvent(payload interface{}) {
	pub.RLock()
	listeners := make([]Listener, 0, len(pub.listeners))
	for _, listener := range pub.listeners {
		listeners = append(listeners, listener)
	}
	pub.RUnlock()

	emitFn := func() {
		for _, listener := range listeners {
			listener.HandleEvent(payload)
		}
	}
	if pub.goRoutine {
		go emitFn()
	} else {
		emitFn()
	}
}

func (pub *_BasePublisher) AddListener(listener Listener) {
	pub.Lock()
	defer pub.Unlock()
	pub.listeners[listener.ListenerId()] = listener
}

func (pub *_BasePublisher) RemoveListener(listener Listener) {
	pub.Lock()
	defer pub.Unlock()
	delete(pub.listeners, listener.ListenerId())
}
