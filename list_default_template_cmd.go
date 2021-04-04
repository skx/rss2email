//
// List our default email-template
//

package main

import (
	"fmt"
	"os"

	"github.com/skx/rss2email/template"
	"github.com/skx/subcommands"
)

// listDefaultTemplateCmd holds our state.
type listDefaultTemplateCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info is part of the subcommand-API
func (l *listDefaultTemplateCmd) Info() (string, string) {
	return "list-default-template", `Display the default email-template.

An embedded template is used to format the emails which are sent by this
application, when new feed items are discovered.  If you wish to change
the way the emails are formed, or formatted, you can replace this template
with a local copy.

To replace the template which is used simple create a new file located at
'~/.rss2email/email.tmpl', with your content.

This sub-command can be used to give you a starting point for your edits:

   $ rss2email list-default-template > ~/.rss2email/email.tmpl


Example:

    $ rss2email list-default-template
`
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
