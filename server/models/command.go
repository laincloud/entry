package models

import (
	"time"

	swaggermodels "github.com/laincloud/entry/server/gen/models"
)

// Command denotes the command typed by user
type Command struct {
	CommandID int64 `gorm:"primary_key"`
	SessionID int64
	Session   Session
	User      string `gorm:"index"`
	Content   string
	CreatedAt time.Time `sql:"not null;DEFAULT:current_timestamp"`
}

// SwaggerModel return the swagger version
func (c Command) SwaggerModel() swaggermodels.Command {
	return swaggermodels.Command{
		CommandID: c.CommandID,
		SessionID: c.SessionID,
		Content:   c.Content,
		User:      c.User,
		CreatedAt: c.CreatedAt.Unix(),
	}
}
