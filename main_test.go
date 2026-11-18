package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestValidateLocalVectors(t *testing.T) {
	accept := []struct {
		in   string
		want string
	}{
		{"http://127.0.0.1:20443", "http://127.0.0.1:20443"},
		{"http://localhost:20443", "http://127.0.0.1:20443"},
		{"http://LOCALHOST:20443/", "http://127.0.0.1:20443"},
		{"http://localhost", "http://127.0.0.1:20443"},
		{"http://[::1]:20443", "http://[::1]:20443"},
		{"http://[::1]", "http://[::1]:20443"},
		{"http://[::1]:30000/", "http://[::1]:30000"},
		{"http://localhost:20499/", "http://127.0.0.1:20499"},
	}
	for _, tc := range accept {
		got, err := ValidateLocal(tc.in)
		if err != nil || got != tc.want {
			t.Errorf("ValidateLocal(%q) = (%q, %v), want (%q, nil)", tc.in, got, err, tc.want)
		}
	}

	reject := []string{
		"",
		"https://127.0.0.1:20443",
		"http://0.0.0.0:20443",
		"http://192.168.1.5:20443",
		"http://localhost.evil.com:20443",
		"http://127.0.0.1.evil.com:20443",
		"http://2130706433:20443",
		"http://0x7f.0.0.1:20443",
		"http://127.1:20443",
		"http://[::ffff:127.0.0.1]:20443",
		"http://user:pw@127.0.0.1:20443",
		"http://127.0.0.1:20443/api",
		"http://127.0.0.1:20443/mcp",
		"http://127.0.0.1:20443/?next=http://evil.com",
		"http://127.0.0.1:20443?x=1",
		"http://127.0.0.1:20443/#frag",
		"http://127.0.0.1:20443#frag",
		"http://127.0.0.1:20443#",
		"http://127.0.0.1:80",
		"http://127.0.0.1:65536",
		"http://127.0.0.1:70000",
		"http://127.0.0.1:",
	}
	for _, bad := range reject {
		if got, err := ValidateLocal(bad); err == nil {
			t.Errorf("ValidateLocal(%q) expected error, got %q", bad, got)
		}
	}
}

func TestResolveDefaultAndLoadAPIToken(t *testing.T) {
	got, err := resolveDaemonURL("")
	if err != nil || got != "http://127.0.0.1:20443" {
		t.Fatalf("resolveDaemonURL empty = (%q, %v)", got, err)
	}

	tokFile := filepath.Join(t.TempDir(), "api-token")
	if err := os.WriteFile(tokFile, []byte("  file-secret-tok\n"), 0600); err != nil {
		t.Fatalf("write token file: %v", err)
	}
	if tok := LoadAPIToken("", tokFile); tok != "file-secret-tok" {
		t.Fatalf("LoadAPIToken file = %q, want file-secret-tok", tok)
	}
	if tok := LoadAPIToken("env-secret-tok", tokFile); tok != "env-secret-tok" {
		t.Fatalf("LoadAPIToken env = %q, want env-secret-tok", tok)
	}
}

func TestInitializeAndBearerAuthAndNoRedirect(t *testing.T) {
	const token = "standalone-secret"
	var redirectHits atomic.Int32

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	daemon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/health" {
			_, _ = w.Write([]byte(`{"version":"v1.1.0","status":"healthy"}`))
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer daemon.Close()

	var out bytes.Buffer
	dc, err := newDaemonClient(daemon.URL, token, &out)
	if err != nil {
		t.Fatalf("newDaemonClient: %v", err)
	}

	dc.handleRequest(jsonRPCRequest{JSONRPC: "2.0", ID: 1, Method: "initialize"})
	var initResp struct {
		Result struct {
			ServerInfo struct {
				Version string `json:"version"`
			} `json:"serverInfo"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out.Bytes(), &initResp); err != nil {
		t.Fatalf("unmarshal initialize: %v", err)
	}
	if initResp.Result.ServerInfo.Version != "v1.1.0" {
		t.Fatalf("serverInfo.version = %q, want v1.1.0", initResp.Result.ServerInfo.Version)
	}

	if res, err := dc.executeTool("chat.rooms.list", nil); err != nil || res != `{"ok":true}` {
		t.Fatalf("chat.rooms.list with token = (%q, %v)", res, err)
	}
	if res, err := dc.executeTool("chat.room.create", map[string]interface{}{"name": "r1"}); err != nil || res != `{"ok":true}` {
		t.Fatalf("chat.room.create with token = (%q, %v)", res, err)
	}

	dcNoAuth, _ := newDaemonClient(daemon.URL, "", &out)
	if _, err := dcNoAuth.executeTool("chat.rooms.list", nil); err == nil {
		t.Fatal("expected 401 error without token")
	}

	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer redir.Close()
	dcRedir, _ := newDaemonClient(redir.URL, token, &out)
	if _, err := dcRedir.executeTool("chat.rooms.list", nil); err == nil {
		t.Fatal("expected error on 302 redirect")
	}
	if redirectHits.Load() != 0 {
		t.Fatalf("redirect target received %d requests, want 0", redirectHits.Load())
	}
}
