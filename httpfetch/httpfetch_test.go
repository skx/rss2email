package httpfetch

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/skx/rss2email/configfile"
	"github.com/skx/rss2email/withstate"
)

var (
	// logger contains a shared logging handle, the code we're testing assumes it exists.
	logger *slog.Logger
)

// init runs at test-time.
func init() {

	// setup logging-level
	lvl := &slog.LevelVar{}
	lvl.Set(slog.LevelWarn)

	// create a handler
	opts := &slog.HandlerOptions{Level: lvl}
	handler := slog.NewTextHandler(os.Stderr, opts)

	// ensure the global-variable is set.
	logger = slog.New(handler)
}

// TestNonFeed confirms we can cope with a remote URL which is not a feed.
func TestNonFeed(t *testing.T) {

	// Not a feed.
	x := New(configfile.Feed{URL: "http://example.com/"}, logger)
	x.content = "this is not an XML file, so not a feed"

	// Parse it, which should fail.
	_, err := x.Fetch()
	if err == nil {
		t.Fatalf("We expected error, but got none!")
	}

	// And confirm it fails in the correct way.
	if !strings.Contains(err.Error(), "Failed to detect feed type") {
		t.Fatalf("got an error, but not what we expected; %s", err.Error())
	}
}

// TestOneEntry confirms a feed contains a single entry
func TestOneEntry(t *testing.T) {

	// The contents of our feed.
	x := New(configfile.Feed{URL: "https://blog.steve.fi/index.rss"}, logger)
	x.content = `<?xml version="1.0"?>
<rdf:RDF
 xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
 xmlns:dc="http://purl.org/dc/elements/1.1/"
 xmlns:foaf="http://xmlns.com/foaf/0.1/"
 xmlns:content="http://purl.org/rss/1.0/modules/content/"
 xmlns="http://purl.org/rss/1.0/"
>
<channel rdf:about="https://blog.steve.fi/">
<title>Steve Kemp&#39;s Blog</title>
<link>https://blog.steve.fi/</link>
<description>Debian and Free Software</description>
<items>
 <rdf:Seq>
  <rdf:li rdf:resource="https://blog.steve.fi/brexit_has_come.html"/>
 </rdf:Seq>
</items>
</channel>

<item rdf:about="https://blog.steve.fi/brexit_has_come.html">
  <title>Brexit has come</title>
  <link>https://blog.steve.fi/brexit_has_come.html</link>
  <guid>https://blog.steve.fi/brexit_has_come.html</guid>
  <content:encoded>Hello, World</content:encoded>
  <dc:date>2020-05-22T09:00:00Z</dc:date>
</item>
</rdf:RDF>
`

	// Parse it which should not fail.
	out, err := x.Fetch()
	if err != nil {
		t.Fatalf("We didn't expect an error, but found %s", err.Error())
	}

	// Confirm there is a single entry.
	if len(out.Items) != 1 {
		t.Fatalf("Expected one entry, but got %d", len(out.Items))
	}
}

// TestRewrite ensures that a broken file is rewriting
func TestRewrite(t *testing.T) {

	// The contents of our feed.
	x := New(configfile.Feed{URL: "https://blog.steve.fi/index.rss"}, logger)
	x.content = `<?xml version="1.0"?>
<rdf:RDF
 xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
 xmlns:dc="http://purl.org/dc/elements/1.1/"
 xmlns:foaf="http://xmlns.com/foaf/0.1/"
 xmlns:content="http://purl.org/rss/1.0/modules/content/"
 xmlns="http://purl.org/rss/1.0/"
>
<channel rdf:about="https://blog.steve.fi/">
<title>Steve Kemp&#39;s Blog</title>
<link>https://blog.steve.fi/</link>
<description>Debian and Free Software</description>
<items>
 <rdf:Seq>
  <rdf:li rdf:resource="https://blog.steve.fi/brexit_has_come.html"/>
 </rdf:Seq>
</items>
</channel>

<item rdf:about="https://blog.steve.fi/brexit_has_come.html">
  <title>Brexit has come</title>
  <link>https://blog.steve.fi/brexit_has_come.html</link>
  <guid>https://blog.steve.fi/brexit_has_come.html</guid>
  <content:encoded>&lt;a href="/foo"&gt;Foo&lt;/a&gt;</content:encoded>
  <dc:date>2020-05-22T09:00:00Z</dc:date>
</item>
</rdf:RDF>
`

	// Parse it which should not fail.
	out, err := x.Fetch()
	if err != nil {
		t.Fatalf("We didn't expect an error, but found %s", err.Error())
	}

	// Confirm there is a single entry.
	if len(out.Items) != 1 {
		t.Fatalf("Expected one entry, but got %d", len(out.Items))
	}

	// Get the parsed-content
	item := withstate.FeedItem{Item: out.Items[0]}
	content, err := item.HTMLContent()
	if err != nil {
		t.Fatalf("unexpected error on item content: %v", err)
	}

	// Confirm that contains : href="https://blog.steve.fi/foo",
	// not: href="/foo"
	if strings.Contains(content, "\"/foo") {
		t.Fatalf("Failed to expand URLS: %s", content)
	}
}

