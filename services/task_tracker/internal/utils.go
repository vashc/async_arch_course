package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func BodyParser(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return ErrUnsupportedMediaType
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<10)

	dec := json.NewDecoder(r.Body)
	// Return error on any fields mismatches
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrRequestBodyDeconding, err.Error())
	}

	return nil
}

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
