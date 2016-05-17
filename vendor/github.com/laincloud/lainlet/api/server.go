package api

import (
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"
	"github.com/laincloud/lainlet/store"
	"github.com/laincloud/lainlet/watcher"
	"github.com/laincloud/lainlet/watcher/config"
	"github.com/laincloud/lainlet/watcher/container"
	"github.com/laincloud/lainlet/watcher/depends"
	"github.com/laincloud/lainlet/watcher/nodes"
	"github.com/laincloud/lainlet/watcher/podgroup"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var (
	debugConns int32
)

func getConnNum() int32 {
	return atomic.LoadInt32(&debugConns)
}

// Server used by lainlet, based on martini
type Server struct {
	*martini.Martini
	martini.Router
}

// New create a http api server; ip is the server ip, it was used by some query;
// version is the lainlet version, used to return by `/version` api;
// st is the backend store, it must be valided, should not be a empty interface or a nil value.
// This function return error when fail to initialize some backend watcher.
func New(ip, version string, st store.Store) (*Server, error) {
	r := martini.NewRouter()
	s := martini.New()

	background := context.Background()
	ctx := context.WithValue(background, "ip", ip)

	configWatcher, err := config.New(st, background)
	if err != nil {
		return nil, err
	}
	containerWatcher, err := container.New(st, background)
	if err != nil {
		return nil, err
	}
	podgroupWatcher, err := podgroup.New(st, background)
	if err != nil {
		return nil, err
	}
	dependsWatcher, err := depends.New(st, background)
	if err != nil {
		return nil, err
	}
	nodesWatcher, err := nodes.New(st, background)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, "configwatcher", configWatcher)
	ctx = context.WithValue(ctx, "dependswatcher", dependsWatcher)
	ctx = context.WithValue(ctx, "podgroupwatcher", podgroupWatcher)
	ctx = context.WithValue(ctx, "containerwatcher", containerWatcher)
	ctx = context.WithValue(ctx, "nodeswatcher", nodesWatcher)

	s.Use(martini.Recovery())
	s.Use(middleWareDebug)
	s.Use(middleWareWatchEvent)

	s.MapTo(r, (*martini.Router)(nil)) // router
	s.Map(log.Logger())                // logger
	s.Map(ctx)                         // context

	s.Action(r.Handle)

	// the debug api
	r.Get("/debug", func() (int, []byte) {
		data := map[string]interface{}{
			"goroutines":  runtime.NumGoroutine(),
			"connections": getConnNum(),
			"watchers": map[string]watcher.Status{
				"config":    configWatcher.Status(),
				"depends":   dependsWatcher.Status(),
				"container": containerWatcher.Status(),
				"podgroup":  podgroupWatcher.Status(),
				"nodes":     nodesWatcher.Status(),
			},
		}
		content, _ := json.Marshal(data)
		return 200, content
	})
	r.Get("/version", func() (int, []byte) {
		return 200, []byte(version)
	})

	return &Server{s, r}, nil
}

// Register a new api. apiserve will auto create a handler for it.
func (s *Server) Register(api API) {
	uri := api.URI()
	if len(uri) == 0 {
		panic("URI() return empty string")
	}
	if uri[0] != '/' {
		uri = "/" + uri
	}
	log.Infof("New api handler, uri=%s", uri)

	// raw http handler
	if handler, ok := api.(http.Handler); ok {
		s.Get("/v2"+uri, handler.ServeHTTP)
		return
	}
	s.Get(
		"/v2"+uri,
		func(w http.ResponseWriter, r *http.Request, es *EventSource, ctx context.Context) {
			if GetBool(r, "watch", false) {
				handleWatch(api, w, r, es, ctx)
			} else {
				handleGet(api, w, r, ctx)
			}
		},
	)
}

func handleWatch(api API, w http.ResponseWriter, r *http.Request, es *EventSource, ctx context.Context) {

	if _, ok := api.(BanWatcher); ok {
		es.SendEvent(0, store.ERROR.String(), api.URI()+" do not support watch action")
		return
	}

	key, err := api.Key(r)
	if err != nil {
		es.SendEvent(0, store.ERROR.String(), err.Error())
		return
	}

	log.Infof("Request want to watch the key %s", key)

	var (
		channel <-chan *watcher.Event
		wer     *watcher.Watcher
	)
	switch api.WatcherName() {
	case watcher.CONFIG:
		wer = ctx.Value("configwatcher").(*watcher.Watcher)
	case watcher.CONTAINER:
		wer = ctx.Value("containerwatcher").(*watcher.Watcher)
	case watcher.PODGROUP:
		wer = ctx.Value("podgroupwatcher").(*watcher.Watcher)
	case watcher.DEPENDS:
		wer = ctx.Value("dependswatcher").(*watcher.Watcher)
	case watcher.NODES:
		wer = ctx.Value("nodeswatcher").(*watcher.Watcher)
	default:
		es.SendEvent(0, store.ERROR.String(), "500 unkown watcher")
		return
	}

	var instance API
	// send the init data
	data, err := wer.Get(key)
	if err != nil {
		log.Errorf("Fail to get %s, %s", key, err.Error())
		es.SendEvent(0, store.ERROR.String(), err.Error())
		return
	}
	instance, _, err = api.Make(data)
	if err != nil {
		log.Errorf("Fail to make data for %s, %s", key, err.Error())
		es.SendEvent(0, store.ERROR.String(), err.Error())
		return
	}
	content, err := instance.Encode()
	if err != nil {
		log.Errorf("Fail to encode data for %s, %s", key, err.Error())
		es.SendEvent(0, store.ERROR.String(), err.Error())
		return
	}
	es.SendEvent(1, store.INIT.String(), content)

	// start watching
	channel, err = wer.Watch(key, ctx)
	if err != nil {
		log.Errorf("Fail to watch %s, %s", key, err.Error())
		es.SendEvent(0, store.ERROR.String(), err.Error())
		return
	}
	for {
		changed := false
		select {
		case event, ok := <-channel:
			if !ok {
				return
			}
			log.Infof("Get a %s event, id=%d action=%s ", api.WatcherName(), event.ID, event.Action.String())
			if event.Action == store.ERROR {
				es.SendEvent(event.ID, event.Action.String(), event.Data)
				continue
			}
			instance, changed, err = instance.Make(event.Data)
			if err != nil {
				es.SendEvent(0, store.ERROR.String(), err.Error())
				return
			}
			if !changed {
				continue
			}
			content, err := instance.Encode()
			if err != nil {
				es.SendEvent(0, store.ERROR.String(), err.Error())
				return
			}
			es.SendEvent(event.ID, event.Action.String(), content)
		case <-ctx.Done():
			log.Infof("Get stop signal, a connection stop watching to %s", key)
			return
		}
	}
	return
}

