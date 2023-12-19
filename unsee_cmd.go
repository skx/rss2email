//
// "Unsee" a feed item
//

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/skx/rss2email/state"
	"go.etcd.io/bbolt"
)

// Structure for our options and state.
type unseeCmd struct {

	// Are our arguments regular expressions?
	regexp bool
}

// Info is part of the subcommand-API.
func (u *unseeCmd) Info() (string, string) {
	return "unsee", `Regard a feed item as new, and unseen.

This sub-command will allow you to mark an item as
unseen, or new, meaning the next time the cron or daemon
commands run they'll trigger an email notification.

You can see the URLs which we regard as having already seen
via the 'seen' sub-command.
`
}

// Arguments handles our flag-setup.
func (u *unseeCmd) Arguments(f *flag.FlagSet) {
	f.BoolVar(&u.regexp, "regexp", false, "Are our arguments regular expressions, instead of literal URLs?")
}

// Entry-point.
func (u *unseeCmd) Execute(args []string) int {

	if len(args) < 1 {
		fmt.Printf("Please specify the URLs to unsee (i.e. remove from the state database).\n")
		return 1
	}

	// Ensure we have a state-directory.
	dir := state.Directory()
	errM := os.MkdirAll(dir, 0666)
	if errM != nil {
		logger.Error("failed to create directory", slog.String("directory", dir), slog.String("error", errM.Error()))
		return 1
	}

	// Now create the database, if missing, or open it if it exists.
	dbPath := filepath.Join(dir, "state.db")
	db, err := bbolt.Open(dbPath, 0666, nil)
	if err != nil {
		logger.Error("failed to open database", slog.String("database", dbPath), slog.String("error", err.Error()))
		return 1
	}

	// Ensure we close when we're done
	defer db.Close()

	// Keep track of buckets here
	var bucketNames []string

	// Record each bucket
	err = db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(bucketName []byte, _ *bbolt.Bucket) error {
			bucketNames = append(bucketNames, string(bucketName))
			return nil
		})
	})

	if err != nil {
		logger.Error("failed to find bucket names", slog.String("database", dbPath), slog.String("error", err.Error()))
		return 1
	}

	// Process each bucket to find the item to remove.
	for _, buck := range bucketNames {

		err = db.Update(func(tx *bbolt.Tx) error {

			// Items to remove
			remove := []string{}

			// Select the bucket, which we know must exist
			b := tx.Bucket([]byte(buck))

			// Get a cursor to the key=value entries in the bucket
			c := b.Cursor()

			// Iterate over the key/value pairs.
			for k, _ := c.First(); k != nil; k, _ = c.Next() {

				// Convert the key to a string
				key := string(k)

				// Is this something to remove?
				for _, arg := range args {

					// If so append it.
					if u.regexp {
						match, _ := regexp.MatchString(arg, key)
						if match {
							remove = append(remove, key)
						}
					} else {

						// Literal string-match
						if arg == key {
							remove = append(remove, key)

							logger.Debug("removed item from history", slog.String("item", key), slog.String("database", dbPath), slog.String("bucket", buck))
						}
					}
				}
			}

			// Now remove
			for _, key := range remove {
				err = b.Delete([]byte(key))
				if err != nil {
					logger.Debug("failed to remove item from history", slog.String("item", key), slog.String("database", dbPath), slog.String("bucket", buck), slog.String("error", err.Error()))

				}
			}
			return nil
		})

		if err != nil {
			logger.Error("failed iterating over bucket", slog.String("database", dbPath), slog.String("bucket", buck), slog.String("error", err.Error()))
			return 1
		}
	}

	return 0
}
