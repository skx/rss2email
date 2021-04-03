// Package httpfetch makes a remote HTTP call to retrieve an URL,
// and parse that into a series of feed-items
//
// It is abstracted into its own class to allow testing.
package httpfetch

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

const ()

// HTTPFetch is our state-storing structure
type HTTPFetch struct {

	// The URL we should fetch
	url string

	// Contents of the remote URL, used for testing
	content string

	// How many times we should attempt to retry a failed
	// fetch before giving up.
	maxRetries int

	// Between retries we should delay to avoid overwhelming
	// the remote server.  This specifies how many times we should
	// do that.
	retryDelay time.Duration
}

// New creates a new object which will fetch our content
func New(url string) *HTTPFetch {
	return &HTTPFetch{url: url,
		maxRetries: 5,
		retryDelay: 200 * time.Millisecond,
	}
}

// Fetch performs the HTTP-fetch, and returns the feed-contents.
//
// If the `content` field is non-empty it will be used in preference
// to the remote URLs content, for testing.
func (h *HTTPFetch) Fetch() (*gofeed.Feed, error) {

	var feed *gofeed.Feed
	var err error

	// Download contents, if not already present.
	for i := 0; h.content == "" && i < h.maxRetries; i++ {

		err = h.fetch()
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i) * h.retryDelay)
	}

	// Failed, after all the retries?
	if err != nil {
		return feed, err
	}

	// Parse it
	fp := gofeed.NewParser()
	feed, err2 := fp.ParseString(h.content)
	if err2 != nil {
		return nil, fmt.Errorf("error parsing %s contents: %s", h.url, err2.Error())
	}

	return feed, nil
}

// fetchURL fetches the text from the remote URL.
func (h *HTTPFetch) fetch() error {

	// Create a HTTP-client
	client := &http.Client{}
	req, err := http.NewRequest("GET", h.url, nil)
	if err != nil {
		return err
	}

	// Make the request, with a valid user-agent
	req.Header.Set("User-Agent", "rss2email (https://github.com/skx/rss2email)")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// save the result
	data, err2 := ioutil.ReadAll(resp.Body)
	h.content = string(data)
	return err2
}
