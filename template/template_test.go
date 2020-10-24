//
// Simple testing of our embedded resource.
//

package template

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//
// Test that we have one embedded resource.
//
func TestResourceCount(t *testing.T) {

	expected := 0

	// We're going to compare what is embedded with
	// what is on-disk.
	//
	// We could just hard-wire the count, but that
	// would require updating the count every time
	// we add/remove a new resource
	err := filepath.Walk("../data",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				expected++
			}
			return nil
		})
	if err != nil {
		t.Errorf("failed to find resources beneath data/ %s", err.Error())
	}

	if expected != 1 {
		t.Fatalf("we expected 1 template-file beneath data/")
	}

	out := getResources()

	if len(out) != expected {
		t.Errorf("We expected %d resources but found %d.", expected, len(out))
	}
}

//
// Test that each of our resources is identical to the master
// version.
//
func TestResourceMatches(t *testing.T) {

	//
	// For each resource
	//
	all := getResources()

	for _, entry := range all {

		//
		// Get the data from our embedded copy
		//
		data, err := getResource(entry.Filename)
		if err != nil {
			t.Errorf("Loading our resource failed:%s", entry.Filename)
		}

		//
		// Get the data from our master-copy.
		//
		master, err := ioutil.ReadFile("../" + entry.Filename)
		if err != nil {
			t.Errorf("Loading our master-resource failed:%s", entry.Filename)
		}

		//
		// Test the lengths match
		//
		if len(master) != len(data) {
			t.Errorf("Embedded and real resources have different sizes.")
		}

		//
		// Test the data-matches
		//
		if string(master) != string(data) {
			t.Errorf("Embedded and real resources have different content.")
		}
	}
}

//
// Test that a missing resource is handled.
//
func TestMissingResource(t *testing.T) {

	//
	// Get the data from our embedded copy
	//
	data, err := getResource("moi/kissa")
	if data != nil {
		t.Errorf("We expected to find no data, but got some.")
	}
	if err == nil {
		t.Errorf("We expected an error loading a missing resource, but got none.")
	}
	if !strings.Contains(err.Error(), "failed to find resource") {
		t.Errorf("Error message differed from expectations.")
	}
}
