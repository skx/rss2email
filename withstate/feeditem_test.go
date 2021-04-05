package withstate

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

// TestBasics ensures we can change the seen state of an entry
func TestBasics(t *testing.T) {

	// Create an item
	x := &FeedItem{&gofeed.Item{}}

	// Give it an identity
	x.GUID = "steve-test"

	// Should be New.
	if !x.IsNew() {
		t.Errorf("With no marker we've got an item that is seen")
	}

	// Mark as seen, twice.
	//
	// The second time is designed to make sure that we handle
	// the time-changing.
	x.RecordSeen()
	x.RecordSeen()

	// Shouldn't be new any longer.
	if x.IsNew() {
		t.Errorf("Item was regarded as new, even though we marked it as read")
	}

	// Now cleanup
	os.Remove(x.path())
}

// TestCollision ensures that different objects hash the same way
func TestCollision(t *testing.T) {

	// So we want to have two feed items with the same
	// GUID.  They should map to the same file, so we
	// can confirm they would be treated as identical
	a := &FeedItem{&gofeed.Item{}}
	b := &FeedItem{&gofeed.Item{}}

	a.GUID = "steve"
	b.GUID = "steve"

	if a.path() != b.path() {
		t.Fatalf("two identical objects have different hashes/paths")
	}

	// Update to confirm that results in a change
	b.GUID = "kemp"

	if a.path() == b.path() {
		t.Fatalf("two different objects have identical hashes/paths")
	}
}

// TestCollisionMissingHome ensures that we can find the home-directory
// of a user, even without the environment
func TestCollisionMissingHome(t *testing.T) {

	// Remove the environmental variable
	cur := os.Getenv("HOME")
	os.Setenv("HOME", "")
	statePrefix = ""

	// So we want to have two feed items with the same
	// GUID.  They should map to the same file, so we
	// can confirm they would be treated as identical
	a := &FeedItem{&gofeed.Item{}}
	b := &FeedItem{&gofeed.Item{}}

	a.GUID = "steve"
	b.GUID = "steve"

	if a.path() != b.path() {
		t.Fatalf("two identical objects have different hashes/paths")
	}

	// Update to confirm that results in a change
	b.GUID = "kemp"

	if a.path() == b.path() {
		t.Fatalf("two different objects have identical hashes/paths")
	}

	// Reset
	os.Setenv("HOME", cur)
}

// TestPrune creates some files and ensures that those that are "old"
// are pruned.
func TestPrune(t *testing.T) {

	// Test case
	type Entry struct {

		// Name of the file we'll create.
		name string

		// old is true if we should give the file an old time.
		//
		// i.e. A time sufficiently far in the past that we'd
		// expect the entry to be pruned.
		old bool

		// Whether we expect this file to remain, post-prune.
		remain bool
	}

	tests := []Entry{

		// These will remain - not be pruned - because
		// their names are not SHA1 hashes
		{"foo", true, true},
		{"bar", false, true},

		// Note "X" in name
		{"9cX5770b3bb4b2a1d59be2d97e34379cd192299f", true, true},

		// We expect this to be reaped ("steve")
		//
		// The file is "old", and has a suitable name.
		{"9ce5770b3bb4b2a1d59be2d97e34379cd192299f", true, false},

		// But not this ("kemp")
		//
		// This file is "new" so the name doesn't matter.
		{"d2e31a60feabe8a58c828264eb0a75257fbe45ad", false, true},
	}

	// Create a temporary directory
	dir, err := ioutil.TempDir("", "prune")
	if err != nil {
		t.Fatalf("failed to create temporary directory:%s", err)
	}

	// Create each file
	for _, tst := range tests {

		// Create the file beneath the temporary dir
		out := filepath.Join(dir, tst.name)

		// Write bogus content
		err = ioutil.WriteFile(out, []byte(tst.name), 0666)
		if err != nil {
			t.Fatalf("failed to write temporary file : %s", err)
		}

		// If this is to be an "old" file then back-date it 100 hours.
		if tst.old {
			t := time.Now()
			t = t.Add(time.Duration(-100) * time.Hour)
			_ = os.Chtimes(out, t, t)
		}
	}

	//
	// Now we've created a temporary file with some specific
	// files in it.
	//
	// Run the prune
	//
	statePrefix = dir
	PruneStateFiles()

	//
	// For each one - see if we got the results we expect
	//
	for _, tst := range tests {

		// Does it exist?
		out := filepath.Join(dir, tst.name)
		exists := fileExists(out)

		// Does it exist in the way we expect?
		if exists != tst.remain {
			t.Fatalf("%s error exists:%t expected:%t", tst.name, exists, tst.remain)
		}
	}
	// Remove our state
	defer os.RemoveAll(dir) // clean up
}

// Helper
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
