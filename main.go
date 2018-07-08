//
// Entry-point for our application.
//

package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
)

//
// Setup our sub-commands and use them.
//
func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&listCmd{}, "")
	subcommands.Register(&addCmd{}, "")
	subcommands.Register(&delCmd{}, "")
	subcommands.Register(&cronCmd{}, "")
	subcommands.Register(&versionCmd{}, "")

	//
	// If we have no arguments then we default to running
	// the cron-action.
	//
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "cron")
	}

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))

}
