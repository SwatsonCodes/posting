package models

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

type Post struct {
	ID        string
	Body      string
	MediaURLs *[]string
}

func ParsePost(form *url.Values) (post *Post, err error) {
	var id, body, numMedia string
	post = new(Post)
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
	post.MediaURLs = &mediaURLs
	return
}
