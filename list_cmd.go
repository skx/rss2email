//
// List our configured-feeds.
//

package main

import (
	"fmt"

	"github.com/skx/rss2email/configfile"
	"github.com/skx/subcommands"
)

// Structure for our options and state.
type listCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info is part of the subcommand-API
func (l *listCmd) Info() (string, string) {
	return "list", `Output the list of feeds which are being polled.

This subcommand lists the feeds which are specified in the
configuration file.

To see details of the configuration file, including the location,
please run:

   $ rss2email help config


Example:

    $ rss2email list
`
}

//
// Entry-point.
//
func (l *listCmd) Execute(args []string) int {

	// Get the configuration-file
	conf := configfile.New()

	// Upgrade it if necessary
	conf.Upgrade()

	// Now do the parsing
	entries, err := conf.Parse()
	if err != nil {
		fmt.Printf("Error with config-file: %s\n", err.Error())
		return 1
	}

	// Show the feeds
	for _, entry := range entries {
		fmt.Printf("%s\n", entry.URL)
	}

	return 0
}
