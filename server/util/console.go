package util

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/sso"
)

type ConsoleAuthConf struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type ConsoleRole struct {
	Role string `json:"role"`
}

type ConsoleAuthResponse struct {
	Message string      `json:"msg"`
	URL     string      `json:"url"`
	Role    ConsoleRole `json:"role"`
}

func validateConsoleRole(authURL, token string, g *global.Global) (*sso.User, error) {
	var (
		err       error
		req       *http.Request
		resp      *http.Response
		respBytes []byte
	)
	if req, err = http.NewRequest("GET", authURL, nil); err != nil {
		return nil, err
	}
	req.Header.Set("access-token", token)
	if resp, err = g.HTTPClient.Do(req); err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if respBytes, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	caResp := ConsoleAuthResponse{}
	if err = json.Unmarshal(respBytes, &caResp); err != nil {
		return nil, err
	}
	if caResp.Role.Role == "" {
		return nil, ErrAuthFailed
	}
	return g.SSOClient.GetMe(token)
}
