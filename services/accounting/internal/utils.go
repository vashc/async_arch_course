package internal

import (
	"net/http"
	"strings"
)

func ExtractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get(HeaderAuth)
	if authHeader == "" {
		return "", ErrUnathorizedUser
	}

	if !strings.Contains(authHeader, HeaderBearer) {
		return "", ErrUnathorizedUser
	}

	token := strings.ReplaceAll(authHeader, HeaderBearer, "")

	return token, nil
}
