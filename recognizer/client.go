package recognizer

import (
	"errors"
	"net/http"
	"net/url"
)

// Client type
type Client struct {
	HTTPClient          *http.Client
	EndPointBase        *url.URL
	AuthenticationEmail string
	AuthenticationToken string
}

// NewClient function
func NewClient(endpoint, email, token string) (*Client, error) {
	if len(email) == 0 {
		return nil, errors.New("missing email")
	}
	if len(token) == 0 {
		return nil, errors.New("missing token")
	}
	parsedURL, err := url.ParseRequestURI(endpoint)
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
