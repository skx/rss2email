package withstate

import (
	"os"
	"testing"

	"github.com/mmcdole/gofeed"
)

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

// Helper
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
