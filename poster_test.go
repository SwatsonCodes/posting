package main

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/swatsoncodes/posting/db"
	"github.com/swatsoncodes/posting/models"
)

type mockPostsDB struct{ shouldErr bool }

func (m mockPostsDB) PutPost(post models.Post) (err error) {
	if m.shouldErr {
		return errors.New("i am a failure")
	}
	return
}

func (m mockPostsDB) GetPosts(offset, limit int) (posts *[]models.Post, isMore bool, err error) {
	if m.shouldErr {
		return nil, true, errors.New("i am a failure")
	}
	return &[]models.Post{models.Post{}}, true, nil
}

type mockUploader struct{ shouldErr bool }

func (m mockUploader) UploadMedia(media io.Reader) (string, error) {
	if m.shouldErr {
		return "", errors.New("i am bad at uploading media")
	}
	return "http://www.example.com", nil
}

func makeFormData(fields map[string]string, files map[string][]byte) (io.Reader, string) {
	bod := &bytes.Buffer{}
	writer := multipart.NewWriter(bod)
	defer writer.Close()
	for k, v := range fields {
		writer.WriteField(k, v)
	}
	for k, f := range files {
		w, _ := writer.CreateFormFile(k, k)
		io.Copy(w, bytes.NewReader(f))
	}
	return bod, writer.FormDataContentType()
}

func TestCreatePost(t *testing.T) {
	var happyDB db.PostsDB = mockPostsDB{shouldErr: false}
	var sadDB db.PostsDB = mockPostsDB{shouldErr: true}
	var happyUploader models.Uploader = mockUploader{shouldErr: false}
	var sadUploader models.Uploader = mockUploader{shouldErr: true}
	testcases := []struct {
		fields     map[string]string
		files      map[string][]byte
		db         *db.PostsDB
		uploader   models.Uploader
		statusCode int
	}{
		{
			map[string]string{
				"Body": "hello",
			},
			nil,
			&happyDB,
			happyUploader,
			http.StatusOK,
		},
		{
			nil,
			map[string][]byte{
				"Pics": []byte("a"),
			},
			&happyDB,
			happyUploader,
			http.StatusOK,
		},
		{
			nil,
			map[string][]byte{
				"Pics": []byte("a"),
			},
			&happyDB,
			happyUploader,
			http.StatusOK,
		},
		{
			map[string]string{
				"Body": "hello",
			},
			map[string][]byte{
				"Pics": []byte("a"),
			},
			&happyDB,
			happyUploader,
			http.StatusOK,
		},
		{
			nil,
			nil,
			&happyDB,
			happyUploader,
			http.StatusBadRequest,
		},
		{
			map[string]string{
				"Body": "hello",
			},
			nil,
			&sadDB,
			happyUploader,
			http.StatusInternalServerError,
		},
		{
			map[string]string{
				"Body": "hello",
			},
			nil,
			&happyDB,
			sadUploader,
			http.StatusOK,
		},
		{
			nil,
			map[string][]byte{
				"Pics": []byte("a"),
			},
			&happyDB,
			sadUploader,
			http.StatusInternalServerError,
		},
		{
			map[string]string{
				"Body": "hello",
			},
			nil,
			&sadDB,
			sadUploader,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		poster := Poster{
			DB:            testcase.db,
			Uploader:      testcase.uploader,
			PostsTemplate: template.Must(template.ParseFiles("templates/posts.html")),
		}
		bod, ct := makeFormData(testcase.fields, testcase.files)
		req, _ := http.NewRequest(http.MethodPost, "/test", bod)
		req.Header.Set("Content-Type", ct)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(poster.CreatePost)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, testcase.statusCode, rr.Code)
	}

}

func TestGetPosts(t *testing.T) {
	var happyDB db.PostsDB = mockPostsDB{shouldErr: false}
	var sadDB db.PostsDB = mockPostsDB{shouldErr: true}
	testcases := []struct {
		db         *db.PostsDB
		statusCode int
	}{
		{&happyDB, http.StatusOK},
		{&sadDB, http.StatusInternalServerError},
	}

	for _, testcase := range testcases {
		poster := Poster{
			DB:            testcase.db,
			PostsTemplate: template.Must(template.ParseFiles("templates/posts.html")),
		}
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(poster.GetPosts)
		handler.ServeHTTP(rr, req)
		assert.Equal(t, testcase.statusCode, rr.Code)
	}
}

/*
func TestIsRequestAuthorized(t *testing.T) {
	// based on https://www.twilio.com/docs/security#validating-requests
	poster := Poster{
		TwilioAuthToken: "12345",
	}
	_url := "https://mycompany.com/myapp.php?foo=1&bar=2"
	validSender := "+12349013030"
	validForm := url.Values{
		"CallSid": {"CA1234567890ABCDE"},
		"Caller":  {"+12349013030"},
		"Digits":  {"1234"},
		"From":    {validSender},
		"To":      {"+18005551212"},
	}
	validHeaders := map[string]string{"X-Twilio-Signature": "0/KCTR6DLpKmkAf8muzZqo1nDgQ="}

	testcases := []struct {
		form          url.Values
		headers       map[string]string
		allowedSender string
		isAuthorized  bool
	}{
		{
			validForm,
			validHeaders,
			validSender,
			true,
		},
		{
			validForm,
			map[string]string{},
			validSender,
			false,
		},
		{
			validForm,
			validHeaders,
			"some other sender",
			false,
		},
		{
			url.Values{
				"foo": {"bar"},
			},
			validHeaders,
			validSender,
			false,
		},
	}

	for _, testcase := range testcases {
		poster.AllowedSender = testcase.allowedSender
		req, _ := http.NewRequest(http.MethodPost, _url, nil)
		req.PostForm = testcase.form
		for k, v := range testcase.headers {
			req.Header.Set(k, v)
		}
		assert.Equal(t, testcase.isAuthorized, poster.IsRequestAuthorized(req))
	}
}
*/
