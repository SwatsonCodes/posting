package models

import (
	"net/url"
	"testing"
	"time"
)

func TestDecodePost(t *testing.T) {
	testcases := []struct {
		form      *url.Values
		shouldErr bool
		id        string
		body      string
		mediaURLs *[]string
	}{
		{
			&url.Values{
				"SmsSid": []string{"abc123"},
				"Body":   []string{"hello"},
			},
			false,
			"abc123",
			"hello",
			nil},
		{
			&url.Values{
				"SmsSid":   []string{"abc123"},
				"Body":     []string{"hello"},
				"NumMedia": []string{"0"},
			},
			false,
			"abc123",
			"hello",
			nil},
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
			&[]string{"http://www.example.com"},
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
			&[]string{"http://example.com/0", "http://example.com/1", "http://example.com/2"},
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
			&[]string{"http://example.com/0"},
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
			true,
			"",
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
			if err == nil {
				t.Error("expected error but got nil")
			}
			if post != nil {
				t.Error("expected post to be nil due to error, but it wasnt")
			}
			continue
		}

		if err != nil {
			t.Errorf("got unexpected error: %s", err.Error())
		}

		if post.ID != testcase.id {
			t.Errorf("expected post ID to be '%s' but it was '%s'", testcase.id, post.ID)
		}
		if post.Body != testcase.body {
			t.Errorf("expected post body to be '%s' but it was '%s'", testcase.body, post.Body)
		}
		if _, e := time.Parse(time.RFC3339, post.CreatedAt); e != nil {
			t.Errorf("got error when parsing post CreatedAt timestamp '%s': '%s'", post.CreatedAt, e.Error())
		}

		if testcase.mediaURLs != nil {
			if post.MediaURLs == nil {
				t.Error("expected post to have MediaURLs, but the field is nil")
				continue
			}
			if len(*testcase.mediaURLs) != len(*post.MediaURLs) {
				t.Errorf("expected post to have %d MediaURLs, but it has %d", len(*testcase.mediaURLs), len(*post.MediaURLs))
			}
			for i, mURL := range *testcase.mediaURLs {
				if pmURL := (*post.MediaURLs)[i]; mURL != pmURL {
					t.Errorf("expected MediaURL%d to be %s, but it was %s", i, mURL, pmURL)
				}
			}
		}

	}

}