func TestDelay(t *testing.T) {

	// Valid number
	n := New(configfile.Feed{URL: "https://blog.steve.fi/index.rss",
		Options: []configfile.Option{
			{Name: "delay", Value: "15"},
		}}, logger)

	if n.retryDelay != 15*time.Millisecond {
		t.Errorf("failed to parse delay value")
	}

	// Invalid number
	i := New(configfile.Feed{URL: "https://blog.steve.fi/index.rss",
		Options: []configfile.Option{
			{Name: "delay", Value: "steve"},
		}}, logger)

	if i.retryDelay != 1000*time.Millisecond {
		t.Errorf("bogus value changed our delay-value")
	}
}

func TestRetry(t *testing.T) {

	// Valid number
	n := New(configfile.Feed{URL: "https://blog.steve.fi/index.rss",
		Options: []configfile.Option{
			{Name: "retry", Value: "33"},
			{Name: "moi", Value: "3"},
		}}, logger)

	if n.maxRetries != 33 {
		t.Errorf("failed to parse retry value")
	}

	// Invalid number
	i := New(configfile.Feed{URL: "https://blog.steve.fi/index.rss",
		Options: []configfile.Option{
			{Name: "retry", Value: "steve"},
		}}, logger)

	if i.maxRetries != 3 {
		t.Errorf("bogus value changed our default")
	}
}

// Make a HTTP-request against a local entry
func TestHTTPFetch(t *testing.T) {

	// Setup a stub server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	// Create a config-entry which points to the fake HTTP-server
	conf := configfile.Feed{URL: ts.URL}

	// Create a fetcher
	obj := New(conf, logger)

	// Now make the HTTP-fetch
	_, err := obj.Fetch()

	if err == nil {
		t.Fatalf("expected an error from the fetch")
	}
	if !strings.Contains(err.Error(), "Failed to detect feed type") {
		t.Fatalf("got an error, but the wrong kind")
	}
}

// Make a HTTP-request against a local entry
func TestHTTPFetchValid(t *testing.T) {

	feed := `<?xml version="1.0"?>
<rdf:RDF
 xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
 xmlns:dc="http://purl.org/dc/elements/1.1/"
 xmlns:foaf="http://xmlns.com/foaf/0.1/"
 xmlns:content="http://purl.org/rss/1.0/modules/content/"
 xmlns="http://purl.org/rss/1.0/"
>
<channel rdf:about="https://blog.steve.fi/">
<title>Steve Kemp&#39;s Blog</title>
<link>https://blog.steve.fi/</link>
<description>Debian and Free Software</description>
<items>
 <rdf:Seq>
  <rdf:li rdf:resource="https://blog.steve.fi/brexit_has_come.html"/>
 </rdf:Seq>
</items>
</channel>

<item rdf:about="https://blog.steve.fi/brexit_has_come.html">
  <title>Brexit has come</title>
  <link>https://blog.steve.fi/brexit_has_come.html</link>
  <guid>https://blog.steve.fi/brexit_has_come.html</guid>
  <content:encoded>Hello, World</content:encoded>
  <dc:date>2020-05-22T09:00:00Z</dc:date>
</item>
</rdf:RDF>
`
	// User-agent setup
	agent := "foo:bar:baz"

	// Setup a stub server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, feed)
	}))
	defer ts.Close()

	// Create a config-entry which points to the fake HTTP-server
	conf := configfile.Feed{URL: ts.URL,
		Options: []configfile.Option{
			{Name: "user-agent", Value: agent},
		},
	}

	// Create a fetcher
	obj := New(conf, logger)

	if obj.userAgent != agent {
		t.Fatalf("failed to setup user-agent")
	}

	// Now make the HTTP-fetch
	res, err := obj.Fetch()

	if err != nil {
		t.Fatalf("unexpected error fetching feed")
	}

	if len(res.Items) != 1 {
		t.Fatalf("wrong feed count")
	}
}
