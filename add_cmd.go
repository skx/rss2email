//
// Add a new feed to the users' list of configured feeds.
//

package main

import (
	"fmt"

	"github.com/skx/rss2email/feedlist"
	"github.com/skx/subcommands"
)

// Structure for our options and state.
type addCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info is part of the subcommand-API
func (a *addCmd) Info() (string, string) {
	return "add", `Add a new feed to our feed-list.

Add one or more specified URLs to the configuration file.

To see details of the configuration file, including the location,
please run:

   $ rss2email help config

Example:

    $ rss2email add https://blog.steve.fi/index.rss
`
}

// Execute is invoked if the user specifies `add` as the subcommand.
func (a *addCmd) Execute(args []string) int {

	// Get the feed-list, from the default location.
	list := feedlist.New("")

	// For each argument add it to the list
	for _, entry := range args {

		// Add the entry
		errors := list.Add(entry)

		// Errors?
		for _, err := range errors {
			fmt.Printf("%s\n", (err.Error()))
		}
	}

	// Save the list.
	err := list.Save()
	if err != nil {
		fmt.Printf("failed to save the updated feed list: %s\n", err.Error())
		return 1
	}

	// All done, with no errors.
	return 0
}
