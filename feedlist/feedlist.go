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

	// entries contains our feed URLs.
	//
	// We use a map to ensure that feed-items are unique
	entries map[string]bool
}

// New returns a new instance of the feedlist.
//
// The existing feed-list will be read, if present, to populate the list of
// feeds.
func New(filename string) *FeedList {

	// Create the object
	m := new(FeedList)

	// Create our map
	m.entries = make(map[string]bool)

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
				m.entries[tmp] = true
			}
		}
	}

	return m
}

// Entries returns the configured feeds.
func (f *FeedList) Entries() []string {

	results := make([]string, len(f.entries))

	i := 0
	for k := range f.entries {
		results[i] = k
		i++
	}

	return (results)
}

// Add adds a new entry to the feed-list.
// You must call `Save` if you wish this addition to be persisted.
func (f *FeedList) Add(uri string) {
	f.entries[uri] = true
}

// Delete removes an entry from our list of feeds.
// You must call `Save` if you wish this removal to be persisted.
func (f *FeedList) Delete(uri string) {
	delete(f.entries, uri)
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
	for i := range f.entries {
		w.WriteString(i + "\n")
	}

	// Close
	w.Flush()
	fh.Close()

	return nil
}
