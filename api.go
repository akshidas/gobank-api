package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type ApiServer struct {
	port  string
	store Storage
}

func (s *ApiServer) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/", handler(root))
	r.HandleFunc("/login", handler(s.login))
	r.HandleFunc("/accounts", handler(s.handleAccounts))
	r.HandleFunc("/accounts/transfer", handler(s.transfer))
	r.HandleFunc("/accounts/{id}", withJWTAuth(handler(s.handleAccountById), s.store))

	log.Printf("🚀 Server starting on port %s\n", s.port)
	err := http.ListenAndServe(s.port, r)
	log.Printf("🔥 Server failed: %s\n", err)
}

func root(w http.ResponseWriter, r *http.Request) error {
	return writeJson(w, http.StatusOK, "🚀 the server is up and running")
}

func (s *ApiServer) login(w http.ResponseWriter, r *http.Request) error {
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

// Accounts
// Handler Routes
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
	toAccount, err := s.store.GetAccountById(newAccountPayload.ToAccount)
	if err != nil {
		return err
	}

	toAccount.Balance = int64(newAccountPayload.Amount)
	if err := s.store.UpdateAccount(toAccount); err != nil {
		return err
	}
	return writeJson(w, http.StatusOK, "transfer complete")
}

func permissionDenied(w http.ResponseWriter) {
	writeJson(w, http.StatusForbidden, &apiError{Error: "permission denied"})

}

// Middleware
func withJWTAuth(handlerFunction http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Calling JWT Auth")
		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)

		if err != nil {
			permissionDenied(w)
			return
		}

		if !token.Valid {
			permissionDenied(w)
			return

		}

		id, err := getId(r)

		if err != nil {
			permissionDenied(w)
			return
		}

		account, err := s.GetAccountById(id)
		if err != nil {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.Number != int64(claims["account_number"].(float64)) {
			permissionDenied(w)
			return
		}

		handlerFunction(w, r)
	}
}

// Helpers
func keyFun(token *jwt.Token) (interface{}, error) {
	secret := os.Getenv("JWT_SECRET")
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(secret), nil
}

func validateJWT(token string) (*jwt.Token, error) {
	return jwt.Parse(token, keyFun)
}

func createJwt(a Account) (string, error) {
	claims := &jwt.MapClaims{
		"expires_at":     jwt.NewNumericDate(time.Unix(1516239022, 0)),
		"account_number": a.Number,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

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

func getId(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	id := vars["id"]
	parsedId, err := strconv.Atoi(id)

	if err != nil {
		return parsedId, fmt.Errorf("invalid id given %s", id)
	}

	return parsedId, nil

}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)

	if err != nil {
		return "", fmt.Errorf("Failed to hash password")
	}
	return string(hashedPassword), nil

}

func isPasswordValid(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	if err == nil {
		return true
	}

	return false
}
