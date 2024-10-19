package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *ApiServer) login(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("Method not allowed")
	}

	loginPayload := &LoginRequest{}

	if err := json.NewDecoder(r.Body).Decode(loginPayload); err != nil {
		return err
	}

	defer r.Body.Close()
	account, err := s.store.GetAccountByNumber(loginPayload.Number)
	if err != nil {
		return err
	}

	if !isPasswordValid(account.Password, loginPayload.Password) {
		permissionDenied(w)
		return nil
	}

	tokenStaring, err := createJwt(*account)
	if err != nil {
		return err
	}

	return writeJson(w, http.StatusCreated, tokenStaring)
}
