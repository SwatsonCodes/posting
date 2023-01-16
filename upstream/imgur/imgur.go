// Package imgur contains tooling for interacting with the Imgur API
package imgur

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const maxBodySizeBytes = 2 * 1024 // 2KiB
const imgurAPIURL = "https://api.imgur.com/3/image"

// Uploader is a simple struct containing the info necessary to communicate with the imgur API
type Uploader struct {
	ClientID string
}
type imgurResponse struct {
	Data struct{ Link string }
}

func (uploader Uploader) postToImgur(media io.Reader, contentType string) (imgurURL string, err error) {
	var client http.Client
	var bod imgurResponse

	req, err := http.NewRequest("POST", imgurAPIURL, media)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", "Client-ID "+uploader.ClientID)
	req.Header.Add("Content-Type", contentType)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("non-200 status code received from imgur: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, maxBodySizeBytes))
	if err != nil {
		return
	}
	if err = json.Unmarshal(body, &bod); err != nil {
		return
	}

	return bod.Data.Link, nil
}

// UploadMedia attempts to upload the given media to the imgur API.
// It returns the URL where the media can be found, or an error.
func (uploader Uploader) UploadMedia(media io.Reader) (imgurURL string, err error) {
	return uploader.postToImgur(media, "multipart/form-data")
}

// RehostImage takes a URL pointing to an image and attempts to rehost that image on imgur via the API.
// It returns the URL where the image can be found, or an error.
func (uploader Uploader) RehostImage(currentURL string) (imgurURL string, err error) {
	return uploader.postToImgur(strings.NewReader(url.Values{"image": {currentURL}}.Encode()), "application/x-www-form-urlencoded")
}
