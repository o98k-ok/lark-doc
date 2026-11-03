package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type loginResponse struct {
	Code *int `json:"code"`
	Data struct {
		NextStep string `json:"next_step"`
		StepInfo struct {
			Token  string `json:"token"`
			Status int    `json:"status"`
		} `json:"step_info"`
	} `json:"data"`
}

func loginRequest(ctx context.Context, client *http.Client, action, flowKey string) (loginResponse, http.Header, error) {
	body := `{"biz_type":null}`
	if action == "init" {
		body = `{"biz_type":null,"redirect_uri":"https://www.feishu.cn"}`
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://login.feishu.cn/accounts/qrlogin/"+action, strings.NewReader(body))
	if err != nil {
		return loginResponse{}, nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-version", "1.0.8")
	request.Header.Set("x-app-id", "2")
	request.Header.Set("x-device-info", "device_id=0;device_name=lark-doc;device_os=Mac")
	request.Header.Set("x-locale", "en-US")
	request.Header.Set("x-terminal-type", "2")
	if flowKey != "" {
		request.Header.Set("x-flow-key", flowKey)
	}
	response, err := client.Do(request)
	if err != nil {
		return loginResponse{}, nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return loginResponse{}, nil, fmt.Errorf("login HTTP status %d", response.StatusCode)
	}
	var result loginResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return result, nil, errors.New("invalid login response")
	}
	if result.Code == nil || *result.Code != 0 {
		return result, nil, errors.New("login API rejected request")
	}
	return result, response.Header, nil
}

// Login follows the Feishu web QR protocol used by deal/scripts/lark.
// display receives the QR payload, not a URL. It must not log or persist it.
func Login(ctx context.Context, client *http.Client, display func(string) error) (string, error) {
	for {
		init, headers, err := loginRequest(ctx, client, "init", "")
		if err != nil {
			return "", err
		}
		flowKey := headers.Get("x-flow-key")
		if init.Data.StepInfo.Token == "" || flowKey == "" {
			return "", errors.New("login response missing QR token or flow key")
		}
		payload, err := json.Marshal(map[string]interface{}{"qrlogin": map[string]string{"token": init.Data.StepInfo.Token}})
		if err != nil {
			return "", err
		}
		if err := display(string(payload)); err != nil {
			return "", err
		}
		for {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(500 * time.Millisecond):
			}
			poll, headers, err := loginRequest(ctx, client, "polling", flowKey)
			if err != nil {
				return "", err
			}
			if poll.Data.NextStep == "enter_app" {
				response := http.Response{Header: headers}
				for _, cookie := range response.Cookies() {
					if cookie.Name == "session" && cookie.Value != "" {
						return cookie.Value, nil
					}
				}
				return "", errors.New("login succeeded without a session cookie")
			}
			if poll.Data.StepInfo.Status == 3 {
				return "", errors.New("login cancelled on phone")
			}
			if poll.Data.StepInfo.Status == 5 {
				break
			}
		}
	}
}
