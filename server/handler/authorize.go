package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-openapi/runtime/middleware"

	"github.com/laincloud/entry/server/gen/models"
	"github.com/laincloud/entry/server/gen/restapi/operations/auth"
	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/util"
)

// Authorize is the callback api in SSO
func Authorize(params auth.AuthorizeParams, g *global.Global) middleware.Responder {
	token, err := g.SSOClient.GetAccessToken(params.Code)
	if err != nil {
		errMsg := err.Error()
		return auth.NewAuthorizeDefault(401).WithPayload(&models.Error{
			Message: &errMsg,
		})
	}

	user, err := util.AuthAPI(token.AccessToken, g)
	if err != nil {
		errMsg := err.Error()
		return auth.NewAuthorizeDefault(401).WithPayload(&models.Error{
			Message: &errMsg,
		})
	}

	vs := url.Values{}
	vs.Add("user", user.Email)
	location := fmt.Sprintf("/web?%s", vs.Encode())
	cookie := http.Cookie{
		Name:     keyAccessToken,
		Value:    token.AccessToken,
		Expires:  time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
		MaxAge:   token.ExpiresIn,
		Secure:   true,
		HttpOnly: true,
	}
	return auth.NewAuthorizeTemporaryRedirect().WithLocation(location).WithSetCookie(cookie.String())
}
