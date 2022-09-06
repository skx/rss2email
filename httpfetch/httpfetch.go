// Package httpfetch makes a remote HTTP call to retrieve an URL,
// and parse that into a series of feed-items
//
// It is abstracted into its own class to allow testing.
package httpfetch

import (
	"io"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/skx/rss2email/configfile"
)

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

	// The User-Agent header to send when making our HTTP fetch
	userAgent string
}

// New creates a new object which will fetch our content.
func New(entry configfile.Feed) *HTTPFetch {

	// Create object with defaults
	state := &HTTPFetch{url: entry.URL,
		maxRetries: 3,
		retryDelay: 1000 * time.Millisecond,
		userAgent:  "rss2email (https://github.com/skx/rss2email)",
	}

	// Are any of our options overridden?
	for _, opt := range entry.Options {

		// Max-retry count.
		if opt.Name == "retry" {

			num, err := strconv.Atoi(opt.Value)
			if err == nil {
				state.maxRetries = num
			}
		}

		// Sleep-delay between fetch-attempts.
		if opt.Name == "delay" {

			num, err := strconv.Atoi(opt.Value)
			if err == nil {
				state.retryDelay = time.Duration(num) * time.Millisecond
			}
		}

		// User-Agent
		if opt.Name == "user-agent" {
			state.userAgent = opt.Value
		}
	}

	return state
}

// Fetch performs the HTTP-fetch, and returns the feed-contents.
//
// If our internal `content` field is non-empty it will be used in preference
// to making a remote request, which is useful for testing.
func (h *HTTPFetch) Fetch() (*gofeed.Feed, error) {

	var feed *gofeed.Feed
	var err error

	// Download contents, if not already present.
	for i := 0; h.content == "" && i < h.maxRetries; i++ {

		err = h.fetch()
		if err == nil {
			break
		}
		time.Sleep(h.retryDelay)

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

// fetch fetches the text from the remote URL.
func (h *HTTPFetch) fetch() error {

	// Create a HTTP-client
	client := &http.Client{}
	req, err := http.NewRequest("GET", h.url, nil)
	if err != nil {
		return err
	}

	// Populate the HTTP User-Agent header.
	//
	// Some sites (e.g. reddit) fail without a header set.
	req.Header.Set("User-Agent", h.userAgent)

	// Make the actual HTTP request.
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// save the result
	data, err2 := io.ReadAll(resp.Body)
	h.content = string(data)
	return err2
}
