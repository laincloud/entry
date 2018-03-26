package handler

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/laincloud/entry/server/gen/restapi/operations/config"
	"github.com/laincloud/entry/server/global"
)

// GetConfig return the corresponding of entry
func GetConfig(params config.GetConfigParams, g *global.Global) middleware.Responder {
	swaggerConfig := g.Config.SwaggerModel()
	return config.NewGetConfigOK().WithPayload(&swaggerConfig)
}
