package main

import (
	"encoding/json"
	"fmt"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strconv"
	"time"
)

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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

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
