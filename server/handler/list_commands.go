package handler

import (
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/jinzhu/gorm"

	swaggermodels "github.com/laincloud/entry/server/gen/models"
	"github.com/laincloud/entry/server/gen/restapi/operations/commands"
	"github.com/laincloud/entry/server/models"
)

// ListCommands list commands in database
func ListCommands(params commands.ListCommandsParams, db *gorm.DB) middleware.Responder {
	var dbCommands []models.Command
	since := time.Unix(*params.Since, 0)
	db.Where("created_at > ? AND content LIKE ?", since, *params.Query).Limit(*params.Limit).Offset(*params.Offset).Find(&dbCommands)
	payload := make([]*swaggermodels.Command, len(dbCommands))
	for i, dbCommand := range dbCommands {
		swaggerCommand := dbCommand.SwaggerModel()
		payload[i] = &swaggerCommand
	}
	return commands.NewListCommandsOK().WithPayload(payload)
}
