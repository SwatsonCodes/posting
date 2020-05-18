package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"sync"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/swatsoncodes/posting/db"
	"github.com/swatsoncodes/posting/models"
)

const cowsay string = `
 _____
< no  >
 -----
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
`
const postsTemplate string = "posts.html"

type Poster struct {
	AllowedSender   string
	TwilioAuthToken string
	DB              *db.PostsDB
	PostsTemplate   *template.Template
}

func NewPoster(allowedSender, twilioAuthToken, templatesPath string, postsDB *db.PostsDB) (*Poster, error) {
	template, err := template.ParseFiles(filepath.Join(templatesPath, postsTemplate))
	if err != nil {
		return nil, err
	}
	return &Poster{allowedSender, twilioAuthToken, postsDB, template}, nil
}

func GoAway(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, cowsay)
}

func (poster Poster) CreatePost(w http.ResponseWriter, r *http.Request) {
	// TODO: reply with twilio-friendly response
	err := r.ParseForm()
	if err != nil {
		log.WithError(err).Warn("failed to parse form body")
		http.Error(w, "unable to parse request form body", http.StatusBadRequest)
		return
	}

	post, err := models.ParsePost(&r.PostForm)
	if err != nil {
		log.WithError(err).Warn("got bad post")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := (*poster.DB).PutPost(*post); err != nil {
		log.WithError(err).Error("failed to put post to DB")
		http.Error(w, "unable to save post", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
}

func (poster Poster) GetPosts(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup

	posts, err := (*poster.DB).GetPosts()
	if err != nil {
		log.WithError(err).Error("failed to get posts from db")
		http.Error(w, "unable to retrieve posts", http.StatusInternalServerError)
		return
	}

	for i := range *posts {
		wg.Add(1)
		go func(post *models.Post) {
			post.ResolveMediaURLs()
			wg.Done()
		}(&(*posts)[i])
	}
	wg.Wait()

	if err = poster.PostsTemplate.Execute(w, *posts); err != nil {
		log.WithError(err).Error(err.Error())
		http.Error(w, "unable to render posts html", http.StatusInternalServerError)
		return
	}
}

func GetExpectedTwilioSignature(url, authToken string, postForm url.Values) (expectedTwilioSignature string) {
	var i int
	var buffer bytes.Buffer
	var postFormLen = len(postForm)
	keys := make([]string, postFormLen)

	// sort keys in request form body
	for key := range postForm {
		keys[i] = key
		i++
	}
	sort.Strings(keys)

	// append sorted key/val pairs to url in order
	buffer.WriteString(url)
	for _, key := range keys {
		buffer.WriteString(key)
		buffer.WriteString(postForm[key][0])
	}
	// sign with HMAC-SHA1 using auth token
	mac := hmac.New(sha1.New, []byte(authToken))
	mac.Write(buffer.Bytes())
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (poster Poster) IsRequestAuthorized(r *http.Request) bool {
	err := r.ParseForm()
	if err != nil {
		log.WithError(err).Error("failed to parse form body")
		return false
	}

	if sig := r.Header.Get("X-Twilio-Signature"); sig != "" {
		if sig != GetExpectedTwilioSignature(
			getClientURL(r),
			poster.TwilioAuthToken,
			r.PostForm,
		) {
			return false
		}
	} else {
		return false
	}

	return r.PostForm.Get("From") == poster.AllowedSender
}

func getClientURL(r *http.Request) string {
	var scheme, host string
	if scheme = r.Header.Get("X-Forwarded-Proto"); scheme != "" {
		goto GetHost
	}
	if scheme = r.Header.Get("X-Forwarded-Scheme"); scheme != "" {
		goto GetHost
	}
	scheme = r.URL.Scheme

GetHost:
	// it appears that API Gateway Lambda proxy integration sets r.Host with its own value
	// so check the headers first for "real" host
	if host = r.Header.Get("Host"); host != "" {
		goto Done
	}
	if host = r.Host; host != "" {
		goto Done
	}
	host = r.URL.Host

Done:
	if r.URL.RawQuery == "" {
		return fmt.Sprintf("%s://%s%s", scheme, host, r.URL.Path)
	}
	return fmt.Sprintf("%s://%s%s?%s", scheme, host, r.URL.Path, r.URL.RawQuery)
}
