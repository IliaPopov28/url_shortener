package api

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrInvalidStatusCode = errors.New("invalid status code")
)

func GetRedirectedURL(url string) (string, error) {
	const op = "api.GetRedirectedURL"

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("%s: %w :%d", op, ErrInvalidStatusCode, resp.StatusCode)
	}

	return resp.Header.Get("Location"), nil
}
