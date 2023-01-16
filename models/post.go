// Package models provides types for modeling Posts
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

// A Post consists primarily of a text blurb and/or one or more media (i.e. pictures).
// Note that the MediaURLs field is not populated until UploadMedia() is called.
type Post struct {
	ID        string    `json:"post_id" firestore:"post_id"`                           // unique ID
	Body      string    `json:"body" firestore:"body"`                                 // text blurb (optional)
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`                     // time Post was created
	MediaURLs []string  `json:"media_urls,omitempty" firestore:"media_urls,omitempty"` // colletion of HTTP URLs to associated media (optional).
	media     []io.Reader
}

// An Uploader uploads media (i.e. an image) to an external host and returns the URL where the media can be accessed.
type Uploader interface {
	UploadMedia(media io.Reader) (mediaURL string, err error)
}

// ParsePost attempts to parse incoming post from an HTTP form and returns the resultant Post object
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

// UploadMedia uploads the media associated with a Post (if any) to an external host by invoking the given Uploader.
// The resultant URLs where the media can be accessed are stored in the Post.MediaURLs field.
// This method invokes the Uploader asynchronously on each piece of media associated with the Post, and blocks until all uploads are complete.
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

// FmtCreatedAt returns the Post's CreatedAt field as a human-readable string.
func (post *Post) FmtCreatedAt() string {
	return post.CreatedAt.Format(createdAtFmt)
}
