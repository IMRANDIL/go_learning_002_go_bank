package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type transferRequest struct {
	FromAccountID int64   `json:"from_account_id"`
	ToAccountID   int64   `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func withJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header from the request
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeAPIError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Extract the JWT token from the header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			writeAPIError(w, http.StatusUnauthorized, "Invalid authorization header")
			return
		}

		secret := os.Getenv("JWT_SECRET")

		// Parse the JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// TODO: Replace with your actual JWT secret or public key
			return []byte(secret), nil
		})
		if err != nil {
			writeAPIError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Check if the token is valid
		if !token.Valid {
			writeAPIError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Extract the user ID from the claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeAPIError(w, http.StatusUnauthorized, "Invalid token claims")
			return
		}
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			writeAPIError(w, http.StatusUnauthorized, "Invalid user ID in token")
			return
		}
		userID := int(userIDFloat)

		// Add the user ID to the request context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		handlerFunc(w, r.WithContext(ctx))
	}
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func (s *APIServer) makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			//handle the error
			http.Error(w, err.Error(), http.StatusInternalServerError)
			// You can also log the error for debugging purposes
			// log.Printf("Error: %v", err)
		}
	}
}

func writeAPIError(w http.ResponseWriter, status int, message string) error {
	return writeJSON(w, status, &APIError{Status: status, Message: message})

}

type APIServer struct {
	listenAddr string
	router     *mux.Router
	storage    Storage
}

func newAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		router:     mux.NewRouter(),
		storage:    store,
	}
}

func (s *APIServer) setupRoutes() {
	//s.router.HandleFunc("/", s.makeHTTPHandleFunc(s.handleAccount)).Methods("GET")
	s.router.HandleFunc("/users/signup", s.makeHTTPHandleFunc(s.handleSignup)).Methods("POST")
	s.router.HandleFunc("/users/login", s.makeHTTPHandleFunc(s.handleLogin)).Methods("POST")
	s.router.HandleFunc("/accounts", s.makeHTTPHandleFunc(s.handleAllAccounts)).Methods("GET")
	s.router.HandleFunc("/user/details", withJWTAuth(s.makeHTTPHandleFunc(s.handleGetUserDetails))).Methods("GET")
	s.router.HandleFunc("/accounts", withJWTAuth(s.makeHTTPHandleFunc(s.handleCreateAccount))).Methods("POST")
	s.router.HandleFunc("/accounts/{id}", s.makeHTTPHandleFunc(s.handleDeleteAccount)).Methods("DELETE")
	s.router.HandleFunc("/accounts/{id}", s.makeHTTPHandleFunc(s.handleAccountById)).Methods("GET")
	s.router.HandleFunc("/accounts/{id}", s.makeHTTPHandleFunc(s.handleUpdateAccount)).Methods("PATCH")
	s.router.HandleFunc("/accounts/transfer", withJWTAuth(s.makeHTTPHandleFunc(s.handleAccountTransfer))).Methods("POST")
}

func (s *APIServer) run() {

	s.setupRoutes()
	fmt.Printf("Server listening on %s...\n", s.listenAddr)
	http.ListenAndServe(s.listenAddr, s.router)

}

// func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
// 	accountDetails := Account{
// 		ID:         1,
// 		FIRST_NAME: "ALI",
// 		LAST_NAME:  "IMRAN",
// 		HOBBY:      "CODING",
// 		AGE:        26,
// 		ACCOUNT:    212233222222,
// 		BALANCE:    "50000000000",
// 	}

// 	err := writeJSON(w, http.StatusOK, accountDetails)

// 	return err
// }

func (s *APIServer) handleAllAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.storage.allAccounts()

	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, accounts)
}

// func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
// 	account := newAccount("Imran", "Adil", "Cricket", 26)

// 	err := writeJSON(w, http.StatusOK, account)
// 	return err
// }

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(createAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return writeAPIError(w, http.StatusBadRequest, "Invalid request data")
	}

	// Get the user ID from the request context
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		return writeAPIError(w, http.StatusUnauthorized, "Invalid user ID in request context")
	}

	account := newAccount(createAccountReq.FIRST_NAME, createAccountReq.LAST_NAME, createAccountReq.HOBBY, createAccountReq.AGE, createAccountReq.BALANCE, userID)

	// Attempt to create the account
	if err := s.storage.createAccount(account); err != nil {
		// Check the specific error to determine the error response
		if strings.Contains(err.Error(), "user ID already exists") {
			return writeAPIError(w, http.StatusConflict, "An account with the same user ID already exists")
		}
		return writeAPIError(w, http.StatusInternalServerError, "Internal server error")
	}

	return writeJSON(w, http.StatusCreated, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	// Extract the account ID from the request URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "Invalid account ID")
		return nil
	}

	// Call the storage method to delete the account by ID
	err = s.storage.deleteAccount(id)
	if err != nil {
		// Check if the error is "account not found"
		if err.Error() == "account not found" {
			writeAPIError(w, http.StatusNotFound, "Account not found")
			return nil
		}

		// Handle other errors
		writeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return nil
	}

	// Respond with success status
	return writeJSON(w, http.StatusOK, map[string]string{"message": "Account deleted successfully"})
}

func (s *APIServer) handleAccountById(w http.ResponseWriter, r *http.Request) error {
	// Extract the account ID from the request URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "Invalid account ID")
		return nil
	}

	// Call the storage method to get the account by ID
	account, err := s.storage.getAccountById(id)
	if err != nil {
		if err.Error() == "account not found" {
			writeAPIError(w, http.StatusNotFound, "Account not found")
			return nil
		}
		writeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return nil
	}

	return writeJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "Invalid account ID")
		return nil
	}

	// Call the storage method to get the existing account by ID
	existingAccount, err := s.storage.getAccountById(id)
	if err != nil {
		if err.Error() == "account not found" {
			writeAPIError(w, http.StatusNotFound, "Account not found")
			return nil
		}
		writeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return nil
	}

	// Decode the updated account details from the request body
	updatedAccount := &Account{}
	if err := json.NewDecoder(r.Body).Decode(updatedAccount); err != nil {
		writeAPIError(w, http.StatusBadRequest, "Invalid request data")
		return nil
	}

	// Update the existing account with the new data if provided
	if updatedAccount.FIRST_NAME != "" {
		existingAccount.FIRST_NAME = updatedAccount.FIRST_NAME
	}
	if updatedAccount.LAST_NAME != "" {
		existingAccount.LAST_NAME = updatedAccount.LAST_NAME
	}
	if updatedAccount.HOBBY != "" {
		existingAccount.HOBBY = updatedAccount.HOBBY
	}
	if updatedAccount.AGE != 0 {
		existingAccount.AGE = updatedAccount.AGE
	}

	if updatedAccount.BALANCE != 0 {
		existingAccount.BALANCE = updatedAccount.BALANCE
	}

	// Call the storage method to update the account
	err = s.storage.updateAccount(existingAccount)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return nil
	}

	// Respond with success status
	return writeJSON(w, http.StatusOK, existingAccount)
}

func (s *APIServer) handleGetUserDetails(w http.ResponseWriter, r *http.Request) error {
	// Get the user ID from the request context (assuming you set it in the middleware)
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		return writeAPIError(w, http.StatusUnauthorized, "Invalid user ID in request context")
	}

	// Call the storage method to retrieve user details
	userDetails, err := s.storage.getUserDetails(userID)
	if err != nil {
		// Handle the error, for example, return a 404 Not Found response
		if err.Error() == "user not found" {
			return writeAPIError(w, http.StatusNotFound, "User not found")
		}
		return writeAPIError(w, http.StatusInternalServerError, "Internal server error")
	}

	// Return the user details in the response
	return writeJSON(w, http.StatusOK, userDetails)
}

func (s *APIServer) handleAccountTransfer(w http.ResponseWriter, r *http.Request) error {
	// Decode the transfer request details from the request body

	// Get the user ID from the request context
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		return writeAPIError(w, http.StatusUnauthorized, "Invalid user ID in request context")
	}
	fmt.Println(userID)

	// Fetch the user's account details using getUserDetails
	user, err := s.storage.getUserDetails(userID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error fetching user details")
		return nil
	}
	// Now you have the user's account information, including the "from" account
	fromAccount := user.Accounts[0] // Assuming the first account in the slice is the "from" account
	fromAccountNumber := fromAccount.ACCOUNT
	// Decode the transfer request details from the request body

	var transferReq transferRequest

	if err := json.NewDecoder(r.Body).Decode(&transferReq); err != nil {
		writeAPIError(w, http.StatusBadRequest, "Invalid request data")
		return nil
	}

	// Call the storage method to perform the balance transfer
	err = s.storage.transferBalance(fromAccountNumber, transferReq.ToAccountID, transferReq.Amount)
	if err != nil {
		// Handle different error scenarios
		if err.Error() == "insufficient balance in the account" {
			writeAPIError(w, http.StatusBadRequest, "Insufficient balance in the account")
			return nil
		}
		if err.Error() == "one or both accounts not found" {
			writeAPIError(w, http.StatusNotFound, "One or both accounts not found")
			return nil
		}
		writeAPIError(w, http.StatusInternalServerError, "Internal server error")
		return nil
	}

	// Respond with success status
	return writeJSON(w, http.StatusOK, map[string]string{"message": "Balance transferred successfully"})
}
