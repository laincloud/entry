package handler

import (
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"

	"github.com/laincloud/entry/server/gen/restapi/operations/auth"
	"github.com/laincloud/entry/server/global"
)

// Logout delete access_token in cookie
func Logout(params auth.LogoutParams, g *global.Global) middleware.Responder {
	cookie := http.Cookie{
		Name:     keyAccessToken,
		Value:    "",
		Expires:  time.Now(),
		MaxAge:   0,
		Secure:   true,
		HttpOnly: true,
	}
	return auth.NewLogoutOK().WithSetCookie(cookie.String())
}
