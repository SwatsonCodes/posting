package imgur

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const maxBodySizeBytes = 2 * 1024 // 2KiB

type Uploader struct {
	ClientID string
}
type imgurResponse struct {
	Data struct{ Link string }
}

// TODO: DRY this out
func (uploader Uploader) UploadMedia(media io.Reader) (imgurURL string, err error) {
	var client http.Client
	var bod imgurResponse

	req, err := http.NewRequest("POST",
		"https://api.imgur.com/3/image",
		media,
	)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", "Client-ID "+uploader.ClientID)
	req.Header.Add("Content-Type", "multipart/form-data")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("non-200 status code received from imgur: %d", resp.StatusCode))
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

func (uploader Uploader) RehostImage(currentURL string) (imgurURL string, err error) {
	var client http.Client
	var bod imgurResponse

	req, err := http.NewRequest("POST",
		"https://api.imgur.com/3/image",
		strings.NewReader(url.Values{"image": {currentURL}}.Encode()),
	)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", "Client-ID "+uploader.ClientID)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("non-200 status code received from imgur: %d", resp.StatusCode))
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
