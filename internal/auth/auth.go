package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	apiKey := headers.Get("Authorization")
	if apiKey == "" {
		return "", errors.New("no API key provided")
	}
	val := strings.Split(apiKey, " ")
	if len(val) != 2 {
		return "", errors.New("malformed API key")
	}
	if val[0] != "ApiKey" {
		return "", errors.New("malformed API key")
	}
	return val[1], nil
}
