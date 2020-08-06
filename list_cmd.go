//
// List our configured-feeds.
//

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
)

//
// The options set by our command-line flags.
//
type listCmd struct {
}

//
// Glue
//
func (*listCmd) Name() string     { return "list" }
func (*listCmd) Synopsis() string { return "List configured feeds." }
func (*listCmd) Usage() string {
	return `Output the list of feeds which are being polled.

Example:

    $ rss2email list
`
}

//
// Flag setup: NOP
//
func (p *listCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *listCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// Create the helper
	//
	list := NewFeed()

	//
	// For each entry in the list ..
	//
	for _, uri := range list.Entries() {

		//
		// Print it
		//
		fmt.Printf("%s\n", uri)
	}
	return subcommands.ExitSuccess
}
