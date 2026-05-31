package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUsersReturnsEmptyArray(t *testing.T) {
	// reset store before test
	defaultUserStore = newUserStore()

	req := httptest.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()

	GetUsers(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("want 200, got %d", res.Code)
	}

	var users []user
	err := json.NewDecoder(res.Body).Decode(&users)
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("expected empty array, got %d users", len(users))
	}
}

func TestGetUsersReturnsJSON(t *testing.T) {
	defaultUserStore = newUserStore()

	req := httptest.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()

	GetUsers(res, req)

	contentType := res.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("want 'application/json', got '%s'", contentType)
	}
}

func TestCreateUserReturns201(t *testing.T) {
	defaultUserStore = newUserStore()

	body := bytes.NewBufferString(`{"name":"David","email":"david@example.com"}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	if res.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", res.Code)
	}
}

func TestCreateUserValidatesName(t *testing.T) {
	defaultUserStore = newUserStore()

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
	defaultUserStore = newUserStore()

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

func TestCreateUserAddsToStore(t *testing.T) {
	defaultUserStore = newUserStore()

	body := bytes.NewBufferString(`{"name":"Eve","email":"eve@example.com"}`)
	req := httptest.NewRequest("POST", "/users", body)
	res := httptest.NewRecorder()

	CreateUser(res, req)

	var created user
	json.NewDecoder(res.Body).Decode(&created)

	// verify user exists in store
	allUsers := defaultUserStore.getAll()
	if len(allUsers) != 1 {
		t.Errorf("expected 1 user, got %d", len(allUsers))
	}
	if allUsers[0].ID != created.ID {
		t.Error("created user not found in storage")
	}
}

func TestGetUserByIDReturns404ForMissingUser(t *testing.T) {
	defaultUserStore = newUserStore()

	req := httptest.NewRequest("GET", "/users/999", nil)
	res := httptest.NewRecorder()

	GetUserByID(res, req)

	if res.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", res.Code)
	}
}

func TestGetUserByIDRequiresID(t *testing.T) {
	defaultUserStore = newUserStore()

	req := httptest.NewRequest("GET", "/users/", nil)
	res := httptest.NewRecorder()

	GetUserByID(res, req)

	if res.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", res.Code)
	}
}

func TestGetUserByIDAfterCreate(t *testing.T) {
	defaultUserStore = newUserStore()

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

func TestGetUserByIDTableDriven(t *testing.T) {
	defaultUserStore = newUserStore()

	// create a user first
	body := bytes.NewBufferString(`{"name":"Test","email":"test@example.com"}`)
	createReq := httptest.NewRequest("POST", "/users", body)
	createRes := httptest.NewRecorder()
	CreateUser(createRes, createReq)

	var created user
	json.NewDecoder(createRes.Body).Decode(&created)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantID     string
	}{
		{"valid user", "/users/" + created.ID, http.StatusOK, created.ID},
		{"missing user", "/users/999", http.StatusNotFound, ""},
		{"empty id", "/users/", http.StatusBadRequest, ""},
		{"no id", "/users", http.StatusBadRequest, ""},
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
