package util

import (
	"net"
	"net/smtp"

	"github.com/laincloud/entry/server/global"
)

// SendMail send mail
func SendMail(msg []byte, g *global.Global) error {
	s := g.Config.SMTP
	host, _, err := net.SplitHostPort(s.Address)
	if err != nil {
		return err
	}

	var auth smtp.Auth
	if s.Password != "" {
		auth = smtp.PlainAuth("", s.FromEmail, s.Password, host)
	}

	to, err := g.SSOClient.GetEntryOwnerEmails()
	if err != nil {
		return err
	}

	return smtp.SendMail(s.Address, auth, s.FromEmail, to, msg)
}
