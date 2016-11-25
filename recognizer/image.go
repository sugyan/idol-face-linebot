package recognizer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"path"
)

// RecognizeFaces function
func (c *Client) RecognizeFaces(contentType string, data []byte) (*RecognizedResults, error) {
	u := *c.EndPointBase
	u.Path = path.Join(c.EndPointBase.Path, "recognizer", "image.json")
	entity := struct {
		Image string `json:"image"`
	}{
		Image: "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data),
	}
	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(entity); err != nil {
		return nil, err
	}
	res, err := c.do("POST", u.String(), buf)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}
	result := &RecognizedResults{}
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

// RecognizedResults type
type RecognizedResults struct {
	Faces   []RecognizedFace `json:"faces"`
	Message string           `json:"message"`
}

// RecognizedFace type
type RecognizedFace struct {
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
}

// ByTopValue implements sort.Interface for []RecognizedFace by top result's value in descending order
type ByTopValue []RecognizedFace

// Len method of ByTopValue
func (faces ByTopValue) Len() int {
	return len(faces)
}

// Swap method of ByTopValue
func (faces ByTopValue) Swap(i, j int) {
	faces[i], faces[j] = faces[j], faces[i]
}

// Less method of ByTopValue compares value of top result
func (faces ByTopValue) Less(i, j int) bool {
	return faces[j].Recognize[0].Value < faces[i].Recognize[0].Value
}
