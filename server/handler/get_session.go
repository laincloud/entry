package handler

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/jinzhu/gorm"

	"github.com/laincloud/entry/server/gen/restapi/operations/sessions"
	"github.com/laincloud/entry/server/models"
)

// GetSession get one session by session_id
func GetSession(params sessions.GetSessionParams, db *gorm.DB) middleware.Responder {
	var dbSession models.Session
	db.Where("session_id = ?", params.SessionID).First(&dbSession)
	swaggerSession := dbSession.SwaggerModel()
	return sessions.NewGetSessionOK().WithPayload(&swaggerSession)
}
