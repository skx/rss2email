package processor

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/skx/rss2email/configfile"
)

var (
	// logger contains a shared logging handle, the code we're testing assumes it exists.
	logger *slog.Logger
)

// init runs at test-time.
func init() {

	// setup logging-level
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelWarn)

	// create a handler
	opts := &slog.HandlerOptions{Level: lvl}
	handler := slog.NewTextHandler(os.Stderr, opts)

	// ensure the global-variable is set.
	logger = slog.New(handler)
}

func TestSendEmail(t *testing.T) {

	p, err := New()

	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer p.Close()

	if !p.send {
		t.Fatalf("unexpected default to sending mail")
	}

	p.SetSendEmail(false)

	if p.send {
		t.Fatalf("unexpected send-setting")
	}

}

func TestVerbose(t *testing.T) {

	p, err := New()

	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}

	defer p.Close()
}

// TestSkipExclude ensures that we can exclude items by regexp
func TestSkipExclude(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "exclude", Value: "foo"},
			{Name: "exclude-title", Value: "test"},
		},
	}

	// Create the new processor
	x, err := New()

	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	if !x.shouldSkip(logger, feed, "Title here", "<p>foo, bar baz</p>") {
		t.Fatalf("failed to skip entry by regexp")
	}

	if !x.shouldSkip(logger, feed, "test", "<p>This matches the title</p>") {
		t.Fatalf("failed to skip entry by title")
	}

	// With no options we're not going to skip
	feed = configfile.Feed{
		URL:     "blah",
		Options: []configfile.Option{},
	}

	if x.shouldSkip(logger, feed, "Title here", "<p>foo, bar baz</p>") {
		t.Fatalf("skipped something with no options!")
	}

}

// TestSkipInclude ensures that we can exclude items by regexp
func TestSkipInclude(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "include", Value: "good"},
		},
	}

	// Create the new processor
	x, err := New()

	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	if x.shouldSkip(logger, feed, "Title here", "<p>This is good</p>") {
		t.Fatalf("this should be included because it contains good")
	}

	if !x.shouldSkip(logger, feed, "Title here", "<p>This should be excluded.</p>") {
		t.Fatalf("This should be excluded; doesn't contain 'good'")
	}

	// If we don't try to make a mandatory include setting
	// nothing should be skipped
	feed = configfile.Feed{
		URL:     "blah",
		Options: []configfile.Option{},
	}

	if x.shouldSkip(logger, feed, "Title here", "<p>This is good</p>") {
		t.Fatalf("nothing specified, shouldn't be skipped")
	}
}

// TestSkipIncludeTitle ensures that we can exclude items by regexp
func TestSkipIncludeTitle(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "include", Value: "good"},
			{Name: "include-title", Value: "(?i)cake"},
		},
	}

	// Create the new processor
	x, err := New()
	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}

	if x.shouldSkip(logger, feed, "Title here", "<p>This is good</p>") {
		t.Fatalf("this should be included because it contains good")
	}
	if x.shouldSkip(logger, feed, "I like Cake!", "<p>Food is good.</p>") {
		t.Fatalf("this should be included because of the title")
	}

	//
	// Second test, only include titles
	//
	feed = configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "include-title", Value: "(?i)cake"},
			{Name: "include-title", Value: "(?i)pie"},
		},
	}

	//
	// Some titles which are OK
	//
	valid := []string{"I like cake", "I like pie", "piecemeal", "cupcake", "pancake"}
	bogus := []string{"I do not like food", "I don't like cooked goods", "cheese is dead milk", "books are fun", "tv is good"}

	// Create the new processor
	x.Close()
	x, err = New()
	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	// include
	for _, entry := range valid {
		if x.shouldSkip(logger, feed, entry, "content") {
			t.Fatalf("this should be included due to include-title")
		}
	}

	// exclude
	for _, entry := range bogus {
		if !x.shouldSkip(logger, feed, entry, "content") {
			t.Fatalf("this shouldn't be included!")
		}
	}
}

// TestSkipOlder ensures that we can exclude items by age
func TestSkipOlder(t *testing.T) {

	feed := configfile.Feed{
		URL: "blah",
		Options: []configfile.Option{
			{Name: "exclude-older", Value: "1"},
		},
	}

	// Create the new processor
	x, err := New()

	if err != nil {
		t.Fatalf("error creating processor %s", err.Error())
	}
	defer x.Close()

	if x.shouldSkipOlder(logger, feed, "X") {
		t.Fatalf("failed to skip non correct published-date")
	}

	if !x.shouldSkipOlder(logger, feed, "Fri, 02 Dec 2022 16:43:04 +0000") {
		t.Fatalf("failed to skip old entry by age")
	}

	if !x.shouldSkipOlder(logger, feed, time.Now().Add(-time.Hour*24*2).Format(time.RFC1123)) {
		t.Fatalf("failed to skip newer entry by age")
	}

	if x.shouldSkipOlder(logger, feed, time.Now().Add(-time.Hour*12).Format(time.RFC1123)) {
		t.Fatalf("skipped new entry by age")
	}

	// With no options we're not going to skip
	feed = configfile.Feed{
		URL:     "blah",
		Options: []configfile.Option{},
	}

	if x.shouldSkipOlder(logger, feed, time.Now().Add(-time.Hour*24*128).String()) {
		t.Fatalf("skipped age with no options!")
	}
}
