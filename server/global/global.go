package global

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/jinzhu/gorm"
	lainlet "github.com/laincloud/lainlet/client"

	"github.com/laincloud/entry/server/config"
	"github.com/laincloud/entry/server/sso"
)

// Global denotes global variables
type Global struct {
	Config        *config.Config
	DB            *gorm.DB
	DockerClient  *docker.Client
	HTTPClient    *http.Client
	LAINDomain    string
	LAINLETClient *lainlet.Client
	SSOClient     *sso.Client
}

// NewGlobal return an initialized Global struct pointer
func NewGlobal(c *config.Config, db *gorm.DB, dockerClient *docker.Client) *Global {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	return &Global{
		Config:        c,
		DB:            db,
		DockerClient:  dockerClient,
		HTTPClient:    &httpClient,
		LAINDomain:    os.Getenv("LAIN_DOMAIN"),
		LAINLETClient: lainlet.New(net.JoinHostPort("lainlet.lain", os.Getenv("LAINLET_PORT"))),
		SSOClient:     sso.NewClient(c.SSO, &httpClient),
	}
}
