package recognizer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
)

// RecognizeFaces function
func (c *Client) RecognizeFaces(photoURL string) error {
	// convert to jpeg image
	file, err := ioutil.TempFile("", "image")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	err = exec.Command("convert", photoURL, "jpeg:"+file.Name()).Run()
	if err != nil {
		return err
	}

	// create post data
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer([]byte{})
	if err = json.NewEncoder(buf).Encode(struct {
		Image string `json:"image"`
	}{
		Image: "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data),
	}); err != nil {
		return err
	}
	// post to API
	u := *c.EndPointBase
	u.Path = path.Join(c.EndPointBase.Path, "recognizer", "api")
	res, err := c.do("POST", u.String(), buf)
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
