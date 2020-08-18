// Package withstate provides a simple wrapper of the gofeed.Item, which
// allows simple tracking of the seen vs. unseen (new vs. old) state of
// an RSS feeds' entry.
//
// State for a feed-item is stored upon the local filesystem.
package withstate

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/mmcdole/gofeed"
)

// FeedItem is a structure wrapping a gofeed.Item, to allow us to record
// state.
type FeedItem struct {

	// Wrapped structure
	*gofeed.Item
}

// IsNew reports whether this particular feed-item is new.
func (item *FeedItem) IsNew() bool {

	file := item.path()
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return true
	}
	return false
}

// RecordSeen updates this item, to record the fact that it has been seen.
func (item *FeedItem) RecordSeen() {

	// Get the file-path
	file := item.path()

	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t := time.Now()
		_ = os.Chtimes(file, t, t)
		return
	}

	// Ensure the parent directory exists
	dir, _ := path.Split(file)
	os.MkdirAll(dir, os.ModePerm)

	// We'll write out the link to the item in the file
	d1 := []byte(item.Link)

	// Write it out
	_ = ioutil.WriteFile(file, d1, 0644)
}

// stateDirectory returns the directory beneath which we store state
func stateDirectory() string {

	// Default to using $HOME
	home := os.Getenv("HOME")

	if home == "" {
		// Get the current user, and use their home if possible.
		usr, err := user.Current()
		if err == nil {
			home = usr.HomeDir
		}
	}

	// Return with a subdirectory
	return path.Join(home, ".rss2email", "seen")
}

// path returns an appropriate marker-file, which is used to record
// the seen vs. unseen state of a particular entry.
func (item *FeedItem) path() string {

	guid := item.GUID
	if guid == "" {
		guid = item.Link
	}

	// Hash the item GUID and convert to hexadecimal
	hexSha1 := fmt.Sprintf("%x", sha1.Sum([]byte(guid)))

	// Finally join the path
	out := path.Join(stateDirectory(), hexSha1)
	return out

}

// isSha1File returns true if a regular file has a name that looks
// like a sha1.  This is an incomplete check, but may prevent a
// non-state file from being removed.
func isSha1File(fi os.FileInfo) bool {

	name := fi.Name()

	if len(name) != 40 {
		return false
	}

	for _, r := range name {
		if r >= '0' && r <= '9' {
			continue
		}
		if r >= 'a' && r <= 'f' {
			continue
		}
		return false
	}

	return fi.Mode().IsRegular()
}

// PruneStateFiles removes no-longer-needed state files
// It returns the number of files pruned and a slice of errors encountered.
func PruneStateFiles() (int, []error) {

	stateDirPath := stateDirectory()

	stateDir, err := os.Open(stateDirectory())
	if err != nil {
		err = fmt.Errorf("failed to open state-file directory: %s", err.Error())
		return 0, []error{err}
	}

	fileInfos, err := stateDir.Readdir(0)
	if err != nil {
		err = fmt.Errorf("failed to list state files: %s", err.Error())
		return 0, []error{err}
	}

	errors := make([]error, 0)
	prunedCount := 0

	// Prune state files older than 4 days.
	for _, fi := range fileInfos {
		if time.Since(fi.ModTime()) > (4*24)*time.Hour {
			if !isSha1File(fi) {
				continue
			}

			err := os.Remove(path.Join(stateDirPath, fi.Name()))
			if err == nil {
				prunedCount++
			} else {
				err = fmt.Errorf("failed to remove state file: %s", err.Error())
				errors = append(errors, err)
			}
		}
	}

	return prunedCount, errors
}
