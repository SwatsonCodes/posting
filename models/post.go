package models

import (
	"errors"
	"io"
	"mime/multipart"
	"sync"
	"time"

	"github.com/google/uuid"
)

const createdAtFmt = "2 Jan 2006 15:04"

type Post struct {
	ID        string    `json:"post_id" firestore:"post_id"`
	Body      string    `json:"body" firestore:"body"`
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
	MediaURLs []string  `json:"media_urls,omitempty" firestore:"media_urls,omitempty"`
	media     []io.Reader
}

type Uploader interface {
	UploadMedia(media io.Reader) (mediaURL string, err error)
}

func ParsePost(form multipart.Form) (post *Post, err error) {
	post = &Post{
		CreatedAt: time.Now(),
		ID:        uuid.New().String(),
	}
	if bod, ok := form.Value["Body"]; ok && len(bod) > 0 {
		post.Body = bod[0]
	}

	if pics, ok := form.File["Pics"]; ok {
		media := make([]io.Reader, len(pics))
		for i, pic := range pics {
			//TODO: validate media filetype
			//TODO: validate media file size
			f, err := pic.Open()
			if err != nil {
				return nil, err
			}
			media[i] = f
		}
		post.media = media
	}
	if post.Body == "" && len(post.media) == 0 {
		return nil, errors.New("Post must contain a body or at least one media")
	}
	return
}

func (post *Post) UploadMedia(uploader Uploader) error {
	if post.media == nil {
		return nil
	}

	var wg sync.WaitGroup
	mediaURLs := make([]string, len(post.media))
	done := make(chan bool, 0)
	errors := make(chan error, 1)

	for i, media := range post.media {
		wg.Add(1)
		go func(i int, m io.Reader) {
			url, err := uploader.UploadMedia(m)
			if err != nil {
				errors <- err
				return
			}
			mediaURLs[i] = url
			wg.Done()
		}(i, media)
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

	post.MediaURLs = mediaURLs
	return nil
}

func (post *Post) FmtCreatedAt() string {
	return post.CreatedAt.Format(createdAtFmt)
}
