package utils

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type contextUserIDKey int

const ContextUserID contextUserIDKey = iota

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

var jwtKey = []byte("my_secret_key")

func generateJWT(id uuid.UUID, expirationTime time.Time) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},

			UserID: id,
		},
	)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var id uuid.UUID
			c, err := r.Cookie("token")
			if err != nil {
				//if r.URL.Path == "/api/user/urls" {
				//	w.WriteHeader(http.StatusUnauthorized)
				//	return
				//}
				if err == http.ErrNoCookie {
					id = uuid.New()
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				id, err = getUserID(c.Value)
				if err != nil {
					if r.URL.Path == "api/user/urls" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					id = uuid.New()
				}
			}
			expirationTime := time.Now().Add(5 * time.Minute)
			tokenString, err := generateJWT(id, expirationTime)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			http.SetCookie(
				w, &http.Cookie{
					Name:    "token",
					Value:   tokenString,
					Expires: expirationTime,
					Path:    "/",
				},
			)
			ctx := context.WithValue(r.Context(), ContextUserID, id)
			h.ServeHTTP(w, r.WithContext(ctx))
		},
	)
}

func getUserID(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtKey, nil
		},
	)
	if err != nil {
		return uuid.Nil, nil
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("token is not valid")
	}

	return claims.UserID, nil
}
