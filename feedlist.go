//
// This file contains functions relating to our feeds.
//

package main

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
	// entries contains an array of feed URLS.
	entries []string
}

// NewFeed returns a new instance of the feedlist.
// The existing feed-list will be read, if present, to populate the list of
// feeds.
func NewFeed() *FeedList {
	m := new(FeedList)

	// Default to using $HOME for our storage
	home := os.Getenv("HOME")

	// Get the current user, and use their home if possible.
	usr, err := user.Current()
	if err == nil {
		home = usr.HomeDir
	}

	// Now build up our file-path
	path := path.Join(home, ".rss2email", "feeds")

	// Open our input-file
	file, err := os.Open(path)
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
	} else {
		fmt.Printf("WARNING: %s not found\n", path)
	}

	return m
}

// Entries returns the configured feeds.
func (f *FeedList) Entries() []string {
	return (f.entries)
}

// Add adds a new entry to the feed-list.
// You must call `Save` if you wish this addition to be persisted.
func (f *FeedList) Add(uri string) {
	f.entries = append(f.entries, uri)
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
func (f *FeedList) Save() {

	// Default to using $HOME for our storage
	home := os.Getenv("HOME")

	// Get the current user, and use their home if possible.
	usr, err := user.Current()
	if err == nil {
		home = usr.HomeDir
	}

	// Now build up our file-path
	file := path.Join(home, ".rss2email", "feeds")

	// Of course we need to make sure the directory exists before
	// we can write beneath it.
	dir, _ := path.Split(file)
	os.MkdirAll(dir, os.ModePerm)

	// Open the file
	fh, err := os.Create(file)
	if err != nil {
		fmt.Printf("Error writing to %s%s - %s\n", dir, "feeds", err.Error())
		os.Exit(1)
	}

	w := bufio.NewWriter(fh)

	// For each entry
	for _, i := range f.entries {

		w.WriteString(i + "\n")
	}

	w.Flush()
	fh.Close()
}
