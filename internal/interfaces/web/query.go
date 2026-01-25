package web

import (
	"fmt"
	"net/http"
)

func GetRequiredQueryParam(r *http.Request, key string) (string, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return "", fmt.Errorf("required query parameter '%s' is missing", key)
	}
	return value, nil
}

func GetOptionalQueryParam(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}
