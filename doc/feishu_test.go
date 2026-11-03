package doc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/o98k-ok/aggregation/auth"
)

func TestSavedSessionSearch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, err := auth.SessionPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.SaveSession(path, "stored-session"); err != nil {
		t.Fatal(err)
	}
	for _, manual := range []string{"", "manual-session"} {
		t.Run("session="+manual, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				want := "stored-session"
				if manual != "" {
					want = manual
				}
				cookie, err := request.Cookie("session")
				if err != nil || cookie.Value != want {
					t.Error("wrong session used")
				}
				if request.URL.Query().Get("query") != "hello" {
					t.Error("search term lost")
				}
				fmt.Fprint(writer, `{"code":0,"data":{"entities":{"objs":{"doc":{"title":"Found","url":"https://example.com/doc"}}}}}`)
			}))
			defer server.Close()
			client := NewLark().CustomSession(manual)
			client.URL = server.URL
			entities, err := client.Query("hello")
			if err != nil || len(entities) != 1 || entities[0].Title != "Found" {
				t.Fatalf("search failed: %v", err)
			}
		})
	}
}

func TestSearchRequiresLogin(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if _, err := NewLark().Query("hello"); err == nil || !strings.Contains(err.Error(), "login") {
		t.Fatalf("missing login error: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, `{"code":5,"msg":"expired"}`)
	}))
	defer server.Close()
	client := NewLark().CustomSession("expired-session")
	client.URL = server.URL
	if _, err := client.Query("hello"); err == nil || !strings.Contains(err.Error(), "login") {
		t.Fatalf("missing re-login guidance: %v", err)
	}
}
