package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (s *ApiServer) router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", handler(root))
	r.HandleFunc("/login", handler(s.login))
	r.HandleFunc("/accounts", handler(s.handleAccounts))
	r.HandleFunc("/accounts/transfer/{id}", withJWTAuth(handler(s.transfer), s.store))
	r.HandleFunc("/accounts/{id}", withJWTAuth(handler(s.handleAccountById), s.store))

	return r
}

func root(w http.ResponseWriter, r *http.Request) error {
	return writeJson(w, http.StatusOK, "🚀 the server is up and running")
}
