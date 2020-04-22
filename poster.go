package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/swatsoncodes/very-nice-website/db"
	"github.com/swatsoncodes/very-nice-website/models"
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

type Poster struct {
	AllowedSender   string
	TwilioAuthToken string
	DB              db.PostsDB
}

func GoAway(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, cowsay)
}

func (poster Poster) CreatePost(w http.ResponseWriter, r *http.Request) {
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

	if err := poster.DB.PutPost(*post); err != nil {
		log.WithError(err).Error("failed to put post to DB")
		http.Error(w, "unable to save post", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
}

func (poster Poster) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := poster.DB.GetPosts()
	if err != nil {
		log.WithError(err).Error("failed to get posts from db")
		http.Error(w, "unable to retrieve posts", http.StatusInternalServerError)
		return
	}

	// TODO: parallelize this
	for _, post := range *posts {
		post.ResolveMediaURLs()
	}

	resp, err := json.Marshal(*posts)
	if err != nil {
		log.WithError(err).Error("failed to marshal posts to json")
		http.Error(w, "unable to retrieve posts", http.StatusInternalServerError)
		return
	}

	// TODO: don't just serve raw json as response
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func GetExpectedTwilioSignature(url, authToken string, postForm url.Values) (expectedTwilioSignature string) {
	log.Info(url)
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
