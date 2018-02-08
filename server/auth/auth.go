package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/laincloud/entry/server/global"
)

const (
	anonymousEmail = "anonymous@anonymous.com"
)

var (
	ErrAuthFailed        = errors.New("authorize failed")
	ErrAuthNotSupported  = errors.New("entry only works on lain-sso authorization")
	ErrContainerNotfound = errors.New("get data successfully but not found the container")
)

// Auth authorizes whether the client with the token has the right to access the application
func Auth(token, appName string, g *global.Global) (*SSOUser, error) {
	var (
		data []byte
		err  error
	)
	if data, err = g.LAINLETClient.Get("/v2/configwatcher?target=auth/console", 2*time.Second); err != nil {
		return nil, err
	}
	authDataMap := make(map[string]string)
	if err = json.Unmarshal(data, &authDataMap); err != nil {
		return nil, err
	}
	if authStr, exist := authDataMap["auth/console"]; exist {
		c := ConsoleAuthConf{}
		if err = json.Unmarshal([]byte(authStr), &c); err != nil {
			return nil, err
		}
		if c.Type == "lain-sso" {
			authURL := fmt.Sprintf("http://console.%s/api/v1/repos/%s/roles/", g.LAINDomain, appName)
			return validateConsoleRole(authURL, token, g)
		}
		return nil, ErrAuthNotSupported
	}

	return &SSOUser{
		Email: anonymousEmail,
	}, nil
}
