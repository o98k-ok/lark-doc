package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/o98k-ok/aggregation/auth"
)

type smokeTransport struct{}

func (smokeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	header := make(http.Header)
	body := `{"code":0,"data":{"next_step":"enter_app"}}`
	header.Add("Set-Cookie", "session=synthetic-session; Secure; HttpOnly")
	if strings.HasSuffix(request.URL.Path, "/init") {
		header.Set("x-flow-key", "synthetic-flow")
		body = `{"code":0,"data":{"step_info":{"token":"synthetic-qr-not-a-real-login"}}}`
	}
	return &http.Response{StatusCode: 200, Header: header, Body: io.NopCloser(strings.NewReader(body))}, nil
}

func TestLoginSavesSession(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	original := http.DefaultTransport
	http.DefaultTransport = smokeTransport{}
	defer func() { http.DefaultTransport = original }()
	if err := login(context.Background()); err != nil {
		t.Fatal(err)
	}
	path, err := auth.SessionPath()
	if err != nil {
		t.Fatal(err)
	}
	session, err := auth.LoadSession(path)
	if err != nil || session != "synthetic-session" {
		t.Fatal("CLI did not save login")
	}
}

func TestLoginKeepsOldSessionWhenCacheCannotBeRemoved(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	path, err := auth.SessionPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.SaveSession(path, "old-session"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { auth.DeleteSession(path) })
	cache, err := cachePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cache, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "block-removal"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	original := http.DefaultTransport
	http.DefaultTransport = smokeTransport{}
	defer func() { http.DefaultTransport = original }()
	if err := login(context.Background()); err == nil {
		t.Fatal("login succeeded despite cache cleanup failure")
	}
	session, err := auth.LoadSession(path)
	if err != nil || session != "old-session" {
		t.Fatal("account changed despite cache cleanup failure")
	}
}
