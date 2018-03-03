package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/laincloud/entry/server/global"
)

// SSOUser denotes the response of sso.${LAIN-Domain}/api/me
type SSOUser struct {
	Email string `json:"email"`
}

// GetSSOUser get user from sso according to accessToken
func GetSSOUser(accessToken string, g *global.Global) (*SSOUser, error) {
	url := fmt.Sprintf("%s/api/me", g.SSOURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user SSOUser
	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
