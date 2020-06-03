package twilio

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	log "github.com/sirupsen/logrus"
)

var xmlHead = []byte(`<?xml version="1.0" encoding="UTF-8"?>`)

type response struct {
	XMLName xml.Name `xml:"Response"`
	Message message
}

type message struct {
	Body string
}

func FormatResponse(body string) (resp []byte, err error) {
	r := &response{Message: message{Body: body}}
	resp, err = xml.Marshal(r)
	if err != nil {
		return
	}
	return append(xmlHead, resp...), nil
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

func IsRequestSigned(r *http.Request, authToken string) bool {
	err := r.ParseForm()
	if err != nil {
		log.WithError(err).Error("failed to parse form body")
		return false
	}

	if sig := r.Header.Get("X-Twilio-Signature"); sig != "" {
		if sig != GetExpectedTwilioSignature(
			getClientURL(r),
			authToken,
			r.PostForm,
		) {
			return false
		}
	} else {
		return false
	}

	return true
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
