package main

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	// Other signup fields
}

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthServer struct {
	store Storage
}

func newAuthServer(store Storage) *AuthServer {
	return &AuthServer{
		store: store,
	}
}

func (as *AuthServer) handleSignup(w http.ResponseWriter, r *http.Request) error {
	signupRequest := new(SignupRequest)
	if err := json.NewDecoder(r.Body).Decode(signupRequest); err != nil {
		return err
	}

	// Validate signup data
	if signupRequest.Username == "" || signupRequest.Password == "" {
		writeAPIError(w, http.StatusBadRequest, "Username and password are required")
	}

	// Check if the username already exists
	existingUser, err := as.store.getUserByUsername(signupRequest.Username)
	if err != nil {
		return err
	}
	if existingUser != nil {
		writeAPIError(w, http.StatusConflict, "Username already exists")
	}

	// Hash the password and create the user
	hashedPassword, err := hashPassword(signupRequest.Password)
	if err != nil {
		return err
	}
	user := &User{
		Username: signupRequest.Username,
		Password: hashedPassword,
	}
	err = as.store.createUser(user)
	if err != nil {
		return err
	}

	// Respond with success status
	return writeJSON(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func hashPassword(password string) (string, error) {
	// Generate a hash of the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// You can add more methods for authentication, login, and token generation here
