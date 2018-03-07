package handler

import (
	"time"

	"github.com/go-openapi/runtime/middleware"

	swaggermodels "github.com/laincloud/entry/server/gen/models"

	"github.com/laincloud/entry/server/gen/restapi/operations/commands"
	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/models"
)

// ListCommands list commands in database
func ListCommands(params commands.ListCommandsParams, g *global.Global) middleware.Responder {
	newDB := g.DB.Joins("inner join sessions on sessions.session_id = commands.session_id")
	since := time.Unix(*params.Since, 0)
	newDB = newDB.Where("commands.created_at > ?", since)
	if params.Content != nil && *params.Content != "" {
		newDB = newDB.Where("commands.content LIKE ?", *params.Content)
	}
	if params.AppName != nil && *params.AppName != "" {
		newDB = newDB.Where("sessions.app_name = ?", *params.AppName)
	}
	if params.User != nil && *params.User != "" {
		newDB = newDB.Where("sessions.user = ?", *params.User)
	}
	if params.SessionID != nil && *params.SessionID != 0 {
		newDB = newDB.Where("commands.session_id = ?", *params.SessionID)
	}
	var dbCommands []models.Command
	newDB.Order("commands.command_id desc").Limit(*params.Limit).Offset(*params.Offset).Preload("Session").Find(&dbCommands)
	payload := make([]*swaggermodels.Command, len(dbCommands))
	for i, dbCommand := range dbCommands {
		swaggerCommand := dbCommand.SwaggerModel()
		payload[i] = &swaggerCommand
	}
	return commands.NewListCommandsOK().WithPayload(payload)
}
