package models

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type Post struct {
	ID        string   `json:"post_id"`
	Body      string   `json:"body"`
	MediaURLs []string `json:"media_urls,omitempty"`
	CreatedAt string   `json:"created_at"`
}

func ParsePost(form *url.Values) (post *Post, err error) {
	var id, body, numMedia string
	post = &Post{CreatedAt: time.Now().Format(time.RFC3339)}
	if id = form.Get("SmsSid"); id == "" {
		return nil, errors.New("SmsSid field not present")
	}
	post.ID = id

	if body = form.Get("Body"); body == "" {
		return nil, errors.New("Body field not present")
	}
	post.Body = body

	if numMedia = form.Get("NumMedia"); numMedia == "" {
		return
	}

	nm, err := strconv.Atoi(numMedia)
	if err != nil {
		return nil, fmt.Errorf("NumMedia value '%s' could not be converted to integer", numMedia)
	}
	if nm <= 0 {
		return
	}
	mediaURLs := make([]string, nm)
	for i := 0; i < nm; i++ {
		if mediaURL := form.Get(fmt.Sprintf("MediaUrl%d", i)); mediaURL != "" {
			mediaURLs[i] = mediaURL
			continue
		}
		return nil, fmt.Errorf("NumMedia claims '%d' MediaURLs are present, but fewer were found", nm)
	}
	post.MediaURLs = mediaURLs
	return
}

func (post Post) ResolveMediaURLs() {
	for i, url := range post.MediaURLs {
		resp, err := http.Get(url)
		if err != nil {
			log.WithError(err).Warnf("failed to GET url '%s'", url)
			continue
		}

		post.MediaURLs[i] = resp.Request.URL.String()
	}
}
