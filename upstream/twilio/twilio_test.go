package twilio

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestIsRequestSigned(t *testing.T) {
	// based on https://www.twilio.com/docs/security#validating-requests
	twilioAuthToken := "12345"
	_url := "https://mycompany.com/myapp.php?foo=1&bar=2"
	validForm := url.Values{
		"CallSid": {"CA1234567890ABCDE"},
		"Caller":  {"+12349013030"},
		"Digits":  {"1234"},
		"From":    {"+12349013030"},
		"To":      {"+18005551212"},
	}
	validHeaders := map[string]string{"X-Twilio-Signature": "0/KCTR6DLpKmkAf8muzZqo1nDgQ="}

	testcases := []struct {
		form         url.Values
		headers      map[string]string
		isAuthorized bool
	}{
		{
			validForm,
			validHeaders,
			true,
		},
		{
			validForm,
			map[string]string{},
			false,
		},
		{
			url.Values{
				"foo": {"bar"},
			},
			validHeaders,
			false,
		},
	}

	for _, testcase := range testcases {
		req, _ := http.NewRequest(http.MethodPost, _url, nil)
		req.PostForm = testcase.form
		for k, v := range testcase.headers {
			req.Header.Set(k, v)
		}
		assert.Equal(t, testcase.isAuthorized, IsRequestSigned(req, twilioAuthToken))
	}
}
