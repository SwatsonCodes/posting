package models

import (
	"mime/multipart"
	"sync"
	"time"

	"github.com/google/uuid"
)

const createdAtFmt = "2 Jan 2006 15:04"

type Post struct {
	ID        string    `json:"post_id" firestore:"post_id"`
	Body      string    `json:"body" firestore:"body"`
	MediaURLs *[]string `json:"media_urls,omitempty" firestore:"media_urls,omitempty"`
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
	media     *[]multipart.File
}

type Uploader interface {
	UploadMedia(media multipart.File) (mediaURL string, err error)
}

func ParsePost(form *multipart.Form) (post *Post, err error) {
	post = &Post{CreatedAt: time.Now(), ID: uuid.New().String()}
	if bod, ok := form.Value["Body"]; ok {
		post.Body = bod[0]
	}

	if images, ok := form.File["Pics"]; ok {
		media := make([]multipart.File, len(images))
		// TODO: do this in parallel
		for i, image := range images {
			//TODO: validate media filetype
			//TODO: validate media file size

			f, err := image.Open()
			if err != nil {
				return nil, err
			}
			media[i] = f
		}
		post.media = &media
	}

	return
}

func (post *Post) UploadMedia(uploader Uploader) error {
	if post.media == nil {
		return nil
	}

	var wg sync.WaitGroup
	mediaURLs := make([]string, len(*post.media))
	done := make(chan bool, 0)
	errors := make(chan error, 1)

	for i, f := range *(post.media) {
		wg.Add(1)
		go func(i int, f multipart.File) {
			url, err := uploader.UploadMedia(f)
			if err != nil {
				errors <- err
				return
			}
			mediaURLs[i] = url
			wg.Done()
		}(i, f)
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

	post.MediaURLs = &mediaURLs
	return nil
}

func (post *Post) FmtCreatedAt() string {
	return post.CreatedAt.Format(createdAtFmt)
}
