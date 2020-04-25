package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/swatsoncodes/very-nice-website/db"
	"github.com/swatsoncodes/very-nice-website/models"
)

type mockPostsDB struct{ shouldErr bool }

func (m mockPostsDB) PutPost(post models.Post) (err error) {
	if m.shouldErr {
		return errors.New("i am a failure")
	}
	return
}

func (m mockPostsDB) GetPosts() (posts *[]models.Post, err error) {
	if m.shouldErr {
		return nil, errors.New("i am a failure")
	}
	return &[]models.Post{models.Post{}}, nil
}

func TestGoAway(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GoAway)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Body.String(), cowsay)
	assert.Equal(t, rr.Code, http.StatusOK)
}

func TestCreatePost(t *testing.T) {
	var happyDB db.PostsDB = mockPostsDB{shouldErr: false}
	var sadDB db.PostsDB = mockPostsDB{shouldErr: true}
	testcases := []struct {
		body        string
		contentType string
		db          *db.PostsDB
		statusCode  int
	}{
		{
			url.Values{
				"SmsSid":   []string{"abc123"},
				"Body":     []string{"hello"},
				"NumMedia": []string{"0"},
			}.Encode(),
			"application/x-www-form-urlencoded",
			&happyDB,
			http.StatusCreated,
		},
		{
			"bad bod",
			"application/x-www-form-urlencoded",
			&happyDB,
			http.StatusBadRequest,
		},
		{
			url.Values{
				"SmsSid":   []string{"abc123"},
				"Body":     []string{"hello"},
				"NumMedia": []string{"0"},
			}.Encode(),
			"text/plain; boundary=",
			&happyDB,
			http.StatusBadRequest,
		},
		{
			url.Values{
				"SmsSid":   []string{"abc123"},
				"Body":     []string{"hello"},
				"NumMedia": []string{"0"},
			}.Encode(),
			"application/x-www-form-urlencoded",
			&sadDB,
			http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		poster := Poster{DB: testcase.db}
		req, _ := http.NewRequest(http.MethodPost, "/test", strings.NewReader(testcase.body))
		req.Header.Set("Content-Type", testcase.contentType)
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
		poster := Poster{DB: testcase.db, PostsTemplate: template.Must(template.ParseFiles("templates/posts.html"))}
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(poster.GetPosts)
		handler.ServeHTTP(rr, req)
		assert.Equal(t, testcase.statusCode, rr.Code)
	}
}

func TestGetExpectedTwilioSignature(t *testing.T) {
	// based on https://www.twilio.com/docs/security#validating-requests
	_url := "https://mycompany.com/myapp.php?foo=1&bar=2"
	authToken := "12345"
	expectedSignature := "0/KCTR6DLpKmkAf8muzZqo1nDgQ="
	postForm := url.Values{
		"CallSid": {"CA1234567890ABCDE"},
		"Caller":  {"+12349013030"},
		"Digits":  {"1234"},
		"From":    {"+12349013030"},
		"To":      {"+18005551212"},
	}

	assert.Equal(t, expectedSignature, GetExpectedTwilioSignature(_url, authToken, postForm))
	postForm["New"] = []string{"data"}
	assert.NotEqual(t, expectedSignature, GetExpectedTwilioSignature(_url, authToken, postForm))
}

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
