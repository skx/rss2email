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
