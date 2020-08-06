//
// Delete a feed from our feed-list.
//

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
	"github.com/skx/rss2email/feedlist"
)

//
// The options set by our command-line flags.
//
type delCmd struct {
}

//
// Glue
//
func (*delCmd) Name() string     { return "delete" }
func (*delCmd) Synopsis() string { return "Remove a feed from our list." }
func (*delCmd) Usage() string {
	return `Remove the specified URLs from the feed list.

Example:

    $ rss2email delete https://blog.steve.fi/index.rss
`
}

//
// Flag setup: NOP
//
func (p *delCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *delCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// Create the helper
	//
	list := feedlist.New()

	//
	// Count the entries
	//
	before := len(list.Entries())

	//
	// For each argument remove it from the list, if present.
	//
	for _, entry := range f.Args() {
		list.Delete(entry)
	}

	//
	// If we made a change then save it.
	//
	if len(list.Entries()) != before {
		list.Save()
	} else {
		fmt.Printf("Feed list unchanged.\nUse 'rss2email list' to check your current feed configuration.\n")
	}

	//
	// All done.
	//
	return subcommands.ExitSuccess
}
