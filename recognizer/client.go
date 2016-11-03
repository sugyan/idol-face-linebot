package recognizer

import (
	"errors"
	"net/http"
	"net/url"
	"os"
)

// Client type
type Client struct {
	HTTPClient          *http.Client
	EndPointBase        *url.URL
	AuthenticationEmail string
	AuthenticationToken string
}

// NewClient function
func NewClient(email, token string) (*Client, error) {
	if len(email) == 0 {
		return nil, errors.New("missing email")
	}
	if len(token) == 0 {
		return nil, errors.New("missing token")
	}
	parsedURL, err := url.Parse(os.Getenv("RECOGNIZER_API_ENDPOINT"))
	if err != nil {
		return nil, err
	}
	return &Client{
		HTTPClient:          &http.Client{},
		EndPointBase:        parsedURL,
		AuthenticationEmail: email,
		AuthenticationToken: token,
	}, nil
}
