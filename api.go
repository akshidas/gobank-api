package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func writeJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

type apiError struct {
	Error string `json:"error"`
}

func handler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJson(w, http.StatusBadRequest, &apiError{Error: err.Error()})
		}
	}
}

type ApiServer struct {
	port  string
	store Storage
}

func (s *ApiServer) Run() {
	r := mux.NewRouter()
	r.HandleFunc("/", handler(root))
	r.HandleFunc("/accounts", handler(s.handleAccounts))
	r.HandleFunc("/accounts/{id}", handler(s.handleGetAccountById))

	log.Printf("🚀 Server starting on port %s\n", s.port)
	err := http.ListenAndServe(s.port, r)
	log.Printf("🔥 Server failed: %s\n", err)
}

func root(w http.ResponseWriter, r *http.Request) error {
	return writeJson(w, http.StatusOK, "🚀 the server is up and running")
}

func (s *ApiServer) handleAccounts(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccounts(w)
	}

	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}
	return writeJson(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *ApiServer) handleGetAccounts(w http.ResponseWriter) error {
	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err
	}

	return writeJson(w, http.StatusOK, accounts)
}
func (s *ApiServer) handleGetAccountById(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	id := vars["id"]
	parsedId, err := strconv.Atoi(id)

	if err != nil {
		return fmt.Errorf("invalid id given %s", id)
	}

	if r.Method == "GET" {

		account, err := s.store.GetAccountById(parsedId)

		if err != nil {
			return err
		}

		return writeJson(w, http.StatusOK, account)
	}

	if r.Method == "DELETE" {
		err := s.store.DeleteAccount(parsedId)
		if err != nil {
			return err
		}

		return writeJson(w, http.StatusOK, "deleted successfully")
	}

	return writeJson(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *ApiServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	newAccountPayload := &CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(newAccountPayload); err != nil {
		return err
	}

	account := NewAccount(newAccountPayload.FirstName, newAccountPayload.LastName)
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return writeJson(w, http.StatusCreated, account)
}
