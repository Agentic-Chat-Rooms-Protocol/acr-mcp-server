package main

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultLocalPort = 20443
	restTimeout      = 10 * time.Second
)

var errRedirect = errors.New("acr-mcp-server: redirects are not followed")

// ValidateLocal accepts only a loopback literal (or localhost pinned to 127.0.0.1)
// so a hosts-file or DNS answer cannot redirect daemon traffic.
func ValidateLocal(raw string) (string, error) {
	if raw == "" || strings.TrimSpace(raw) != raw || strings.Contains(raw, "#") {
		return "", errors.New("local: invalid URL")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.ForceQuery ||
		u.Fragment != "" || u.Opaque != "" || (u.Path != "" && u.Path != "/") ||
		u.RawPath != "" || strings.HasSuffix(u.Host, ":") {
		return "", errors.New("local: scheme, userinfo, path, query or fragment not allowed")
	}
	host := u.Hostname()
	switch {
	case strings.EqualFold(host, "localhost"):
		host = "127.0.0.1"
	case host == "127.0.0.1" || host == "::1":
	default:
		return "", errors.New("local: host must be localhost, 127.0.0.1 or ::1")
	}
	port := defaultLocalPort
	if p := u.Port(); p != "" {
		for i := 0; i < len(p); i++ {
			if p[i] < '0' || p[i] > '9' {
				return "", errors.New("local: port must be 1024-65535")
			}
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 1024 || n > 65535 {
			return "", errors.New("local: port must be 1024-65535")
		}
		port = n
	}
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port)), nil
}

func noRedirect(*http.Request, []*http.Request) error { return errRedirect }

// NewHTTPClient returns a bounded HTTP client that refuses HTTP redirects.
func NewHTTPClient() *http.Client {
	return &http.Client{Timeout: restTimeout, CheckRedirect: noRedirect}
}

// DefaultAPITokenPath returns ~/.acr/api-token.
func DefaultAPITokenPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".acr", "api-token"), nil
}

// LoadAPIToken reads ACR_API_TOKEN or falls back to tokenPath (~/.acr/api-token).
func LoadAPIToken(envToken, tokenPath string) string {
	if t := strings.TrimSpace(envToken); t != "" {
		return t
	}
	if tokenPath == "" {
		var err error
		tokenPath, err = DefaultAPITokenPath()
		if err != nil {
			return ""
		}
	}
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
