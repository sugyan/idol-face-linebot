package inferences

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
)

var endpointBase = os.Getenv("INFERENCES_API_ENDPOINT")

// BulkFetch function
func BulkFetch(userID string) ([]inference, error) {
	ch := make(chan []inference)
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			res, err := fetch(userID, page)
			if err != nil {
				log.Println(err)
				return
			}
			ch <- res.Inferences
		}(i + 1)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	inferences := make([]inference, 0, 500)
	for results := range ch {
		inferences = append(inferences, results...)
	}
	return inferences, nil
}

// Accept function
func Accept(userID, inferenceID string) error {
	url := endpointBase + "/inferences/" + inferenceID + "/accept.json?"
	res, err := do("POST", url, userID)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	log.Println(string(bytes))
	return nil
}

func fetch(userID string, page int) (*result, error) {
	values := url.Values{}
	values.Add("page", strconv.Itoa(page))
	url := endpointBase + "/inferences.json?" + values.Encode()
	res, err := do("GET", url, userID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	result := &result{}
	if err = json.NewDecoder(res.Body).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

func do(method, url, userID string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-User-Token", os.Getenv("INFERENCES_API_TOKEN"))
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
