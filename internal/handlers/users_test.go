package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUsersReturns200(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()

	GetUsers(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("want 200, got %d", res.Code)
	}
}

func TestGetUsersReturnsJSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()

	GetUsers(res, req)

	contentType := res.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("want 'application/json', got '%s'", contentType)
	}
}

func TestGetUsersReturnsArray(t *testing.T) {
	req := httptest.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()

	GetUsers(res, req)

	var users []user
	err := json.NewDecoder(res.Body).Decode(&users)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if len(users) == 0 {
		t.Error("expected at least one user")
	}
}

func TestCreateUserReturns201(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"David","email":"david@example.com"}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	if res.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", res.Code)
	}
}

func TestCreateUserValidatesName(t *testing.T) {
	body := bytes.NewBufferString(`{"email":"test@example.com"}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", res.Code)
	}

	var resp map[string]string
	json.NewDecoder(res.Body).Decode(&resp)
	if resp["error"] != "name is required" {
		t.Errorf("want 'name is required', got '%s'", resp["error"])
	}
}

func TestCreateUserValidatesEmail(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"Grace"}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", res.Code)
	}

	var resp map[string]string
	json.NewDecoder(res.Body).Decode(&resp)
	if resp["error"] != "email is required" {
		t.Errorf("want 'email is required', got '%s'", resp["error"])
	}
}

func TestGetUserByIDReturnsUser(t *testing.T) {
	originalStore := defaultUserStore
	defaultUserStore = newUserStore()
	defer func() { defaultUserStore = originalStore }()

	req := httptest.NewRequest("GET", "/users/1", nil)
	res := httptest.NewRecorder()

	GetUserByID(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("want 200, got %d", res.Code)
	}

	var u user
	err := json.NewDecoder(res.Body).Decode(&u)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if u.ID != "1" {
		t.Errorf("want id '1', got '%s'", u.ID)
	}
	if u.Name != "Alice" {
		t.Errorf("want name 'Alice', got '%s'", u.Name)
	}
	if u.Email != "alice@example.com" {
		t.Errorf("want email 'alice@example.com', got '%s'", u.Email)
	}
}

func TestGetUserByIDReturns404ForMissingUser(t *testing.T) {
	originalStore := defaultUserStore
	defaultUserStore = newUserStore()
	defer func() { defaultUserStore = originalStore }()

	req := httptest.NewRequest("GET", "/users/999", nil)
	res := httptest.NewRecorder()

	GetUserByID(res, req)

	if res.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", res.Code)
	}

	var resp map[string]string
	err := json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}
	if resp["error"] != "user not found" {
		t.Errorf("want 'user not found', got '%s'", resp["error"])
	}
}

func TestGetUserByIDReturns404ForDeletedUser(t *testing.T) {
	originalStore := defaultUserStore
	defaultUserStore = newUserStore()
	defer func() { defaultUserStore = originalStore }()

	req := httptest.NewRequest("GET", "/users/99999", nil)
	res := httptest.NewRecorder()

	GetUserByID(res, req)

	if res.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", res.Code)
	}
}

func TestGetUserByIDRequiresID(t *testing.T) {
	req := httptest.NewRequest("GET", "/users/", nil)
	res := httptest.NewRecorder()

	GetUserByID(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", res.Code)
	}
}

func TestGetUserByIDWithTrailingSlash(t *testing.T) {
	originalStore := defaultUserStore
	defaultUserStore = newUserStore()
	defer func() { defaultUserStore = originalStore }()

	req := httptest.NewRequest("GET", "/users/2/", nil)
	res := httptest.NewRecorder()

	GetUserByID(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("want 200, got %d", res.Code)
	}
}

func TestGetUserByIDReturnsJSON(t *testing.T) {
	originalStore := defaultUserStore
	defaultUserStore = newUserStore()
	defer func() { defaultUserStore = originalStore }()

	req := httptest.NewRequest("GET", "/users/1", nil)
	res := httptest.NewRecorder()

	GetUserByID(res, req)

	contentType := res.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("want 'application/json', got '%s'", contentType)
	}
}

func TestGetUserByIDAfterCreate(t *testing.T) {
	originalStore := defaultUserStore
	defaultUserStore = newUserStore()
	defer func() { defaultUserStore = originalStore }()

	// create a new user
	body := bytes.NewBufferString(`{"name":"NewUser","email":"new@example.com"}`)
	createReq := httptest.NewRequest("POST", "/users", body)
	createRes := httptest.NewRecorder()
	CreateUser(createRes, createReq)

	var created user
	json.NewDecoder(createRes.Body).Decode(&created)

	// fetch the created user
	getReq := httptest.NewRequest("GET", "/users/"+created.ID, nil)
	getRes := httptest.NewRecorder()
	GetUserByID(getRes, getReq)

	if getRes.Code != http.StatusOK {
		t.Errorf("want 200, got %d", getRes.Code)
	}

	var fetched user
	json.NewDecoder(getRes.Body).Decode(&fetched)

	if fetched.ID != created.ID {
		t.Errorf("want id '%s', got '%s'", created.ID, fetched.ID)
	}
	if fetched.Name != "NewUser" {
		t.Errorf("want name 'NewUser', got '%s'", fetched.Name)
	}
}

// table driven test for all get user by id cases
func TestGetUserByIDTableDriven(t *testing.T) {
	originalStore := defaultUserStore
	defaultUserStore = newUserStore()
	defer func() { defaultUserStore = originalStore }()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantID     string
		wantError  string
	}{
		{"valid user", "/users/1", http.StatusOK, "1", ""},
		{"valid user 2", "/users/2", http.StatusOK, "2", ""},
		{"valid user 3", "/users/3", http.StatusOK, "3", ""},
		{"missing user", "/users/999", http.StatusNotFound, "", "user not found"},
		{"empty id", "/users/", http.StatusBadRequest, "", ""},
		{"no id", "/users", http.StatusBadRequest, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			res := httptest.NewRecorder()

			GetUserByID(res, req)

			if res.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, res.Code)
			}

			if tt.wantStatus == http.StatusOK {
				var u user
				err := json.NewDecoder(res.Body).Decode(&u)
				if err != nil {
					t.Fatalf("failed to parse json: %v", err)
				}
				if u.ID != tt.wantID {
					t.Errorf("want id '%s', got '%s'", tt.wantID, u.ID)
				}
			}
		})
	}
}
