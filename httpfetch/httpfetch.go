// Package httpfetch makes a remote HTTP call to retrieve an URL,
// and parse that into a series of feed-items
//
// It is abstracted into its own class to allow testing.
package httpfetch

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/skx/rss2email/configfile"
)

var (
	// cache contains the values we can use to be cache-friendly
	cache map[string]CacheHelper

	// ErrUnchanged is returned by our HTTP-fetcher if the content was previously
	// fetched and has not changed.
	ErrUnchanged = errors.New("UNCHANGED")
)

// CacheHelper is a struct used to store modification-data relating to the
// URL we're fetching.
type CacheHelper struct {

	// Etag contains the Etag the server sent, if any.
	Etag string

	// LastModified contains the Last-Modified header the server sent, if any.
	LastModified string
}

// init is called once at startup, and creates the cache-map we use to avoid
// making too many HTTP-probes against remote URLs (i.e. feeds)
func init() {
	cache = make(map[string]CacheHelper)
}

// HTTPFetch is our state-storing structure
type HTTPFetch struct {

	// The URL we should fetch
	url string

	// Contents of the remote URL, used for testing
	content string

	// How many times we should attempt to retry a failed
	// fetch before giving up.
	maxRetries int

	// insecure will cause invalid SSL certificate options to be ignored
	insecure bool

	// Between retries we should delay to avoid overwhelming
	// the remote server.  This specifies how many times we should
	// do that.
	retryDelay time.Duration

	// The User-Agent header to send when making our HTTP fetch
	userAgent string

	// logger contains the logging handle to use, if any
	logger *slog.Logger
}

// New creates a new object which will fetch our content.
func New(entry configfile.Feed, log *slog.Logger) *HTTPFetch {

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

		// Disable fatal TLS errors.  Horrid
		if opt.Name == "insecure" {

			// downcase the value
			val := strings.ToLower(opt.Value)

			// if it is enabled then set the flag
			if val == "yes" || val == "true" {
				state.insecure = true
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

	// Create a local logger with some dedicated information
	state.logger = log.With(
		slog.Group("httpfetch",
			slog.String("link", entry.URL),
			slog.String("user-agent", state.userAgent),
			slog.Bool("insecure", state.insecure),
			slog.Int("retry-max", state.maxRetries),
			slog.Int("retry-delay", int(state.retryDelay/time.Millisecond)/1000)))

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

		// Log the fetch attempt
		h.logger.Debug("fetching URL",
			slog.Int("attempt", i+1))

		// fetch the contents
		err = h.fetch()

		// no error? that means we're good and we've downloaded
		if err == nil {
			break
		}

		// we've got a page that matches a previous fetch?
		// We'll return that then
		if err == ErrUnchanged {
			return nil, ErrUnchanged
		}

		// if we got here we have to retry, but we should
		// show the error too
		h.logger.Debug("fetching URL failed",
			slog.String("error", err.Error()))

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

		h.logger.Warn("failed to parse content",
			slog.String("error", err2.Error()))

		return nil, fmt.Errorf("error parsing %s contents: %s", h.url, err2.Error())
	}

	return feed, nil
}

// fetch fetches the text from the remote URL.
func (h *HTTPFetch) fetch() error {

	// Do we have a cache-entry
	prevCache, okCache := cache[h.url]
	if okCache {
		fmt.Printf("We have a cached value from a previous fetch!\n")
	}

	// Create a HTTP-client
	client := &http.Client{}

	// Setup a transport which disables TLS-checks
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// If we're ignoring the TLS
	if h.insecure {

		// Use the non-validating transport
		client.Transport = tr
	}

	req, err := http.NewRequest("GET", h.url, nil)
	if err != nil {
		return err
	}

	// If we've previously fetched this URL populate the caching
	// values in the request.
	if okCache {
		fmt.Printf("Setting cache headers for our outgoing request\n")

		if prevCache.Etag != "" {
			req.Header.Set("If-None-Match", prevCache.Etag)
		}
		if prevCache.LastModified != "" {
			req.Header.Set("If-Modified-Since", prevCache.LastModified)
		}
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

	//
	// Read the reponse headers and save any cache-like things
	// we can use to avoid excessive load in the future.
	//
	x := CacheHelper{Etag: resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("LastModified"),
	}

	fmt.Printf("Storing in cache: %v\n", x)
	cache[h.url] = x

	//
	// Did the remote page not change?
	//
	status := resp.StatusCode
	if status >= 300 && status < 400 {
		fmt.Printf("Unchanged!\n")
		return ErrUnchanged
	}

	fmt.Printf("HTTP fetch returned %s:%d\n", resp.Status, resp.StatusCode)

	// Otherwise we save the result away and
	// return any error/not as a result of reading
	// the body.
	data, err2 := io.ReadAll(resp.Body)
	h.content = string(data)
	return err2
}
