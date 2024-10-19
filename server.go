package main

import (
	"log"
	"net/http"
)

type ApiServer struct {
	port  string
	store Storage
}

func (s *ApiServer) Run() {
	log.Printf("🚀 Server starting on port %s\n", s.port)
	err := http.ListenAndServe(s.port, s.router())
	log.Printf("🔥 Server failed: %s\n", err)
}
