package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
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
	s.router.HandleFunc("/", s.makeHTTPHandleFunc(s.handleAccount)).Methods("GET")
	s.router.HandleFunc("/accounts", s.makeHTTPHandleFunc(s.handleAllAccounts)).Methods("GET")
	s.router.HandleFunc("/accounts", s.makeHTTPHandleFunc(s.handleCreateAccount)).Methods("POST")
	s.router.HandleFunc("/accounts/{id}", s.makeHTTPHandleFunc(s.handleDeleteAccount)).Methods("DELETE")
	s.router.HandleFunc("/accounts/{id}", s.makeHTTPHandleFunc(s.handleAccountById)).Methods("GET")
	s.router.HandleFunc("/accounts/transfer", s.makeHTTPHandleFunc(s.handleAccountTransfer)).Methods("POST")
}

func (s *APIServer) run() {

	s.setupRoutes()
	fmt.Printf("Server listening on %s...\n", s.listenAddr)
	http.ListenAndServe(s.listenAddr, s.router)

}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	accountDetails := Account{
		ID:         1,
		FIRST_NAME: "ALI",
		LAST_NAME:  "IMRAN",
		HOBBY:      "CODING",
		AGE:        26,
		ACCOUNT:    212233222222,
		BALANCE:    "50000000000",
	}

	err := writeJSON(w, http.StatusOK, accountDetails)

	return err
}

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
	//createAccountReq := createAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}

	account := newAccount(createAccountReq.FIRST_NAME, createAccountReq.LAST_NAME, createAccountReq.HOBBY, createAccountReq.AGE)

	if err := s.storage.createAccount(account); err != nil {
		return err
	}

	return writeJSON(w, http.StatusCreated, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {

	return nil
}

func (s *APIServer) handleAccountTransfer(w http.ResponseWriter, r *http.Request) error {

	return nil
}

func (s *APIServer) handleAccountById(w http.ResponseWriter, r *http.Request) error {
	// Extract the account ID from the request URL
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("invalid account ID")
	}

	// Call the storage method to get the account by ID
	account, err := s.storage.getAccountById(id)
	if err != nil {
		return err
	}

	// Write the account details as a JSON response
	return writeJSON(w, http.StatusOK, account)
}
