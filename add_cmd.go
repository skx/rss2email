//
// Add a new feed to our list
//

package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
	"github.com/skx/rss2email/feedlist"
)

//
// The options set by our command-line flags.
//
type addCmd struct {
}

//
// Glue
//
func (*addCmd) Name() string     { return "add" }
func (*addCmd) Synopsis() string { return "Add a new feed to our feed-list." }
func (*addCmd) Usage() string {
	return `Add one or more specified URLs to our feed-list.

Example:

    $ rss2email add https://blog.steve.fi/index.rss
`
}

//
// Flag setup: NOP
//
func (p *addCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *addCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// Get the feed-list
	//
	list := feedlist.New("")

	//
	// For each argument add it to the list
	//
	for _, entry := range f.Args() {
		list.Add(entry)
	}

	list.Save()

	//
	// All done.
	//
	return subcommands.ExitSuccess
}
