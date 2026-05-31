package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type user struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type createUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userStore struct {
	mu      sync.RWMutex
	users   map[string]user
	nextID  int
}

func newUserStore() *userStore {
	store := &userStore{
		users:   make(map[string]user),
		nextID:  4,
	}
	store.users["1"] = user{ID: "1", Name: "Alice", Email: "alice@example.com"}
	store.users["2"] = user{ID: "2", Name: "Bob", Email: "bob@example.com"}
	store.users["3"] = user{ID: "3", Name: "Charlie", Email: "charlie@example.com"}
	return store
}

func (s *userStore) getAll() []user {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]user, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users
}

func (s *userStore) create(name, email string) user {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := strconv.Itoa(s.nextID)
	s.nextID++

	u := user{
		ID:    id,
		Name:  name,
		Email: email,
	}
	s.users[u.ID] = u
	return u
}

func (s *userStore) getByID(id string) (user, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[id]
	return u, ok
}

var defaultUserStore = newUserStore()

func GetUsers(w http.ResponseWriter, r *http.Request) {
	users := defaultUserStore.getAll()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid json body"})
		return
	}

	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "name is required"})
		return
	}

	if req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "email is required"})
		return
	}

	newUser := defaultUserStore.create(req.Name, req.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

func GetUserByID(w http.ResponseWriter, r *http.Request) {
	// extract id from url path /users/{id}
	path := strings.TrimSuffix(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user id"})
		return
	}

	id := parts[len(parts)-1]
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user id is required"})
		return
	}

	// find user in storage
	user, found := defaultUserStore.getByID(id)
	if !found {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}

	// return user as json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
