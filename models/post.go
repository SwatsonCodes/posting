package models

import (
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/swatsoncodes/posting/upstream/imgur"
)

const createdAtFmt = "2 Jan 2006 15:04"

type Post struct {
	ID        string    `json:"post_id" firestore:"post_id"`
	Body      string    `json:"body" firestore:"body"`
	MediaURLs []string  `json:"media_urls,omitempty" firestore:"media_urls,omitempty"`
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
}

func ParsePost(form *url.Values) (post *Post, err error) {
	post = &Post{CreatedAt: time.Now(), ID: uuid.New().String()}
	post.Body = form.Get("Body")
	return
}

func (post *Post) RehostImagesOnImgur(uploader imgur.Uploader) error {
	var wg sync.WaitGroup
	imgurURLs := make([]string, len(post.MediaURLs))
	done := make(chan bool, 0)
	errors := make(chan error, 1)

	for i, url := range post.MediaURLs {
		wg.Add(1)
		go func(i int, url string) {
			imgurURL, err := uploader.UploadImage(url)
			if err != nil {
				errors <- err
				return
			}
			imgurURLs[i] = imgurURL
			wg.Done()
		}(i, url)
	}
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		break
	case err := <-errors:
		return err
	}

	for i, url := range imgurURLs {
		post.MediaURLs[i] = url
	}
	return nil
}

func (post *Post) FmtCreatedAt() string {
	return post.CreatedAt.Format(createdAtFmt)
}
