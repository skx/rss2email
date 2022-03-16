//
// Entry-point for our application.
//

package main

import (
	"fmt"
	"os"

	"github.com/skx/subcommands"
)

//
// Recovery is good
//
func recoverPanic() {
	if r := recover(); r != nil {
		fmt.Printf("recovered from panic while running %v\n%s\n", os.Args, r)
	}
}

//
// Register the subcommands, and run the one the user chose.
//
func main() {

	//
	// Catch errors
	//
	defer recoverPanic()

	//
	// Register each of our subcommands.
	//
	subcommands.Register(&addCmd{})
	subcommands.Register(&cronCmd{})
	subcommands.Register(&configCmd{})
	subcommands.Register(&daemonCmd{})
	subcommands.Register(&delCmd{})
	subcommands.Register(&exportCmd{})
	subcommands.Register(&importCmd{})
	subcommands.Register(&listCmd{})
	subcommands.Register(&listDefaultTemplateCmd{})
	subcommands.Register(&seenCmd{})
	subcommands.Register(&versionCmd{})

	//
	// Execute the one the user chose.
	//
	os.Exit(subcommands.Execute())
}
