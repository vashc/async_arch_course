package internal

import (
	"log"
	"net/http"
)

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("Request URI: %s\n", r.RequestURI)
		next.ServeHTTP(rw, r.WithContext(r.Context()))
	})
}
