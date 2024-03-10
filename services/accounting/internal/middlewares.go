package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/golang-jwt/jwt"
)

// LogRequest is for logging current handler URI
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("Request URI: %s\n", r.RequestURI)
		next.ServeHTTP(rw, r.WithContext(r.Context()))
	})
}

// MiddlewareUserAuth is an authorization middleware
func MiddlewareUserAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := ExtractToken(r)
		if err != nil {
			log.Printf("task_tracker.ExtractToken error: %s\n", err.Error())
			code := http.StatusInternalServerError
			if errors.Is(err, ErrUnathorizedUser) {
				code = http.StatusUnauthorized
			}

			http.Error(w, http.StatusText(code), code)
			return
		}

		//nolint:staticcheck,revive // It's ok for now
		ctx := context.WithValue(r.Context(), CtxAuthToken, token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// MiddlewareUserCtx is a middleware for getting user context
func MiddlewareUserCtx(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error
			tokenString := jwtauth.TokenFromHeader(r)

			var token *jwt.Token
			token, err = jwt.ParseWithClaims(
				tokenString,
				&JWTClaims{},
				func(token *jwt.Token) (interface{}, error) {
					// Algorithm type validation
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("%w: %v", ErrWrongSignMethod, token.Header["alg"])
					}

					return []byte(config.JWTSecret), nil
				},
			)
			if err != nil {
				log.Printf("jwt.ParseWithClaims error: %s\n", err.Error())
				code := http.StatusInternalServerError
				http.Error(w, http.StatusText(code), code)
				return
			}

			if !token.Valid {
				log.Println("token is not valid")
				code := http.StatusInternalServerError
				http.Error(w, http.StatusText(code), code)
				return
			}

			claims, ok := token.Claims.(*JWTClaims)
			if !ok {
				log.Println("token.Claims.(*JWTClaims) is not ok")
				code := http.StatusInternalServerError
				http.Error(w, http.StatusText(code), code)
				return
			}

			//nolint:staticcheck,revive // It's ok for now
			ctx := context.WithValue(r.Context(), requestParamUserID, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
