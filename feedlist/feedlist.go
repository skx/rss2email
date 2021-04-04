// Package feedlist is a trivial wrapper for maintaining a list
// of RSS feeds in a file.
//
// NOTE: This is the legacy configuration-file reader. It will be removed
// in future releases.
package feedlist

import (
	"bufio"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// expandedEntry is a url with its comment from the feeds file.
type expandedEntry struct {
	// url is the feed's url
	url string

	// comments contains the blank lines and comments preceding the url
	comments []string
}

// FeedList is the list of our feeds.
type FeedList struct {

	// filename is the name of the state-file we use
	filename string

	// expandedEntries contains an array of feed URLS.
	expandedEntries []expandedEntry
}

// New returns a new instance of the feedlist.
//
// The existing feedlist file will be read, if present, to populate the
// list of feeds.
func New(filename string) *FeedList {

	// Create the object
	m := new(FeedList)

	// If there was no path specified then create something
	// sensible.
	if filename == "" {

		// Default to using $HOME for our storage
		home := os.Getenv("HOME")

		// If that fails then get the current user, and use
		// their home if possible.
		if home == "" {
			usr, err := user.Current()
			if err == nil {
				home = usr.HomeDir
			}
		}

		// Now build up our file-path
		filename = filepath.Join(home, ".rss2email", "feeds")
	}

	// Save our updated filename
	m.filename = filename

	// Open our input-file
	file, err := os.Open(filename)
	if err == nil {
		defer file.Close()

		seenFeed := make(map[string]bool)

		//
		// Process it line by line.
		//
		comments := make([]string, 0)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			tmp := scanner.Text()
			tmp = strings.TrimSpace(tmp)

			//
			// Save non-url lines as comments
			//
			if tmp == "" || strings.HasPrefix(tmp, "#") {
				comments = append(comments, tmp)
				continue
			}

			eEntry := expandedEntry{url: tmp, comments: comments}
			comments = make([]string, 0)

			if !seenFeed[eEntry.url] {
				m.expandedEntries = append(m.expandedEntries, eEntry)
				seenFeed[eEntry.url] = true
			}
		}
	}

	return m
}

// Entries returns the configured feeds.
func (f *FeedList) Entries() []string {
	urls := make([]string, len(f.expandedEntries))
	for i, eEntry := range f.expandedEntries {
		urls[i] = eEntry.url
	}
	return (urls)
}
