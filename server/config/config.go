package config

import (
	"encoding/json"
	"fmt"
	"os"

	swaggermodels "github.com/laincloud/entry/server/gen/models"
)

// WriteBufferSize assign the websocket write buffer size
const WriteBufferSize = 10240

// Config denotes configuration
type Config struct {
	MySQL MySQL `json:"mysql"`
	SMTP  SMTP  `json:"smtp"`
	SSO   SSO   `json:"sso"`
}

// NewConfig return an initialized configuration
func NewConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config Config
	if err = json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SwaggerModel return the swagger version
func (c Config) SwaggerModel() swaggermodels.Config {
	return swaggermodels.Config{
		Sso: &swaggermodels.ConfigSso{
			Domain:      &c.SSO.Domain,
			ClientID:    &c.SSO.ClientID,
			RedirectURI: &c.SSO.RedirectURI,
			Scope:       &c.SSO.Scope,
		},
	}
}

// MySQL denotes MySQL configuration
type MySQL struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"db_name"`
}

// DataSourceName return the data source name to connect mysql
func (m MySQL) DataSourceName() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", m.Username, m.Password, m.Host, m.Port, m.DBName)
}

// SSO denotes SSO configuration
type SSO struct {
	Domain       string `json:"domain"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	EntryGroup   string `json:"entry_group"`
	Scope        string `json:"scope"`
}

// SMTP denotes SMTP configuration
type SMTP struct {
	Address   string `json:"address"`
	FromEmail string `json:"from_email"`
	Password  string `json:"password"`
}
