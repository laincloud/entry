package handler

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/laincloud/entry/server/gen/restapi/operations/ping"
)

// Ping return health status
func Ping(params ping.PingParams) middleware.Responder {
	return ping.NewPingOK().WithPayload("OK")
}
