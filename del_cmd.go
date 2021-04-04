//
// Delete a feed from our feed-list.
//

package main

import (
	"fmt"

	"github.com/skx/rss2email/configfile"
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

	// Get the configuration-file
	conf := configfile.New()

	// Upgrade it if necessary
	conf.Upgrade()

	_, err := conf.Parse()
	if err != nil {
		fmt.Printf("Error parsing file: %s\n", err.Error())
		return 1
	}

	// For each argument remove it from the list, if present.
	for _, entry := range args {
		conf.Delete(entry)
	}

	// Save the list.
	err = conf.Save()
	if err != nil {
		fmt.Printf("failed to save the updated feed list: %s\n", err.Error())
		return 1
	}

	// All done, with no errors.
	return 0

}
