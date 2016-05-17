package api

import (
	"net/http"
)

// API is a common interface which a api instance must realize
type API interface {
	// decode the data
	Decode([]byte) error

	// encode the data
	Encode() ([]byte, error)

	// the http route uri for this api
	URI() string

	// which watcher want to use
	WatcherName() string

	// the key used to watch,
	// eg. return '/lain/config/vip' for coreinfoWatcher,
	//     return 'console' as appname for coreifnoWacher and podgroupWatcher
	Key(r *http.Request) (string, error)

	// create new api by data
	// return a new instance of API with real data in it
	// second boolean value represent if the data really having some changes
	Make(data map[string]interface{}) (API, bool, error)
}

// BanWatcher is a interface to ban watch feature. you can realize this interface if your do not want your api support watch request.
type BanWatcher interface {
	BanWatch()
}
