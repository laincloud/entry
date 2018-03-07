package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/sso"
)

const (
	anonymousEmail = "anonymous@anonymous.com"
)

var (
	ErrAuthFailed        = errors.New("authorize failed")
	ErrAuthNotSupported  = errors.New("entry only works on lain-sso authorization")
	ErrContainerNotfound = errors.New("get data successfully but not found the container")
)

// AuthContainer authorizes whether the client with the token has the right to access the application's container
func AuthContainer(token, appName string, g *global.Global) (*sso.User, error) {
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

	return &sso.User{
		Email: anonymousEmail,
	}, nil
}

// AuthAPI authorizes whether the client with this token has right to access the API
func AuthAPI(accessToken string, g *global.Global) (*sso.User, error) {
	user, err := g.SSOClient.GetUser(accessToken)
	if err != nil {
		return nil, err
	}

	if !g.SSOClient.IsEntryOwner(*user) {
		return nil, fmt.Errorf("%s is not entry's owner", user.Email)
	}

	return user, nil
}
