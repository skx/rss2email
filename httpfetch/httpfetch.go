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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/skx/rss2email/configfile"
)

var (
	// cache contains the values we can use to be cache-friendly.
	cache map[string]CacheHelper

	// ErrUnchanged is returned by our HTTP-fetcher if the content was previously
	// fetched and has not changed since then.
	ErrUnchanged = errors.New("UNCHANGED")
)

// CacheHelper is a struct used to store modification-data relating to the
// URL we're fetching.
//
// We use this to make conditional HTTP-requests, rather than fetching the
// feed from scratch each time.
type CacheHelper struct {

	// Etag contains the Etag the server sent, if any.
	Etag string

	// LastModified contains the Last-Modified header the server sent, if any.
	LastModified string

	// Updated contains the timestamp of when the feed was last fetched (successfully).
	Updated time.Time
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

	// frequency controls the poll frequency.  If this is set to 1hr then
	// we don't fetch the feed until 1 hour after the last fetch, even if
	// we were executed as a daemon with a SLEEP setting of 5 (minutes).
	frequency time.Duration

	// The User-Agent header to send when making our HTTP fetch
	userAgent string

	// logger contains the logging handle to use, if any
	logger *slog.Logger
}

// New creates a new object which will fetch our content.
func New(entry configfile.Feed, log *slog.Logger, version string) *HTTPFetch {

	// Create object with defaults
	state := &HTTPFetch{url: entry.URL,
		maxRetries: 3,
		retryDelay: 5 * time.Second,
		userAgent:  fmt.Sprintf("rss2email %s (https://github.com/skx/rss2email)", version),
	}

	// Get the user's sleep period - if overridden this will become the
	// default frequency for each feed item.
	sleep := os.Getenv("SLEEP")
	if sleep == "" {
		state.frequency = 15 * time.Minute
	} else {
		v, err := strconv.Atoi(sleep)
		if err == nil {
			state.frequency = time.Duration(v) * time.Minute
		}
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

		// Sleep-delay between failed fetch-attempts.
		if opt.Name == "delay" {

			num, err := strconv.Atoi(opt.Value)
			if err == nil {
				state.retryDelay = time.Duration(num) * time.Second
			}
		}

		// User-Agent
		if opt.Name == "user-agent" {
			state.userAgent = opt.Value
		}

		// Polling frequency
		if opt.Name == "frequency" {
			num, err := strconv.Atoi(opt.Value)
			if err == nil {
				state.frequency = time.Duration(num) * time.Minute
			}
		}
	}

	// Create a local logger with some dedicated information
	state.logger = log.With(
		slog.Group("httpfetch",
			slog.String("link", entry.URL),
			slog.String("user-agent", state.userAgent),
			slog.Bool("insecure", state.insecure),
			slog.Int("retry-max", state.maxRetries),
			slog.Duration("retry-delay", state.retryDelay),
			slog.Duration("frequency", state.frequency)))

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

		// no error? that means we're good and we've retrieved
		// the content.
		if err == nil {
			break
		}

		// The remote content hasn't changed?
		if err == ErrUnchanged {
			return nil, ErrUnchanged
		}

		// if we got here we have to retry, but we should
		// show the error too.
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

	// Do we have a cache-entry?
	prevCache, okCache := cache[h.url]
	if okCache {
		h.logger.Debug("we have cached headers saved from a previous request",
			slog.String("etag", prevCache.Etag),
			slog.String("last-modified", prevCache.LastModified))
	}

	// Create a HTTP-client
	client := &http.Client{}

	// Setup a transport which disables TLS-checks
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// If we're ignoring the TLS then use a non-validating transport.
	if h.insecure {
		client.Transport = tr
	}

	// We only support making HTTP GET requests.
	req, err := http.NewRequest("GET", h.url, nil)
	if err != nil {
		return err
	}

	// If we've previously fetched this URL set the appropriate
	// cache-related headers in our new request.
	if okCache {

		// If there is a frequency for this feed AND the time has not yet
		// been reached then we terminate early.
		if time.Since(prevCache.Updated) < h.frequency {
			h.logger.Debug("avoiding this fetch, the feed was retrieved already within the frequency limit",
				slog.Time("last", prevCache.Updated),
				slog.Duration("duration", h.frequency))
			return ErrUnchanged
		}

		// Otherwise set the cache-related headers.

		if prevCache.Etag != "" {
			h.logger.Debug("setting HTTP-header on outgoing request",
				slog.String("url", h.url),
				slog.String("If-None-Match", prevCache.Etag))

			req.Header.Set("If-None-Match", prevCache.Etag)
		}
		if prevCache.LastModified != "" {
			h.logger.Debug("setting HTTP-header on outgoing request",
				slog.String("url", h.url),
				slog.String("If-Modified-Since", prevCache.LastModified))
			req.Header.Set("If-Modified-Since", prevCache.LastModified)
		}

	}

	// Populate the HTTP User-Agent header - some sites (e.g. reddit) fail without this.
	req.Header.Set("User-Agent", h.userAgent)

	// Make the actual HTTP request.
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the response headers and save any cache-like things
	// we can use to avoid excessive load in the future.
	x := CacheHelper{
		Etag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		Updated:      time.Now(),
	}
	cache[h.url] = x

	//
	// Did the remote page not change?
	//
	status := resp.StatusCode
	if status >= 300 && status < 400 {
		h.logger.Debug("response from request was unchanged",
			slog.String("status", resp.Status),
			slog.Int("code", resp.StatusCode))
		return ErrUnchanged
	}

	// Otherwise we save the result away and
	// return any error/not as a result of reading
	// the body.
	data, err2 := io.ReadAll(resp.Body)
	h.content = string(data)

	h.logger.Debug("response from request",
		slog.String("url", h.url),
		slog.String("status", resp.Status),
		slog.Int("code", resp.StatusCode),
		slog.Int("size", len(h.content)))

	return err2
}
