package sso

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/laincloud/entry/server/config"
)

const (
	roleAdmin = "admin"
)

// AccessToken denotes the response of https://${sso.domain}/oauth2/token
type AccessToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// Group denotes the response of https://${sso.domain}/api/groups/{groupname}
type Group struct {
	Members      []Member      `json:"members"`
	GroupMembers []GroupMember `json:"group_members"`
}

// Member denotes the member of group
type Member struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

// GroupMember denotes the subgroup of group
type GroupMember struct {
	Name     string `json:"name"`
	Fullname string `json:"fullname"`
	Role     string `json:"role"`
}

// User denotes sso user
type User struct {
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
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

// GetMe get user from SSO according to accessToken
func (c Client) GetMe(accessToken string) (*User, error) {
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

// GetEntryOwnerEmails get emails of the owners of entry
func (c Client) GetEntryOwnerEmails() ([]string, error) {
	return c.getGroupEmails(c.entryGroup)
}

func (c Client) getGroupEmails(groupName string) ([]string, error) {
	group, err := c.GetGroup(groupName)
	if err != nil {
		return nil, err
	}

	emails := make([]string, 0)
	for _, member := range group.Members {
		if member.Role == roleAdmin {
			user, err := c.GetUser(member.Name)
			if err != nil {
				return nil, err
			}

			emails = append(emails, user.Email)
		}
	}

	for _, subGroup := range group.GroupMembers {
		if subGroup.Role == roleAdmin {
			groupEmails, err := c.getGroupEmails(subGroup.Name)
			if err != nil {
				return nil, err
			}

			emails = append(emails, groupEmails...)
		}
	}

	return emails, nil
}

// GetGroup get group info from sso
func (c Client) GetGroup(groupName string) (*Group, error) {
	groupURL := fmt.Sprintf("https://%s/api/groups/%s", c.domain, groupName)
	resp, err := c.httpClient.Get(groupURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var group Group
	if err = json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, err
	}

	return &group, nil
}

// GetUser get user info from sso
func (c Client) GetUser(userName string) (*User, error) {
	userURL := fmt.Sprintf("https://%s/api/users/%s", c.domain, userName)
	resp, err := c.httpClient.Get(userURL)
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