func handleGet(api API, w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var (
		wer *watcher.Watcher
	)

	key, err := api.Key(r)
	if err != nil {
		Return(w, 400, err.Error())
		return
	}

	switch api.WatcherName() {
	case watcher.CONFIG:
		wer = ctx.Value("configwatcher").(*watcher.Watcher)
	case watcher.CONTAINER:
		wer = ctx.Value("containerwatcher").(*watcher.Watcher)
	case watcher.PODGROUP:
		wer = ctx.Value("podgroupwatcher").(*watcher.Watcher)
	case watcher.DEPENDS:
		wer = ctx.Value("dependswatcher").(*watcher.Watcher)
	case watcher.NODES:
		wer = ctx.Value("nodeswatcher").(*watcher.Watcher)
	default:
		Return(w, 500, "Unkown watcher "+api.WatcherName())
		return
	}

	data, err := wer.Get(key)
	if err != nil {
		Return(w, 500, err.Error())
		return
	}
	instance, _, err := api.Make(data)
	if err != nil {
		Return(w, 500, err.Error())
		return
	}
	content, err := instance.Encode()
	if err != nil {
		Return(w, 500, err.Error())
		return
	}
	Return(w, 200, content)
}

func middleWareDebug(mctx martini.Context) {
	atomic.AddInt32(&debugConns, 1)
	mctx.Next()
	atomic.AddInt32(&debugConns, -1)
}

func middleWareWatchEvent(w http.ResponseWriter, r *http.Request, mctx martini.Context, ctx context.Context, router martini.Router) {
	log.Infof("New Request, %s %s %s", r.RemoteAddr, r.Method, r.URL)

	if len(router.MethodsFor(r.URL.Path)) == 0 {
		Return(w, 404, "404 not found")
		return
	}

	if !strings.HasPrefix(r.URL.Path, "/v2/") {
		return
	}

	if !GetBool(r, "watch", false) {
		log.Debugf("Request is not a watch action, create a empty eventsource in case of panic")
		mctx.Map(&EventSource{}) // add a useless event source
		return
	}

	/*********** Create a EventSource ***********/
	es, err := NewEventSource(w)
	if err != nil {
		log.Errorf("Fail to create a event source, %s", err.Error())
		w.WriteHeader(500)
		w.Write([]byte("Fail to create event source"))
		return
	}
	mctx.Map(es)

	/********* Create a heartbeat goroutine *********/
	if heartbeat := GetInt(r, "heartbeat", 0); heartbeat > 0 {
		tick := time.Duration(heartbeat) * time.Second
		log.Debugf("start a heartbeat goroutine, tick=%s", tick)
		go func(d time.Duration) {
			ticker := time.Tick(d)
			for {
				select {
				case <-ticker:
					es.SendEvent(0, "heartbeat", "")
				case <-es.CloseNotify():
					log.Debugf("heartbeat goroutine exit")
					return
				}
			}
		}(tick)
	}

	/******** Create a cancelable context ********/
	newCtx, cancel := context.WithCancel(ctx)
	mctx.Map(newCtx) // reset the context

	go func() {
		<-es.CloseNotify()
		cancel()
	}()

	mctx.Next() // Next() aim to call es.Close() after all the handler finished
	log.Infof("Request finished, close the eventsource")
	es.Close()
}

// Return write response to w, code is the http code, data is the content will write into responseBody, can be string or []byte.
func Return(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)
	switch data.(type) {
	case string:
		w.Write([]byte(data.(string)))
	case []byte:
		w.Write(data.([]byte))
	}
}

// GetInt search the integer argument by given name from the URL and PostForm, if not exists, return the given value as default.
func GetInt(r *http.Request, name string, value int) int {
	v := r.FormValue(name)
	if i, err := strconv.Atoi(v); err == nil {
		return i
	}
	return value
}

// GetString search the string argument by given name from the URL and PostForm, if not exists, return the given value as default.
func GetString(r *http.Request, name string, value string) string {
	if v := r.FormValue(name); v != "" {
		return v
		return value
	}
	return value
}

// GetBool search the boolean argument by given name from the URL and PostForm, if not exists, return the given value as default.
func GetBool(r *http.Request, name string, value bool) bool {
	switch r.FormValue(name) {
	case "", "0", "false":
		return false
	default:
		return true
	}
}
