// This file contains code which allows us to see if a given
// RSS-feed item has been seen before.

package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/mmcdole/gofeed"
)

// item2Path is used to return a (unique) filename for a specific feed
// item.
// We assume it is possible to determine whether a feed-item has been
// seen before via the presence of this file.
func item2Path(item *gofeed.Item) string {

	// Default to using $HOME
	home := os.Getenv("HOME")

	// Get the current user, and use their home if possible.
	usr, err := user.Current()
	if err == nil {
		home = usr.HomeDir
	}

	// Hash the item GUID
	hasher := sha1.New()
	hasher.Write([]byte(item.GUID))
	hashBytes := hasher.Sum(nil)

	// Hexadecimal conversion
	hexSha1 := hex.EncodeToString(hashBytes)

	// Finally join the path
	out := path.Join(home, ".rss2email", "seen", hexSha1)
	return out

}

// HasSeen will return true if we've previously notified the user about
// this feed-entry.
func HasSeen(item *gofeed.Item) bool {

	file := item2Path(item)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

// RecordSeen will mark the given feed-item as having been processed/seen
// in the past.
func RecordSeen(item *gofeed.Item) {

	// Get the file-path
	file := item2Path(item)

	// Ensure the parent directory exists
	dir, _ := path.Split(file)
	os.MkdirAll(dir, os.ModePerm)

	// We'll write out the link to the item in the file
	d1 := []byte(item.Link)

	// Write it out
	_ = ioutil.WriteFile(file, d1, 0644)
}
