package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mijia/sweb/log"

	swaggermodels "github.com/laincloud/entry/server/gen/models"
	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/util"
)

const (
	keyAccessToken = "access_token"
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
