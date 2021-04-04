//
// Delete a feed from our feed-list.
//

package main

import (
	"fmt"

	"github.com/skx/rss2email/feedlist"
	"github.com/skx/subcommands"
)

// Structure for our options and state.
type delCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info is part of the subcommand-API
func (d *delCmd) Info() (string, string) {
	return "delete", `Remove a feed from our feed-list.

Remove one or more specified URLs from the configuration file.

To see details of the configuration file, including the location,
please run:

   $ rss2email help config

Example:

    $ rss2email delete https://blog.steve.fi/index.rss
`
}

//
// Entry-point.
//
func (d *delCmd) Execute(args []string) int {

	// Get the feed-list, from the default location.
	list := feedlist.New("")

	// Count the entries, so we can determine whether we
	// removed any entries.
	before := len(list.Entries())

	// For each argument remove it from the list, if present.
	for _, entry := range args {
		list.Delete(entry)
	}

	// If we made a change then save it.
	if len(list.Entries()) != before {
		list.Save()
	} else {
		fmt.Printf("Feed list unchanged.\n")
		fmt.Printf("Use 'rss2email list' to check your current feed list.\n")
	}

	// All done.
	return 0
}
