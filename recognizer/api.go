package recognizer

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

var endpointBase = os.Getenv("RECOGNIZER_API_ENDPOINT")

// Labels function
func Labels(userID, token, query string) ([]label, error) {
	values := url.Values{}
	values.Add("q", query)
	url := endpointBase + "/labels.json?" + values.Encode()
	res, err := do("GET", url, userID, token)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	results := []label{}
	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		return nil, err
	}
	return results, nil
}

// Inferences function
func Inferences(userID, token string, ids []int) ([]inference, error) {
	values := url.Values{}
	for _, id := range ids {
		values.Add("id", strconv.Itoa(id))
	}
	// values.Add("q", query)
	url := endpointBase + "/inferences.json?" + values.Encode()
	res, err := do("GET", url, userID, token)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	result := &struct {
		Inferences []inference `json:"inferences"`
	}{}
	if err = json.NewDecoder(res.Body).Decode(result); err != nil {
		return nil, err
	}
	return result.Inferences, nil
}

// AcceptInference function
func AcceptInference(userID, token, inferenceID string) (string, error) {
	url := endpointBase + "/inferences/" + inferenceID + "/accept.json?"
	res, err := do("POST", url, userID, token)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	result := &struct {
		Result  string `json:"result"`
		FaceURL string `json:"face_url"`
	}{}
	if err = json.NewDecoder(res.Body).Decode(result); err != nil {
		return "", nil
	}
	return result.FaceURL, nil
}

func do(method, url, userID, token string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-User-Token", token)
	req.Header.Set("X-User-Email", userID+"@line.me")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}
	return res, err
}
