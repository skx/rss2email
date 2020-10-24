//
// List our default email-template
//

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/skx/rss2email/template"
)

// listDefaultTemplateCmd holds our state.
type listDefaultTemplateCmd struct {
}

// Info is part of the subcommand-API
func (l *listDefaultTemplateCmd) Info() (string, string) {
	return "list-default-template", `Output the default email-template.

This command outputs the default template which is used to generate the
emails which are sent.

If you create a new template located at ~/.rss2email/email.tmpl this will
be used in preference to the default file.  So this sub-command can be used
to give you a starting point for your edits:

   $ rss2email list-default-template > ~/.rss2email/email.tmpl


Example:

    $ rss2email list-default-template


`
}

// Arguments handles our flag-setup.
func (l *listDefaultTemplateCmd) Arguments(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (l *listDefaultTemplateCmd) Execute(args []string) int {

	// Load the default template from the embedded resource.
	content, err := template.EmailTemplate()
	if err != nil {
		fmt.Printf("failed to load embedded resource: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("%s\n", string(content))
	return 0
}
