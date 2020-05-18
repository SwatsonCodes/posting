package models

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecodePost(t *testing.T) {
	testcases := []struct {
		form      *url.Values
		shouldErr bool
		id        string
		body      string
		mediaURLs []string
	}{
		{
			&url.Values{
				"SmsSid": []string{"abc123"},
				"Body":   []string{"hello"},
			},
			false,
			"abc123",
			"hello",
			nil,
		},
		{
			&url.Values{
				"SmsSid":   []string{"abc123"},
				"Body":     []string{"hello"},
				"NumMedia": []string{"0"},
			},
			false,
			"abc123",
			"hello",
			nil,
		},
		{
			&url.Values{
				"SmsSid":    []string{"abc123"},
				"Body":      []string{"hello"},
				"NumMedia":  []string{"1"},
				"MediaUrl0": []string{"http://www.example.com"},
			},
			false,
			"abc123",
			"hello",
			[]string{"http://www.example.com"},
		},
		{
			&url.Values{
				"SmsSid":    []string{"abc123"},
				"Body":      []string{"hello"},
				"NumMedia":  []string{"3"},
				"MediaUrl0": []string{"http://example.com/0"},
				"MediaUrl1": []string{"http://example.com/1"},
				"MediaUrl2": []string{"http://example.com/2"},
			},
			false,
			"abc123",
			"hello",
			[]string{"http://example.com/0", "http://example.com/1", "http://example.com/2"},
		},
		{
			&url.Values{
				"SmsSid":    []string{"abc123"},
				"Body":      []string{"hello"},
				"NumMedia":  []string{"1"},
				"MediaUrl0": []string{"http://example.com/0"},
				"MediaUrl1": []string{"http://example.com/1"},
			},
			false,
			"abc123",
			"hello",
			[]string{"http://example.com/0"},
		},
		{
			&url.Values{
				"SmsSid":    []string{"abc123"},
				"Body":      []string{"hello"},
				"NumMedia":  []string{"0"},
				"MediaUrl0": []string{"http://example.com/0"},
				"MediaUrl1": []string{"http://example.com/1"},
			},
			false,
			"abc123",
			"hello",
			nil,
		},
		{
			&url.Values{
				"SmsSid":   []string{"abc123"},
				"Body":     []string{"hello"},
				"NumMedia": []string{"-1"},
			},
			false,
			"abc123",
			"hello",
			nil,
		},
		{
			&url.Values{
				"Body": []string{"hello"},
			},
			true,
			"",
			"",
			nil,
		},
		{
			&url.Values{
				"SmsSid": []string{"abc123"},
			},
			false,
			"abc123",
			"",
			nil,
		},
		{
			&url.Values{
				"SmsSid":   []string{"abc123"},
				"Body":     []string{"hello"},
				"NumMedia": []string{"one"},
			},
			true,
			"",
			"",
			nil,
		},
		{
			&url.Values{
				"SmsSid":   []string{"abc123"},
				"Body":     []string{"hello"},
				"NumMedia": []string{"1"},
			},
			true,
			"",
			"",
			nil,
		},
		{
			&url.Values{
				"SmsSid":    []string{"abc123"},
				"Body":      []string{"hello"},
				"NumMedia":  []string{"2"},
				"MediaUrl0": []string{"http://foo.bar"},
			},
			true,
			"",
			"",
			nil,
		},
	}

	for _, testcase := range testcases {
		post, err := ParsePost(testcase.form)
		if testcase.shouldErr {
			assert.Error(t, err)
			assert.Nil(t, post)
			continue
		}

		assert.Nil(t, err)
		assert.Equal(t, testcase.id, post.ID)
		assert.Equal(t, testcase.body, post.Body)
		_, e := time.Parse(time.RFC3339, post.CreatedAt)
		assert.Nil(t, e, "expected post timestamp to be in ISO format")
		assert.Equal(t, testcase.mediaURLs, post.MediaURLs)
	}
}
