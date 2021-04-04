package httpfetch

import (
	"strings"
	"testing"

	"github.com/skx/rss2email/configfile"
	"github.com/skx/rss2email/withstate"
)

// TestNonFeed confirms we can cope with a remote URL which is not a feed.
func TestNonFeed(t *testing.T) {

	// Not a feed.
	x := New(configfile.Feed{URL: "http://example.com/"})
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
	x := New(configfile.Feed{URL: "https://blog.steve.fi/index.rss"})
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
	x := New(configfile.Feed{URL: "https://blog.steve.fi/index.rss"})
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
