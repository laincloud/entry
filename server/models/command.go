package models

import (
	"time"

	swaggermodels "github.com/laincloud/entry/server/gen/models"
)

// Command denotes the command typed by user
type Command struct {
	CommandID int64   `gorm:"primary_key"`
	Session   Session `gorm:"foreignkey:SessionID;association_foreignkey:SessionID"`
	SessionID int64
	User      string `gorm:"index"`
	Content   string
	CreatedAt time.Time `sql:"not null;DEFAULT:current_timestamp"`
}

// SwaggerModel return the swagger version
func (c Command) SwaggerModel() swaggermodels.Command {
	return swaggermodels.Command{
		CommandID:  c.CommandID,
		User:       c.User,
		AppName:    c.Session.AppName,
		ProcName:   c.Session.ProcName,
		InstanceNo: c.Session.InstanceNo,
		Content:    c.Content,
		SessionID:  c.SessionID,
		CreatedAt:  c.CreatedAt.Unix(),
	}
}
