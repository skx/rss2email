//
// Add a new feed to our list
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
type addCmd struct {
}

//
// Glue
//
func (*addCmd) Name() string     { return "add" }
func (*addCmd) Synopsis() string { return "Add a new feed to our feed-list." }
func (*addCmd) Usage() string {
	return `add :
  Add the specified URL to our feed-list
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

	list := NewFeed()

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
