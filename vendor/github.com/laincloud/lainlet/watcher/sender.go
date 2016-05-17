package watcher

import (
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
	"github.com/laincloud/lainlet/store"
	"strings"
	"sync"
)

// TODO: give a Release() function to release all the receiver,

// Sender having a cache, when the data in cache changed, call Broadcast() sending new data to receivers.
type Sender struct {
	sync.Mutex
	*Cacher
	receivers []*Receiver
	// ctx context.Context // use a context to control release
}

// Receiver represents a receiver of sender, which can receive data from sender
type Receiver struct {
	key     string
	ctx     context.Context
	ch      chan *Event
	counter uint64
}

// NewSender create a sender and initialze it's cacher by the given data
func NewSender(data map[string]interface{}) *Sender {
	return &Sender{
		Cacher:    NewCacher(data),
		receivers: make([]*Receiver, 0, 10),
	}
}

// Broadcast the change event
// keys: the changed keys in cache
// action: the store action
func (s *Sender) Broadcast(keys []string, action store.Action) {
	s.Lock()
	defer s.Unlock()
	nilCounter := 0

	if len(s.receivers) > 0 {
		log.Debugf("Sender broadcast a new event, %s %v", action, keys)
	}
	for i, receiver := range s.receivers {
		if receiver == nil { // nil receiver, ignore it
			nilCounter++
			continue
		}
		select {
		case <-receiver.ctx.Done(): // receiver was canceled, set it nil
			log.Infof("Sender find a receiver watching %s was canceled, remove it", receiver.key)
			close(receiver.ch)
			s.receivers[i] = nil
			continue
		default:
		}
		// check is there any changed keys receiver is watching
		for _, key := range keys {
			if strings.HasPrefix(key, receiver.key) {
				select {
				case receiver.ch <- &Event{
					ID:     receiver.counter,
					Action: action,
					Data:   s.Get(receiver.key),
				}:
					log.Debugf("send success, increase counter")
					receiver.counter++
				default:
					log.Warnf("Sender fail to send event to the %s receiver, may closed or blocked", receiver.key)
				}
				break // in case of noticing multi times
			}
		}
	}
	// too many nil receiver, clear it
	if nilCounter > 0 && nilCounter >= len(s.receivers)/2 {
		go s.clearNilReceivers()
	}
}

// Watch a key in sender, this will add a new receiver in sender;
// the return channel will be closed when context was canceled.
func (s *Sender) Watch(key string, ctx context.Context) <-chan *Event {
	log.Infof("A new receiver watching to %s", key)
	r := &Receiver{
		key:     key,
		ctx:     ctx,
		ch:      make(chan *Event, 1),
		counter: 2,
	}
	s.receivers = append(s.receivers, r)
	return (<-chan *Event)(r.ch)
}

func (s *Sender) clearNilReceivers() {
	s.Lock()
	defer s.Unlock()
	log.Infof("Sender starting clear the nil receivers")
	newList := make([]*Receiver, 0, len(s.receivers))
	for _, receiver := range s.receivers {
		if receiver != nil {
			newList = append(newList, receiver)
		}
	}
	log.Debugf("Sender cleard %d nil receivers", len(s.receivers)-len(newList))
	s.receivers = newList
}
