package handler

import (
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/jinzhu/gorm"

	"github.com/laincloud/entry/server/gen/restapi/operations/sessions"
	"github.com/laincloud/entry/server/models"
	swaggermodels "github.com/laincloud/entry/server/gen/models"
)

// ListSessions list sessions in database
func ListSessions(params sessions.ListSessionsParams, db *gorm.DB) middleware.Responder {
	var dbSessions []models.Session
	since := time.Unix(*params.Since, 0)
	db.Where("created_at >= ?", since).Limit(*params.Limit).Offset(*params.Offset).Find(&dbSessions)
	payload := make([]*swaggermodels.Session, len(dbSessions))
	for i, dbSession := range dbSessions {
		swaggerSession := dbSession.SwaggerModel()
		payload[i] = &swaggerSession
	}
	return sessions.NewListSessionsOK().WithPayload(payload)
}
