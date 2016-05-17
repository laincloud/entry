package adoc

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

func newHttpClient(u *url.URL, tlsConfig *tls.Config, timeout time.Duration, rwTimeout time.Duration) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	switch u.Scheme {
	case "unix":
		socketPath := u.Path
		transport.Dial = func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}
		u.Scheme = "http"
		u.Host = "unix.sock"
		u.Path = ""
	default:
		transport.Dial = func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, timeout)
		}
	}
	return &http.Client{
		Transport: transport,
		Timeout:   rwTimeout,
	}
}

func formatBoolToIntString(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
