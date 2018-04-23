package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/websocket"
	"github.com/mijia/sweb/log"

	"github.com/laincloud/entry/server/config"
	swaggermodels "github.com/laincloud/entry/server/gen/models"
	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/util"
)

const (
	keyAccessToken = "access_token"
	readBufferSize = 1024
)

var (
	noAuthAPIPaths = []string{
		"/enter",
		"/attach",
		"/api/authorize",
		"/api/config",
		"/api/logout",
		"/api/ping",
	}

	upgrader = websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

// AuthAPI judge whether the request has right to our API
func AuthAPI(h http.Handler, g *global.Global) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Receive a request: %+v.", r)
		defer func() {
			log.Infof("Request: %+v has been handled.", r)
		}()

		for _, p := range noAuthAPIPaths {
			if r.URL.Path == p {
				h.ServeHTTP(w, r)
				return
			}
		}

		accessToken, err := r.Cookie(keyAccessToken)
		if err != nil {
			errMsg := err.Error()
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(swaggermodels.Error{
				Message: &errMsg,
			})
			return
		}

		if _, err = util.AuthAPI(accessToken.Value, g); err != nil {
			errMsg := err.Error()
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(swaggermodels.Error{
				Message: &errMsg,
			})
			return
		}

		h.ServeHTTP(w, r)
	})
}

type websocketHandlerFunc func(ctx context.Context, conn *websocket.Conn, r *http.Request, g *global.Global)

// HandleWebsocket handle websocket request
func HandleWebsocket(ctx context.Context, f websocketHandlerFunc, r *http.Request, g *global.Global) middleware.Responder {
	return middleware.ResponderFunc(func(w http.ResponseWriter, _ runtime.Producer) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Errorf("Upgrade websocket protocol error: %s", err.Error())
			return
		}
		defer conn.Close()

		f(ctx, conn, r, g)
	})
}
