package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
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
}

func newAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		router:     mux.NewRouter(),
	}
}

func (s *APIServer) setupRoutes() {
	s.router.HandleFunc("/", s.makeHTTPHandleFunc(s.handleAccount)).Methods("GET")
	s.router.HandleFunc("/accounts", s.makeHTTPHandleFunc(s.handleGetAccount)).Methods("GET")
	s.router.HandleFunc("/accounts", s.makeHTTPHandleFunc(s.handleCreateAccount)).Methods("POST")
	s.router.HandleFunc("/accounts/{id}", s.makeHTTPHandleFunc(s.handleDeleteAccount)).Methods("DELETE")
	s.router.HandleFunc("/accounts/transfer", s.makeHTTPHandleFunc(s.handleAccountTransfer)).Methods("POST")
}

func (s *APIServer) run() {

	s.setupRoutes()
	fmt.Printf("Server listening on %s...\n", s.listenAddr)
	http.ListenAndServe(s.listenAddr, s.router)

}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {

	_, err := fmt.Fprintf(w, "Hi ali")
	return err
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {

	return nil
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {

	return nil
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {

	return nil
}

func (s *APIServer) handleAccountTransfer(w http.ResponseWriter, r *http.Request) error {

	return nil
}
