// Package feedlist is a trivial wrapper for maintaining a list
// of RSS feeds in a file.
package feedlist

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
)

// FeedList is the list of our feeds.
type FeedList struct {

	// filename is the name of the state-file we use
	filename string

	// entries contains an array of feed URLS.
	entries []string
}

// New returns a new instance of the feedlist.
//
// The existing feed-list will be read, if present, to populate the list of
// feeds.
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
		filename = path.Join(home, ".rss2email", "feeds")
	}

	// Save our updated filename
	m.filename = filename

	// Open our input-file
	file, err := os.Open(filename)
	if err == nil {
		defer file.Close()

		//
		// Process it line by line.
		//
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			tmp := scanner.Text()
			tmp = strings.TrimSpace(tmp)

			//
			// Skip lines that begin with a comment.
			//
			if (tmp != "") && (!strings.HasPrefix(tmp, "#")) {
				m.entries = append(m.entries, tmp)
			}
		}
	}

	return m
}

// Entries returns the configured feeds.
func (f *FeedList) Entries() []string {
	return (f.entries)
}

// Add adds new entries to the feed-list, avoiding duplicates.
// You must call `Save` if you wish this addition to be persisted.
func (f *FeedList) Add(uris ...string) {

	// Maintain a map of seen entries to avoid duplicates
	seen := make(map[string]bool)

	for _, entry := range f.entries {
		seen[entry] = true
	}

	for _, uri := range uris {
		if !seen[uri] {
			f.entries = append(f.entries, uri)
		}

		seen[uri] = true
	}
}

// Delete removes an entry from our list of feeds.
// You must call `Save` if you wish this removal to be persisted.
func (f *FeedList) Delete(uri string) {

	var tmp []string

	for _, i := range f.entries {
		if i != uri {
			tmp = append(tmp, i)
		}
	}

	f.entries = tmp
}

// Save syncs our entries to disc.
func (f *FeedList) Save() error {

	// Of course we need to make sure the directory exists before
	// we can write beneath it.
	dir, _ := path.Split(f.filename)
	os.MkdirAll(dir, os.ModePerm)

	// Open the file
	fh, err := os.Create(f.filename)
	if err != nil {
		return fmt.Errorf("error writing to %s - %s", f.filename, err.Error())
	}

	// Write out each entry
	w := bufio.NewWriter(fh)
	for _, i := range f.entries {
		w.WriteString(i + "\n")
	}

	// Close
	w.Flush()
	fh.Close()

	return nil
}
