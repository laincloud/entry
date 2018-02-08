package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config denotes configuration
type Config struct {
	SSOURL string `json:"sso_url"`
	MySQL  MySQL  `json:"mysql"`
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
