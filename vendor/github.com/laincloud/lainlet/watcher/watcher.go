package watcher

import (
	"fmt"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
	"github.com/laincloud/lainlet/store"
	"strings"
	"time"
)

const (
	// NIL represents a nil watcher, which will watch nothing
	NIL = "nil"
	// CONFIG represents a config watcher, watch /lain/config
	CONFIG = "config"
	// PODGROUP represents a podgroup watcher, watch /lain/deployd/pod_groups
	PODGROUP = "podgroup"
	// DEPENDS represents a depends watcher, watch /lain/deployd/depends
	DEPENDS = "depends"
	// NODES represents a node watcher, it's based on podgroup watcher
	NODES = "nodes"
	// CONTAINER represents a container watcher, it's based on podgroup watcher
	CONTAINER = "container"
)

// Event represents a watcher event
type Event struct {
	// ID is a event id, it is cumulative when you get a new event
	ID uint64
	// Action is the event type, init, update, delete or error
	Action store.Action
	// Data is event data, it always return the newest data, no matter what the action is
	Data map[string]interface{}
}

// Watcher having a sender, store, and convert function.
// It read and watch data from store, and use convert() to convert, then use sender to cache data and broadcast the event.
type Watcher struct {
	key       string
	convert   ConvertFunc
	status    Status
	ckey2skey func(string) string
	Store     store.Store
	Ctx       context.Context
	*Sender
}

// Status of a watcher
type Status struct {
	NumReceivers int
	UpdateTime   time.Time
	LastEvent    store.Event
	TotalKeys    int
}

// ConvertFunc convert the data from store into a general type
type ConvertFunc func([]*store.KVPair) (map[string]interface{}, error)

// New create a new watcher
func New(s store.Store, ctx context.Context, key string, convert ConvertFunc, ckey2skey func(string) string) (*Watcher, error) {
	watcher := &Watcher{
		key:       key,
		convert:   convert,
		ckey2skey: ckey2skey,
		Store:     s,
		Ctx:       ctx,
		Sender:    NewSender(nil),
	}
	go watcher.watchStore(key)
	return watcher, nil
}

func (w *Watcher) refresh() error {
	pairs, err := w.Store.GetTree(w.key)
	if err != nil {
		if err != store.ErrKeyNotFound {
			return err
		}
		log.Warnf("key %s do not exists on etcd", w.key)
		pairs = []*store.KVPair{}
	}
	data, err := w.convert(pairs)
	if err != nil {
		return err
	}
	w.Reset(data)
	return nil
}

// a general watch function
func (w *Watcher) watchStore(key string) {
	keys := make([]string, 0, 10)
	var (
		lastIndex             uint64
		broadcastAfterRefresh = false
	)
	for {
	START:
		if err := w.refresh(); err != nil {
			log.Errorf("Fail to refresh data for %s, %s", w.key, err.Error())
			time.Sleep(time.Second * 3)
			continue
		}
		if broadcastAfterRefresh {
			data := w.GetAll()
			keys := make([]string, 0, len(data))
			for k := range data {
				keys = append(keys, k)
			}
			w.Broadcast(keys, store.SET)
			broadcastAfterRefresh = false
		}
		log.Infof("A watcher starting to watch %s, from index %d", key, lastIndex)
		eventCh, err := w.Store.Watch(key, w.Ctx, true, lastIndex)
		if err != nil {
			log.Errorf("Fail to watch etcd, %s, retry watching after 3 seconds", err.Error())
			time.Sleep(time.Second * 3)
			continue
		}
		for {
			select {
			case <-w.Ctx.Done():
				log.Infof("Watcher for %s was canceled", key)
				return
			case event, ok := <-eventCh:
				if !ok {
					time.Sleep(time.Second * 3)
					broadcastAfterRefresh = true
					goto START
				}
				log.Debugf("Watcher get a store event, %s %s", event.Action, event.Key)

				// update watcher status
				w.status.LastEvent = *event
				w.status.UpdateTime = time.Now()

				switch event.Action {
				case store.SET, store.UPDATE:
					data, err := w.convert(event.Data)
					if err != nil {
						log.Errorf("Fail to convert event data, %s", err.Error())
						continue
					}
					for k, v := range data {
						w.Put(k, v)
						keys = append(keys, k)

					}
					lastIndex = event.ModifiedIndex
				case store.DELETE:
					for k := range w.GetAll() { // check all the cache data whether affected by this Key
						if strings.HasPrefix(w.ckey2skey(k), event.Key) {
							w.Delete(k, true)
							keys = append(keys, k)
						}
					}
					lastIndex = event.ModifiedIndex
				default:
					continue
				}
				log.Debugf("Watcher broadcast data by keys %v", keys)
				w.Broadcast(keys, event.Action)
				keys = keys[:0]
			}
		}
		log.Errorf("Etcd watche channel was closed by mistake, retry watching after 3 seconds")
		time.Sleep(time.Second * 3)
	}
}

// Watch function watch the keys having `prefix` prefix, it watch all the keys when prefix is '*'.
// it return a event channel. error wil be returned only when key is empty.
// the return channel have the newest data with init action in it.
func (w *Watcher) Watch(prefix string, ctx context.Context) (<-chan *Event, error) {
	switch prefix {
	case "":
		return nil, fmt.Errorf("empty key")
	case "*":
		return w.Sender.Watch("", ctx), nil
	default:
		return w.Sender.Watch(prefix, ctx), nil
	}
}

// Get function get the newest data for keys having `prefix` prefix. it returned error only when key is empty.
// it return all the data when prefix is '*'
func (w *Watcher) Get(prefix string) (map[string]interface{}, error) {
	switch prefix {
	case "":
		return nil, fmt.Errorf("empty key")
	case "*":
		return w.Sender.GetAll(), nil
	default:
		return w.Sender.Get(prefix), nil
	}
}

// Status return the watcher stats
func (w *Watcher) Status() Status {
	w.status.NumReceivers = len(w.receivers)
	w.status.TotalKeys = len(w.Sender.GetAll())
	return w.status
}
