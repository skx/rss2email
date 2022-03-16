//
// Show feeds and their contents
//

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/skx/rss2email/state"
	"github.com/skx/subcommands"
	"go.etcd.io/bbolt"
)

// Structure for our options and state.
type seenCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info is part of the subcommand-API.
func (s *seenCmd) Info() (string, string) {
	return "seen", `Show all the feed-items we've seen.

This sub-command will report upon all the feeds to which
you're subscribed, and show the link to each feed-entry
to which you've been notified in the past.

(i.e. This walks the internal database which is used to
store state, and outputs the list of recorded items which
are no longer regarded as new/unseen.)
`
}

//
// Entry-point.
//
func (s *seenCmd) Execute(args []string) int {

	// Ensure we have a state-directory.
	dir := state.Directory()
	os.MkdirAll(dir, 0666)

	// Now create the database, if missing, or open it if it exists.
	db, err := bbolt.Open(filepath.Join(dir, "state.db"), 0666, nil)
	if err != nil {
		fmt.Printf("Error opening database: %s\n", err.Error())
		return 1
	}

	// Ensure we close when we're done
	defer db.Close()

	// Keep track of buckets here
	var bucketNames [][]byte

	db.View(func(tx *bbolt.Tx) error {
		tx.ForEach(func(bucketName []byte, _ *bbolt.Bucket) error {
			bucketNames = append(bucketNames, bucketName)
			return nil
		})
		return nil
	})

	// Now we have a list of buckets, we'll show the contents
	for _, buck := range bucketNames {

		fmt.Printf("%s\n", buck)

		err = db.View(func(tx *bbolt.Tx) error {

			// Select the bucket, which we know must exist
			b := tx.Bucket([]byte(buck))

			// Get a cursor to the key=value entries in the bucket
			c := b.Cursor()

			// Iterate over the key/value pairs.
			for k, _ := c.First(); k != nil; k, _ = c.Next() {

				// Convert the key to a string
				key := string(k)

				fmt.Printf("\t%s\n", key)
			}

			return nil
		})

		if err != nil {
			fmt.Printf("error iterating over bucket %s: %s\n", buck, err.Error())
			return 1
		}
	}

	return 0
}
