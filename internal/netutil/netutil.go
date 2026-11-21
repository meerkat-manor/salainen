package netutil

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const requestTimeout = 30 * time.Second

func ValidateEndpoint(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL: %v", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("endpoint URL must use http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("endpoint URL must include a host")
	}
	if parsed.User != nil {
		return fmt.Errorf("endpoint URL must not include credentials")
	}
	if parsed.Scheme == "http" && !isLoopback(parsed.Hostname()) {
		return fmt.Errorf("endpoint URL must use https unless it targets localhost")
	}
	return nil
}

func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: requestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 && !sameOrigin(req.URL, via[0].URL) {
				return fmt.Errorf("cross-origin redirects are not allowed")
			}
			return nil
		},
	}
}

func isLoopback(host string) bool {
	host = strings.ToLower(strings.Trim(host, "[]"))
	if host == "localhost" || host == "localhost.localdomain" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}
