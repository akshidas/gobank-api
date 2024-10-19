package main

import (
	jwt "github.com/golang-jwt/jwt/v5"
	"net/http"
)

func withJWTAuth(handlerFunction http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
