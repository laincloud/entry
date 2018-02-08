package global

import (
	"net/http"

	"github.com/fsouza/go-dockerclient"
	"github.com/jinzhu/gorm"
	lainlet "github.com/laincloud/lainlet/client"
)

// Global denotes global variables
type Global struct {
	DB            *gorm.DB
	DockerClient  *docker.Client
	HTTPClient    *http.Client
	LAINDomain    string
	LAINLETClient *lainlet.Client
	SSOURL        string
}
