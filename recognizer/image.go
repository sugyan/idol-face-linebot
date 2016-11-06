package recognizer

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

// RecognizeFaces function
func (c *Client) RecognizeFaces(photoURL string) error {
	// fetch from recognizer API
	values := url.Values{}
	values.Set("image_url", photoURL)
	u := *c.EndPointBase
	u.Path = path.Join(c.EndPointBase.Path, "recognizer", "faces")
	u.RawQuery = values.Encode()
	res, err := c.do("GET", u.String(), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	// TODO
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	println(string(b))
	return nil
}

type recognizeResult struct {
	Faces []struct {
		Bounding []struct {
			X int `json:"x"`
			Y int `json:"y"`
		} `json:"bounding"`
		Angle struct {
			Roll  float64 `json:"roll"`
			Yaw   float64 `json:"yaw"`
			Pitch float64 `json:"pitch"`
		} `json:"angle"`
		Recognize []struct {
			Label struct {
				Description string `json:"description"`
				FacesCount  int    `json:"faces_count"`
				ID          int    `json:"id"`
				IndexNumber int    `json:"index_number"`
				LabelURL    string `json:"label_url"`
				Name        string `json:"name"`
				Twitter     string `json:"twitter"`
			} `json:"label"`
			Value float64 `json:"value"`
		} `json:"recognize"`
	} `json:"faces"`
	Message string `json:"message"`
}
