package poster

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/swatsoncodes/posting/upstream/imgur"
	"github.com/swatsoncodes/posting/upstream/twilio"
)

const postsTemplate string = "posts.html"
const createdAtFmt = "2 Jan 2006 15:04"
const okay, badRequest, internalErr string = "👍", "🚮 bad post!", "🔥 internal error"

// Poster is the primary class of the blog.
// It holds the necessary data to communicate with 3rd party APIs and render HTML templates.
// A Poster creates new Posts by receiving incoming webook requests from Twilio and storing them in the DB.
// It can display those Posts by retreiving them from the DB and rendering them in an HTML template
type Poster struct {
	AllowedSender   string // sole phone number allowed to create new posts
	TwilioAuthToken string
	ImgurUploader   imgur.Uploader // used for rehosting images on Imgur
	DB              *PostsDB
	PageSize        int                // number of posts to display on a single page
	PostsTemplate   *template.Template // html template for rendering Posts
}

// Post represents a single Post received from Twilio which originated from my phone via SMS
// Normally the Post struct and its associated methods live in their own file in a "models" package to keep things tidy and organized,
// but for the purposes of my Thorn application I've merged it into poster.go to provide a larger code sample.
type Post struct {
	ID        string    `json:"post_id" firestore:"post_id"`
	Body      string    `json:"body" firestore:"body"`
	MediaURLs []string  `json:"media_urls,omitempty" firestore:"media_urls,omitempty"`
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
}

// NewPoster creates a new Poster
func NewPoster(allowedSender, twilioAuthToken, imgurClientID, templatesPath string, pageSize int, postsDB *PostsDB) (*Poster, error) {
	template, err := template.ParseFiles(filepath.Join(templatesPath, postsTemplate))
	if err != nil {
		return nil, err
	}
	return &Poster{allowedSender, twilioAuthToken, imgur.Uploader{ClientID: imgurClientID}, postsDB, pageSize, template}, nil
}

// CreatePost creates a new Post by
//  1) parsing it from an HTTP POST form sent via Twilio webhook
//  2) rehosting the images associated with the Post (if any) on Imgur
//  3) saving the Post data to the DB
// The response it writes is forwarded to the sender's phone thanks to Twilio
func (poster Poster) CreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.WithError(err).Warn("unable to parse form body")
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	post, err := ParsePost(&r.PostForm)
	if err != nil {
		log.WithError(err).Warn("got bad post")
		http.Error(w, badRequest, http.StatusBadRequest)
		return
	}

	if err := post.RehostImagesOnImgur(poster.ImgurUploader); err != nil {
		log.WithError(err).Error("failed to upload images to imgur")
		http.Error(w, internalErr, http.StatusInternalServerError)
		return
	}

	if err := (*poster.DB).PutPost(*post); err != nil {
		log.WithError(err).Error("failed to save post to DB")
		http.Error(w, internalErr, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(okay))
}

// GetPosts retrieves Posts from the DB and renders them using the HTML template.
// It uses the "page" URL query param to determine which Posts to display
func (poster Poster) GetPosts(w http.ResponseWriter, r *http.Request) {
	curPage := getPageNum(r)
	posts, isMore, err := (*poster.DB).GetPosts(curPage*poster.PageSize, poster.PageSize)
	if err != nil {
		log.WithError(err).Error("failed to get posts from db")
		http.Error(w, internalErr, http.StatusInternalServerError)
		return
	}

	// NextPage and PrevPage are used for displaying HTML navigation buttons
	// If NextPage or PrevPage are < 0, it indicates there are no older or newer posts to fetch, respectively
	nextPage := -1
	if isMore {
		nextPage = curPage + 1
	}
	templatePayload := struct {
		Posts              []Post
		NextPage, PrevPage int
	}{
		*posts,
		nextPage,
		curPage - 1,
	}

	if err = poster.PostsTemplate.Execute(w, templatePayload); err != nil {
		log.WithError(err).Error(err.Error())
		http.Error(w, internalErr, http.StatusInternalServerError)
		return
	}
}

// IsRequestAuthorized determines if an incoming request originates from Twilio by checking the request signature
// it is intended to be configured as middleware by the server to protect the CreatePost endpoint (or any endpoint that should only be hit by Twilio)
func (poster Poster) IsRequestAuthorized(r *http.Request) bool {
	return twilio.IsRequestSigned(r, poster.TwilioAuthToken) && r.PostForm.Get("From") == poster.AllowedSender
}

// ParsePost attempts to deserialize an HTTP POST form into a Post object
// If the POST form or its fields are not well-formed, ParsePost returns an error
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

// RehostImagesOnImgur uploads a Post's associated images to Imgur and replaces its MediaURLs with the new Imgur URLs.
// Rehosting images on Imgur gives better latency than leeching off Twilio's CDN and avoids exposing our Twilio public key.
// Images are uploaded to Imgur asynchronously. If an image cannot be uploaded we exit immediately with an error.
// The Post's MediaURLs are not updated unless all uploads succeed.
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

	// wait until all images finish uploading, or exit immediately if we encounter an error
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

// FmtCreatedAt formats a Post's CreatedAt timestamp to a human-readable string
func (post *Post) FmtCreatedAt() string {
	return post.CreatedAt.Format(createdAtFmt)
}

// getPageNum determines which page number the requester wants using the "page" URL query param
// if "page" is not present, not an integer, or < 0, this function returns 0
func getPageNum(r *http.Request) (offset int) {
	if page, ok := r.URL.Query()["page"]; ok {
		if len(page) == 0 {
			return
		}
		if p, err := strconv.Atoi(page[0]); err == nil {
			if p < 0 {
				return
			}
			return p
		}
	}
	return
}
