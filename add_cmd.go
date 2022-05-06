//
// Add a new feed to the users' list of configured feeds.
//

package main

import (
	"flag"
	"fmt"

	"github.com/skx/rss2email/configfile"
)

// Structure for our options and state.
type addCmd struct {

	// Configuration file, used for testing
	config *configfile.ConfigFile
}

// Arguments handles argument-flags we might have.
//
// In our case we use this as a hook to setup our configuration-file,
// which allows testing.
func (a *addCmd) Arguments(flags *flag.FlagSet) {
	a.config = configfile.New()
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

	_, err := a.config.Parse()
	if err != nil {
		fmt.Printf("Error parsing file: %s\n", err.Error())
		return 1
	}

	// For each argument add it to the list
	for _, entry := range args {

		// Add the entry
		a.config.Add(entry)
	}

	// Save the list.
	err = a.config.Save()
	if err != nil {
		fmt.Printf("failed to save the updated feed list: %s\n", err.Error())
		return 1
	}

	// All done, with no errors.
	return 0
}
