package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type loginTransport struct{ target *url.URL }

func (transport loginTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	request = request.Clone(request.Context())
	request.URL.Scheme = transport.target.Scheme
	request.URL.Host = transport.target.Host
	return http.DefaultTransport.RoundTrip(request)
}

func TestLogin(t *testing.T) {
	for _, scenario := range []string{"success-after-expiry", "cancelled", "missing-cookie", "api-error", "http-error"} {
		t.Run(scenario, func(t *testing.T) {
			initializations, polls, displayed := 0, 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodPost || request.Header.Get("x-app-id") != "2" {
					t.Error("invalid login request")
				}
				var body map[string]interface{}
				if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if value, exists := body["biz_type"]; !exists || value != nil {
					t.Error("missing null biz_type")
				}
				if strings.HasSuffix(request.URL.Path, "/init") {
					initializations++
					if body["redirect_uri"] != "https://www.feishu.cn" {
						t.Error("wrong redirect URI")
					}
					writer.Header().Set("x-flow-key", fmt.Sprintf("flow-%d", initializations))
					fmt.Fprintf(writer, `{"code":0,"data":{"step_info":{"token":"qr-%d"}}}`, initializations)
					return
				}
				polls++
				if request.Header.Get("x-flow-key") != fmt.Sprintf("flow-%d", initializations) {
					t.Error("stale flow key")
				}
				switch scenario {
				case "cancelled":
					fmt.Fprint(writer, `{"code":0,"data":{"step_info":{"status":3}}}`)
				case "api-error":
					fmt.Fprint(writer, `{"code":9,"message":"secret-must-not-leak"}`)
				case "http-error":
					writer.WriteHeader(http.StatusUnauthorized)
				case "missing-cookie":
					fmt.Fprint(writer, `{"code":0,"data":{"next_step":"enter_app"}}`)
				default:
					if polls == 1 {
						fmt.Fprint(writer, `{"code":0,"data":{"step_info":{"status":5}}}`)
						return
					}
					if polls == 2 {
						fmt.Fprint(writer, `{"code":0,"data":{"step_info":{"status":4},"next_step":"qr_login_polling"}}`)
						return
					}
					http.SetCookie(writer, &http.Cookie{Name: "other", Value: "ignore"})
					http.SetCookie(writer, &http.Cookie{Name: "session", Value: "test-session"})
					fmt.Fprint(writer, `{"code":0,"data":{"next_step":"enter_app"}}`)
				}
			}))
			defer server.Close()
			target, _ := url.Parse(server.URL)
			client := &http.Client{Transport: loginTransport{target}}
			session, err := Login(context.Background(), client, func(payload string) error {
				displayed++
				if payload != fmt.Sprintf(`{"qrlogin":{"token":"qr-%d"}}`, displayed) {
					t.Errorf("unexpected QR payload: %s", payload)
				}
				return nil
			})
			if scenario == "success-after-expiry" {
				if err != nil || session != "test-session" || displayed != 2 {
					t.Fatalf("login failed: displayed=%d err=%v", displayed, err)
				}
			} else if err == nil || session != "" || strings.Contains(err.Error(), "secret-must-not-leak") {
				t.Fatalf("expected safe failure, got %v", err)
			}
		})
	}
}

func TestLoginCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Login(ctx, http.DefaultClient, func(string) error { t.Fatal("unexpected QR"); return nil }); err == nil {
		t.Fatal("cancelled login succeeded")
	}
}
