package v2

import (
	"encoding/json"
	"github.com/laincloud/lainlet/api"
	"github.com/laincloud/lainlet/auth"
	"net/http"
)

func GetAppNameAPI(rw http.ResponseWriter, req *http.Request) (int, string) {
	if !auth.IsSuper(req.RemoteAddr) {
		return 400, "authorize failed, super required"
	}
	appname, err := auth.AppName(api.GetString(req, "ip", req.RemoteAddr))
	if err != nil {
		return 500, err.Error()
	}
	content, err := json.Marshal(map[string]string{"appname": appname})
	if err != nil {
		return 500, err.Error()
	}
	return 200, string(content)
}
