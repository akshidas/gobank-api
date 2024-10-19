package main

import (
	"encoding/json"
	"net/http"
)

func (s *ApiServer) handleAccounts(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.getAccounts(w)
	}

	if r.Method == "POST" {
		return s.createAccounts(w, r)
	}
	return writeJson(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *ApiServer) handleAccountById(w http.ResponseWriter, r *http.Request) error {

	if r.Method == "GET" {
		return s.getAccountById(w, r)
	}

	if r.Method == "DELETE" {
		return s.deleteAccout(w, r)
	}

	return writeJson(w, http.StatusMethodNotAllowed, "method not allowed")
}

// Handler Methods
func (s *ApiServer) getAccounts(w http.ResponseWriter) error {
	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err
	}

	return writeJson(w, http.StatusOK, accounts)
}

func (s *ApiServer) createAccounts(w http.ResponseWriter, r *http.Request) error {
	newAccountPayload := &CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(newAccountPayload); err != nil {
		return err
	}

	defer r.Body.Close()
	hashedPassword, err := hashPassword(newAccountPayload.Password)
	if err != nil {
		return err
	}

	account := NewAccount(newAccountPayload.FirstName, newAccountPayload.LastName, hashedPassword)
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	tokenStaring, err := createJwt(*account)
	if err != nil {
		return err
	}

	return writeJson(w, http.StatusCreated, tokenStaring)
}

func (s *ApiServer) getAccountById(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}

	account, err := s.store.GetAccountById(id)
	if err != nil {
		return err
	}

	return writeJson(w, http.StatusOK, account)
}

func (s *ApiServer) deleteAccout(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)

	if err != nil {
		return err
	}

	err = s.store.DeleteAccount(id)
	if err != nil {
		return err
	}

	return writeJson(w, http.StatusOK, "deleted successfully")

}

func (s *ApiServer) transfer(w http.ResponseWriter, r *http.Request) error {
	newAccountPayload := &TransferAmountRequest{}
	if err := json.NewDecoder(r.Body).Decode(newAccountPayload); err != nil {
		return err
	}

	defer r.Body.Close()

	id, err := getId(r)
	if err != nil {
		return err
	}

	fromAccount, err := s.store.GetAccountById(id)
	if err != nil {
		return err
	}

	if fromAccount.Balance < int64(newAccountPayload.Amount) {
		return writeJson(w, http.StatusBadRequest, "insufficient balance")
	}

	toAccount, err := s.store.GetAccountById(newAccountPayload.ToAccount)
	if err != nil {
		return err
	}

	if toAccount.Id == fromAccount.Id {
		return writeJson(w, http.StatusBadRequest, "cannot do self transfer")

	}

	toAccount.Balance = toAccount.Balance + int64(newAccountPayload.Amount)
	fromAccount.Balance = fromAccount.Balance - int64(newAccountPayload.Amount)

	if err := s.store.UpdateAccount(toAccount); err != nil {
		return err
	}

	if err := s.store.UpdateAccount(fromAccount); err != nil {
		return err
	}

	return writeJson(w, http.StatusOK, "transfer complete")
}

func permissionDenied(w http.ResponseWriter) {
	writeJson(w, http.StatusForbidden, &apiError{Error: "permission denied"})

}
