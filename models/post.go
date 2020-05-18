package models

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const createdAtFmt = "2 Jan 2006 15:04"

type Post struct {
	ID        string    `json:"post_id" firestore:"post_id"`
	Body      string    `json:"body" firestore:"body"`
	MediaURLs []string  `json:"media_urls,omitempty" firestore:"media_urls,omitempty"`
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
}

func ParsePost(form *url.Values) (post *Post, err error) {
	var id, numMedia string
	post = &Post{CreatedAt: time.Now()}
	if id = form.Get("SmsSid"); id == "" {
		return nil, errors.New("SmsSid field not present")
	}
	post.ID = id

	post.Body = form.Get("Body")

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

func (post *Post) ResolveMediaURLs() {
	var wg sync.WaitGroup
	for i, url := range post.MediaURLs {
		wg.Add(1)
		go func(i int, url string) {
			if resp, err := http.Get(url); err == nil {
				post.MediaURLs[i] = resp.Request.URL.String()
			}
			wg.Done()
		}(i, url)
	}
	wg.Wait()
}

func (post *Post) FmtCreatedAt() string {
	return post.CreatedAt.Format(createdAtFmt)
}
