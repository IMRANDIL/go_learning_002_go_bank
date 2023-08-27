package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
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

func (s *APIServer) handleSignup(w http.ResponseWriter, r *http.Request) error {
	signupRequest := new(SignupRequest)
	if err := json.NewDecoder(r.Body).Decode(signupRequest); err != nil {
		return err
	}

	// Validate signup data
	if signupRequest.Username == "" || signupRequest.Password == "" {
		return writeAPIError(w, http.StatusBadRequest, "Username and password are required")
	}

	// Check if the username already exists
	existingUser, err := s.storage.getUserByUsername(signupRequest.Username)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return writeAPIError(w, http.StatusConflict, "Username already exists")

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
	err = s.storage.createUser(user)
	if err != nil {
		return err
	}

	// Respond with success status
	return writeJSON(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	loginRequest := new(SignupRequest)
	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		return err
	}

	// Validate login data
	if loginRequest.Username == "" || loginRequest.Password == "" {
		return writeAPIError(w, http.StatusBadRequest, "Username and password are required")
	}

	// Authenticate the user
	user, err := s.storage.authenticateUser(loginRequest.Username, loginRequest.Password)
	if err != nil {
		if err.Error() == "invalid password" {
			writeAPIError(w, http.StatusUnauthorized, "Invalid username or password")
			return nil
		}
		return err
	}
	if user == nil {
		return writeAPIError(w, http.StatusUnauthorized, "Invalid username or password")
	}

	// Generate JWT token
	tokenString, err := s.generateJWTToken(user.ID)
	if err != nil {
		return err
	}

	// Respond with JWT token
	return writeJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

func (s *APIServer) generateJWTToken(userID int) (string, error) {
	// TODO: Replace with your actual JWT secret or private key
	secret := []byte(os.Getenv("JWT_SECRET"))

	// Create the token claims
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate the token string
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
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
