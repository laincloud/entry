package handler

import (
	"time"

	"github.com/go-openapi/runtime/middleware"

	swaggermodels "github.com/laincloud/entry/server/gen/models"

	"github.com/laincloud/entry/server/gen/restapi/operations/sessions"
	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/models"
)

// ListSessions list sessions in database
func ListSessions(params sessions.ListSessionsParams, g *global.Global) middleware.Responder {
	var dbSessions []models.Session
	since := time.Unix(*params.Since, 0)
	newDB := g.DB.Where("created_at >= ?", since)
	if params.User != nil && *params.User != "" {
		newDB = newDB.Where("user = ?", *params.User)
	}
	if params.AppName != nil && *params.AppName != "" {
		newDB = newDB.Where("app_name = ?", *params.AppName)
	}
	newDB.Order("session_id desc").Limit(*params.Limit).Offset(*params.Offset).Find(&dbSessions)
	payload := make([]*swaggermodels.Session, len(dbSessions))
	for i, dbSession := range dbSessions {
		swaggerSession := dbSession.SwaggerModel()
		payload[i] = &swaggerSession
	}
	return sessions.NewListSessionsOK().WithPayload(payload)
}
