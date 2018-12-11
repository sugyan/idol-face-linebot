package recognizer

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
)

// RegisterUser method
func (c *Client) RegisterUser(userID, displayName string) (string, error) {
	u := *c.EndPointBase
	u.Path = path.Join(c.EndPointBase.Path, "users.json")
	entity := struct {
		ScreenName string `json:"screen_name"`
		Email      string `json:"email"`
	}{
		ScreenName: displayName,
		Email:      userID + "@line.me",
	}
	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(entity); err != nil {
		return "", err
	}
	res, err := c.do("POST", u.String(), buf)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	result := struct {
		AuthenticationToken string `json:"authentication_token"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.AuthenticationToken, nil
}

// Labels method
func (c *Client) Labels(query string) ([]Label, error) {
	values := url.Values{}
	values.Add("q", query)
	u := *c.EndPointBase
	u.RawQuery = values.Encode()
	u.Path = path.Join(c.EndPointBase.Path, "labels.json")
	res, err := c.do("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	results := []Label{}
	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		return nil, err
	}
	return results, nil
}

func (c *Client) do(method, urlStr string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-User-Email", c.AuthenticationEmail)
	req.Header.Set("X-User-Token", c.AuthenticationToken)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}
	return res, err
}
