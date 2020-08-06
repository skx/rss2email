package withstate

import (
	"os"
	"testing"

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

	// Mark as seen
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
