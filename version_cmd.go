//
// Show our version.
//

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
)

//
// modified during testing
//
var out io.Writer = os.Stdout

var (
	version = "unreleased"
)

// Structure for our options and state.
type versionCmd struct {
	// verbose controls whether our version information includes
	// the go-version.
	verbose bool
}

// Info is part of the subcommand-API.
func (v *versionCmd) Info() (string, string) {
	return "version", `Report upon our version, and exit.`
}

// Arguments handles our flag-setup.
func (v *versionCmd) Arguments(f *flag.FlagSet) {
	f.BoolVar(&v.verbose, "verbose", false, "Show go version the binary was generated with.")
}

//
// Show the version - using the "out"-writer.
//
func showVersion(verbose bool) {
	fmt.Fprintf(out, "%s\n", version)
	if verbose {
		fmt.Fprintf(out, "Built with %s\n", runtime.Version())
	}
}

//
// Entry-point.
//
func (v *versionCmd) Execute(args []string) int {

	showVersion(v.verbose)

	return 0
}
