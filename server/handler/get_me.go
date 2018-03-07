package handler

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/laincloud/entry/server/gen/models"
	"github.com/laincloud/entry/server/gen/restapi/operations/auth"
	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/util"
)

// GetMe return the user corresponding to the access_token
func GetMe(params auth.GetMeParams, g *global.Global) middleware.Responder {
	accessToken, err := params.HTTPRequest.Cookie("access_token")
	if err != nil {
		errMsg := err.Error()
		return auth.NewGetMeDefault(401).WithPayload(&models.Error{
			Message: &errMsg,
		})
	}

	user, err := util.AuthAPI(accessToken.Value, g)
	if err != nil {
		errMsg := err.Error()
		return auth.NewGetMeDefault(401).WithPayload(&models.Error{
			Message: &errMsg,
		})
	}

	swaggerUser := user.SwaggerModel()
	return auth.NewGetMeOK().WithPayload(&swaggerUser)
}
