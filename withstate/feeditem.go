// Package withstate provides a simple wrapper of the gofeed.Item, which
// allows simple tracking of the seen vs. unseen (new vs. old) state of
// an RSS feeds' entry.
//
// State for a feed-item is stored upon the local filesystem.
package withstate

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/mmcdole/gofeed"
)

// FeedItem is a structure wrapping a gofeed.Item, to allow us to record
// state.
type FeedItem struct {

	// Wrapped structure
	*gofeed.Item

	// Local state here.
	// TODO: Allow the prefix to be specified.
}

// IsNew reports whether this particular feed-item new.
func (item *FeedItem) IsNew() bool {

	file := item.path()
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return true
	}
	return false
}

// RecordSeen updates this item, to record the entry has haven been seen.
func (item *FeedItem) RecordSeen() {

	// Get the file-path
	file := item.path()

	// Ensure the parent directory exists
	dir, _ := path.Split(file)
	os.MkdirAll(dir, os.ModePerm)

	// We'll write out the link to the item in the file
	d1 := []byte(item.Link)

	// Write it out
	_ = ioutil.WriteFile(file, d1, 0644)
}

// path returns an appropriate marker-file, which is used to record
// the seen vs. unseen state of a particular entry.
func (item *FeedItem) path() string {

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
