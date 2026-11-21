package netutil

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestValidateEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "https", value: "https://vault.example.com"},
		{name: "localhost http", value: "http://127.0.0.1:8200"},
		{name: "ipv6 loopback http", value: "http://[::1]:8080"},
		{name: "remote http", value: "http://vault.example.com", wantErr: true},
		{name: "unsupported scheme", value: "ftp://vault.example.com", wantErr: true},
		{name: "missing host", value: "https://", wantErr: true},
		{name: "embedded credentials", value: "https://user:pass@vault.example.com", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateEndpoint(test.value)
			if test.wantErr && err == nil {
				t.Errorf("expected error for %q", test.value)
			}
			if !test.wantErr && err != nil {
				t.Errorf("unexpected error for %q: %v", test.value, err)
			}
		})
	}
}

func TestNewHTTPClientRedirectPolicy(t *testing.T) {
	client := NewHTTPClient()
	if client.Timeout != 30*time.Second {
		t.Errorf("unexpected timeout: %v", client.Timeout)
	}

	first, err := url.Parse("https://first.example/path")
	if err != nil {
		t.Fatal(err)
	}
	same, err := url.Parse("https://first.example/other")
	if err != nil {
		t.Fatal(err)
	}
	cross, err := url.Parse("https://second.example/path")
	if err != nil {
		t.Fatal(err)
	}

	via := []*http.Request{{URL: first}}
	if err := client.CheckRedirect(&http.Request{URL: same}, via); err != nil {
		t.Errorf("same-origin redirect rejected: %v", err)
	}
	if err := client.CheckRedirect(&http.Request{URL: cross}, via); err == nil {
		t.Errorf("cross-origin redirect was allowed")
	}
}
