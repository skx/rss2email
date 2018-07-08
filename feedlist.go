//
// This file contains functions relating to our feeds.
//

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// FeedList is the list of our feeds.
type FeedList struct {
	entries []string
}

// New returns a new instance of the feedlist
func NewFeed() *FeedList {
	m := new(FeedList)

	//
	// Open our input-file
	//
	path := os.Getenv("HOME") + "/.rss2email/feeds"
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
	}

	return m
}

// Entries returns the currently configured entries
func (f *FeedList) Entries() []string {
	return (f.entries)
}

// Add adds a new entry
func (f *FeedList) Add(uri string) {
	f.entries = append(f.entries, uri)
}

// Delete removes an entry from our list
func (f *FeedList) Delete(uri string) {

	var tmp []string

	for _, i := range f.entries {
		if i != uri {
			tmp = append(tmp, i)
		}
	}

	f.entries = tmp
}

// Save saves our entries to disc
func (f *FeedList) Save() {

	// Ensure we have a directory.
	dir := os.Getenv("HOME") + "/.rss2email/"
	_ = os.Mkdir(dir, os.ModePerm)

	// Open the file
	file, err := os.Create(dir + "feeds")
	if err != nil {
		fmt.Printf("Error writing to %s%s - %s\n", dir, "feeds", err.Error())
		os.Exit(1)
	}

	w := bufio.NewWriter(file)

	// For each entry
	for _, i := range f.entries {

		w.WriteString(i + "\n")
	}

	w.Flush()
	file.Close()
}
