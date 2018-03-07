package sso

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	swaggermodels "github.com/laincloud/entry/server/gen/models"

	"github.com/laincloud/entry/server/config"
)

// AccessToken denotes the response of https://${sso.domain}/oauth2/token
type AccessToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// User denotes the response of https://${sso.domain}/api/me
type User struct {
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
}

// SwaggerModel return the swagger version
func (u User) SwaggerModel() swaggermodels.User {
	return swaggermodels.User{
		Email: &u.Email,
	}
}

// Client communicate with the SSO service
type Client struct {
	clientID     string
	clientSecret string
	domain       string
	entryGroup   string
	redirectURI  string
	httpClient   *http.Client
}

// NewClient return an initialized Client struct pointer
func NewClient(s config.SSO, httpClient *http.Client) *Client {
	return &Client{
		clientID:     s.ClientID,
		clientSecret: s.ClientSecret,
		domain:       s.Domain,
		entryGroup:   s.EntryGroup,
		redirectURI:  s.RedirectURI,
		httpClient:   httpClient,
	}
}

// GetUser get user from SSO according to accessToken
func (c Client) GetUser(accessToken string) (*User, error) {
	meURL := fmt.Sprintf("https://%s/api/me", c.domain)
	req, err := http.NewRequest("GET", meURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAccessToken get accessToken from SSO
func (c Client) GetAccessToken(authorizationCode string) (*AccessToken, error) {
	tokenURL := fmt.Sprintf("https://%s/oauth2/token", c.domain)
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", authorizationCode)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("redirect_uri", c.redirectURI)
	resp, err := c.httpClient.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var token AccessToken
	if err = json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// IsEntryOwner judge whether user is entry's owner
func (c Client) IsEntryOwner(user User) bool {
	for _, group := range user.Groups {
		if group == c.entryGroup {
			return true
		}
	}

	return false
}
