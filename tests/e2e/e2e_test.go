//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type registerResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID string `json:"id"`
	} `json:"user"`
}

type listResponse struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
}

type wishResponse struct {
	ID string `json:"id"`
}

type authTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func Test_E2E_FullFlow(t *testing.T) {
	baseURL := strings.TrimSuffix(os.Getenv("E2E_BASE_URL"), "/")
	if baseURL == "" {
		t.Skip("E2E_BASE_URL is not set")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	seed := time.Now().UnixNano()
	const password = "password123"

	user1Name := fmt.Sprintf("e2e_user_%d", seed)
	user2Name := fmt.Sprintf("e2e_user2_%d", seed)
	user1 := registerUser(t, client, baseURL, user1Name, fmt.Sprintf("e2e_user_%d@mail.test", seed), password)
	user2 := registerUser(t, client, baseURL, user2Name, fmt.Sprintf("e2e_user2_%d@mail.test", seed), password)

	// Login
	var loginResp registerResponse
	doJSON(t, client, http.MethodPost, baseURL+"/auth/login", "", map[string]any{"username": user1Name, "password": password}, &loginResp)
	if loginResp.AccessToken == "" || loginResp.RefreshToken == "" {
		t.Fatal("login response missing tokens")
	}

	// Refresh tokens
	var refreshed authTokensResponse
	doJSON(t, client, http.MethodPost, baseURL+"/auth/refresh", "", map[string]any{"refresh_token": user1.RefreshToken}, &refreshed)
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatal("refresh response missing tokens")
	}

	// Update current user
	updateUsername := "user" + user1.User.ID[4:]
	doJSON(t, client, http.MethodPatch, baseURL+"/users/me", user1.AccessToken, map[string]any{"username": updateUsername}, nil)

	// Get current user + by ID
	doNoBody(t, client, http.MethodGet, baseURL+"/users/me", user1.AccessToken)
	doNoBody(t, client, http.MethodGet, baseURL+"/users/"+user1.User.ID, user1.AccessToken)

	// Create list
	var list listResponse
	doJSON(t, client, http.MethodPost, baseURL+"/lists", user1.AccessToken, map[string]any{"title": "E2E list", "notes": "notes"}, &list)
	if list.ID == "" || list.Slug == "" {
		t.Fatal("list response missing id/slug")
	}

	// Get current user lists
	doNoBody(t, client, http.MethodGet, baseURL+"/lists", user1.AccessToken)

	// Get list by ID
	doNoBody(t, client, http.MethodGet, baseURL+"/lists/"+list.ID, user1.AccessToken)

	// Update list
	doJSON(t, client, http.MethodPatch, baseURL+"/lists/"+list.ID, user1.AccessToken, map[string]any{"title": "Updated list"}, nil)

	// Rotate share link
	var rotated struct {
		Slug string `json:"slug"`
	}
	doJSON(t, client, http.MethodPost, baseURL+"/lists/"+list.ID+"/rotate-share-link", user1.AccessToken, nil, &rotated)
	if rotated.Slug == "" {
		t.Fatal("rotate slug is empty")
	}
	list.Slug = rotated.Slug

	// Public lists by user ID
	doNoBody(t, client, http.MethodGet, baseURL+"/users/"+user1.User.ID+"/lists", user1.AccessToken)

	// Shared link (guest + authorized)
	doNoBody(t, client, http.MethodGet, baseURL+"/lists/shared/"+list.Slug, "")
	doNoBody(t, client, http.MethodGet, baseURL+"/lists/shared/"+list.Slug, user1.AccessToken)

	// Create wish
	var wish wishResponse
	doJSON(t, client, http.MethodPost, baseURL+"/lists/"+list.ID+"/wishes", user1.AccessToken, map[string]any{"title": "Steam Deck", "notes": "512GB", "price": 59900, "currency": "RUB"}, &wish)
	if wish.ID == "" {
		t.Fatal("wish id is empty")
	}

	// Update wish
	doJSON(t, client, http.MethodPatch, baseURL+"/lists/"+list.ID+"/wishes/"+wish.ID, user1.AccessToken, map[string]any{"notes": "OLED", "price": 69900}, nil)

	// Reserve + release by user2
	doNoBody(t, client, http.MethodPost, baseURL+"/lists/"+list.ID+"/wishes/"+wish.ID+"/reserve", user2.AccessToken)
	doNoBody(t, client, http.MethodDelete, baseURL+"/lists/"+list.ID+"/wishes/"+wish.ID+"/reserve", user2.AccessToken)

	// Delete wish
	doNoBody(t, client, http.MethodDelete, baseURL+"/lists/"+list.ID+"/wishes/"+wish.ID, user1.AccessToken)

	// Delete list
	doNoBody(t, client, http.MethodDelete, baseURL+"/lists/"+list.ID, user1.AccessToken)

	// Logout
	doJSON(t, client, http.MethodPost, baseURL+"/auth/logout", user1.AccessToken, map[string]any{"refresh_token": user1.RefreshToken}, nil)
	doJSON(t, client, http.MethodPost, baseURL+"/auth/logout", user2.AccessToken, map[string]any{"refresh_token": user2.RefreshToken}, nil)
}

func registerUser(t *testing.T, client *http.Client, baseURL, username, email, password string) registerResponse {
	registerBody := map[string]any{
		"name":     "E2E User",
		"username": username,
		"password": password,
	}
	if email != "" {
		registerBody["email"] = email
	}
	var resp registerResponse
	doJSON(t, client, http.MethodPost, baseURL+"/auth/register", "", registerBody, &resp)
	if resp.AccessToken == "" || resp.RefreshToken == "" || resp.User.ID == "" {
		t.Fatal("register response missing fields")
	}
	return resp
}

func doJSON(t *testing.T, client *http.Client, method, url, token string, reqBody any, out any) {
	t.Helper()

	var payload []byte
	if reqBody != nil {
		var err error
		payload, err = json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
	}

	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, url, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.Fatalf("request %s %s status=%d body=%s", method, url, resp.StatusCode, string(raw))
	}

	if out != nil {
		if err = json.Unmarshal(raw, out); err != nil {
			t.Fatalf("unmarshal response: %v (body=%s)", err, string(raw))
		}
	}
}

func doNoBody(t *testing.T, client *http.Client, method, url, token string) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, url, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.Fatalf("request %s %s status=%d body=%s", method, url, resp.StatusCode, string(raw))
	}
}
