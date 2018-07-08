//
// Delete a feed from our feed-list.
//

package main

import (
	"context"
	"flag"

	"github.com/google/subcommands"
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
	return `delete :
  This command updates our configured feed-list to remove the specified entry.
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
	list := NewFeed()

	//
	// Count the entries
	//
	before := len(list.Entries())

	//
	// For each argument add it to the list
	//
	for _, entry := range f.Args() {
		list.Delete(entry)
	}

	//
	// If we made a change then save it
	//
	if len(list.Entries()) != before {
		list.Save()
	}

	//
	// All done.
	//
	return subcommands.ExitSuccess
}
