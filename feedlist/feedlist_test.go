package feedlist

import (
	"io/ioutil"
	"os"
	"testing"
)

// TestEmpty ensures we can handle a missing file
func TestEmpty(t *testing.T) {

	obj := New("/path/does/not/exist")

	entries := obj.Entries()
	if len(entries) != 0 {
		t.Fatalf("Found error reading a missing file")
	}
}

// TestSave ensures we create a file with an entry in it.
func TestSave(t *testing.T) {

	// Create a temporary file
	file, err := ioutil.TempFile(os.TempDir(), "testsave")
	if err != nil {
		t.Fatalf("failed to make temporary file: %s", err.Error())
	}
	defer os.Remove(file.Name())

	// Create a new feed
	list := New(file.Name())
	list.Add("https://example.com/foo.atom")

	// Save it to disk
	err = list.Save()
	if err != nil {
		t.Fatalf("failed to save feed list: %s", err)
	}

	// The file should now have contents - we can reload it
	// and confirm
	updated := New(file.Name())
	found := updated.Entries()

	if len(found) != 1 {
		t.Errorf("expected one entry, found %d", len(found))
	}
	if found[0] != "https://example.com/foo.atom" {
		t.Errorf("unexpected entry found: %s", found[0])
	}
}

// TestDelete ensures we can delete an entry from a feedlist
func TestDelete(t *testing.T) {
	// Create a temporary file
	file, err := ioutil.TempFile(os.TempDir(), "testsave")
	if err != nil {
		t.Fatalf("failed to make temporary file: %s", err.Error())
	}
	defer os.Remove(file.Name())

	// Create a new feed with two entries
	list := New(file.Name())
	list.Add("https://example.com/foo.atom")
	list.Add("https://example.com/foo.xml")

	// Save it to disk
	err = list.Save()
	if err != nil {
		t.Fatalf("failed to save feed list: %s", err)
	}

	// The file should now have contents - we can reload it
	// and confirm the count is as-expected
	updated := New(file.Name())
	found := updated.Entries()

	if len(found) != 2 {
		t.Errorf("expected two entries, found %d", len(found))
	}

	// Now delete the atom entry & save
	updated.Delete("https://example.com/foo.atom")
	err = updated.Save()
	if err != nil {
		t.Fatalf("failed to save feed list: %s", err)
	}

	// If the delete worked we have a feedlist with a single
	// entry present now.
	final := New(file.Name())
	if len(final.Entries()) != 1 {
		t.Errorf("expected one entry, found %d", len(found))
	}

	if final.Entries()[0] != "https://example.com/foo.xml" {
		t.Errorf("after deletion we have an unexpected entry")
	}
}
