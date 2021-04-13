package processor

import (
	"testing"

	"github.com/skx/rss2email/configfile"
)

func TestSendEmail(t *testing.T) {

	p := New()

	if !p.send {
		t.Fatalf("unexpected default to sending mail")
	}

	p.SetSendEmail(false)

	if p.send {
		t.Fatalf("unexpected send-setting")
	}

}

func TestVerbose(t *testing.T) {

	p := New()

	if p.verbose {
		t.Fatalf("unexpected default to verbose")
	}

	p.SetVerbose(true)

	if !p.verbose {
		t.Fatalf("unexpected verbose setting")
	}
}

// TestSkipExclude ensures that we can exclude items by regexp
func TestSkipExclude(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			configfile.Option{Name: "exclude", Value: "foo"},
			configfile.Option{Name: "exclude-title", Value: "test"},
		},
	}

	// Create the new processor
	x := New()

	// Set it as verbose
	x.SetVerbose(true)

	if !x.shouldSkip(feed, "Title here", "<p>foo, bar baz</p>") {
		t.Fatalf("failed to skip entry by regexp")
	}

	if !x.shouldSkip(feed, "test", "<p>This matches the title</p>") {
		t.Fatalf("failed to skip entry by title")
	}

	// With no options we're not going to skip
	feed = configfile.Feed{
		URL:     "blah",
		Options: []configfile.Option{},
	}

	if x.shouldSkip(feed, "Title here", "<p>foo, bar baz</p>") {
		t.Fatalf("skipped something with no options!")
	}

}

// TestSkipInclude ensures that we can exclude items by regexp
func TestSkipInclude(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			configfile.Option{Name: "include", Value: "good"},
		},
	}

	// Create the new processor
	x := New()

	// Set it as verbose
	x.SetVerbose(true)

	if x.shouldSkip(feed, "Title here", "<p>This is good</p>") {
		t.Fatalf("this should be included because it contains good")
	}

	if !x.shouldSkip(feed, "Title here", "<p>This should be excluded.</p>") {
		t.Fatalf("This should be excluded; doesn't contain 'good'")
	}

	// If we don't try to make a mandatory include setting
	// nothing should be skipped
	feed = configfile.Feed{
		URL:     "blah",
		Options: []configfile.Option{},
	}

	if x.shouldSkip(feed, "Title here", "<p>This is good</p>") {
		t.Fatalf("nothing specified, shouldn't be skipped")
	}
}

// TestSkipIncludeTitle ensures that we can exclude items by regexp
func TestSkipIncludeTitle(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			configfile.Option{Name: "include", Value: "good"},
			configfile.Option{Name: "include-title", Value: "(?i)cake"},
		},
	}

	// Create the new processor
	x := New()

	// Set it as verbose
	x.SetVerbose(true)

	if x.shouldSkip(feed, "Title here", "<p>This is good</p>") {
		t.Fatalf("this should be included because it contains good")
	}
	if x.shouldSkip(feed, "I like Cake!", "<p>Food is good.</p>") {
		t.Fatalf("this should be included because of the title")
	}

	//
	// Second test, only include titles
	//
	feed = configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			configfile.Option{Name: "include-title", Value: "(?i)cake"},
			configfile.Option{Name: "include-title", Value: "(?i)pie"},
		},
	}

	//
	// Some titles which are OK
	//
	valid := []string{"I like cake", "I like pie", "piecemeal", "cupcake", "pancake"}
	bogus := []string{"I do not like food", "I don't like cooked goods", "cheese is dead milk", "books are fun", "tv is good"}

	// Create the new processor
	x = New()

	// Set it as verbose
	x.SetVerbose(true)

	// include
	for _, entry := range valid {
		if x.shouldSkip(feed, entry, "content") {
			t.Fatalf("this should be included due to include-title")
		}
	}

	// exclude
	for _, entry := range bogus {
		if !x.shouldSkip(feed, entry, "content") {
			t.Fatalf("this shouldn't be included!")
		}
	}
}
