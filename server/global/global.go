package global

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/jinzhu/gorm"
	lainlet "github.com/laincloud/lainlet/grpcclient"

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

// New return an initialized Global struct pointer
func New(c *config.Config, db *gorm.DB, dockerClient *docker.Client) (*Global, error) {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	lainletClient, err := lainlet.New(&lainlet.Config{
		Addr: net.JoinHostPort("lainlet.lain", os.Getenv("LAINLET_PORT")),
	})
	if err != nil {
		return nil, err
	}

	return &Global{
		Config:        c,
		DB:            db,
		DockerClient:  dockerClient,
		HTTPClient:    &httpClient,
		LAINDomain:    os.Getenv("LAIN_DOMAIN"),
		LAINLETClient: lainletClient,
		SSOClient:     sso.NewClient(c.SSO, &httpClient),
	}, nil
}
