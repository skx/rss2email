//
// Show help about our configuration file.
//

package main

import (
	"flag"
	"fmt"

	"github.com/skx/rss2email/configfile"
)

// Structure for our options and state.
type configCmd struct {

	// Configuration file, used for testing
	config *configfile.ConfigFile
}

// Arguments handles argument-flags we might have.
//
// In our case we use this as a hook to setup our configuration-file,
// which allows testing.
func (c *configCmd) Arguments(flags *flag.FlagSet) {
	c.config = configfile.New()
}

// Info is part of the subcommand-API
func (c *configCmd) Info() (string, string) {

	// Get some details of the (new) configuration file.
	if c.config == nil {
		c.config = configfile.New()
	}
	path := c.config.Path()

	name := "config"
	doc := `Provide documentation for our configuration file.

About
-----

RSS2Email is a simple command-line utility which will fetch items from
remote Atom and RSS feeds and generate emails.

In order to operate it needs a list of remote Atom/RSS feeds to
process, which are stored in a configuration file.


Configuration File Location
---------------------------

As of release 3.x of rss2email the configuration file will be loaded from
the following location:

     ` + path

	doc += `

Configuration File Format
-------------------------

The format of the configuration file is plain-text, and at its simplest
it consists of nothing more than a series of URLs, one per line, like so:

       https://blog.steve.fi/index.rss
       http://floooh.github.io/feed.xml

Entries can be commented out via the '#' character, temporarily:

       https://blog.steve.fi/index.rss
       # http://floooh.github.io/feed.xml

In addition to containing a list of feed-locations the configuration file
allows per-feed configuration options to be set.  The general form of this
support looks like this:

       https://foo.example.com/
        - key:value
        - key:value2
       https://foo.example.com/
        - key2:value2

Here you see that lines prefixed with " -" will be used to specify a key
and value separated with a ":" character.  Configuration-options apply to
the URL above their appearance.

The first example demonstrates that configuration-keys may be repeated multiple
times, if you desire.

As configuration-items refer to feeds it is a fatal error for such a thing
to appear before a URL.

Per-Feed Configuration Options
------------------------------

Key           | Purpose
--------------+--------------------------------------------------------------
delay         | The amount of time to sleep between retried HTTP-fetches.
exclude       | Exclude any item which matches the given regular-expression.
exclude-title | Exclude any item with title matching the given regular-expression.
exclude-older | Exclude any items whose publication date is older than the 
              | specified number of days.
include       | Include only items which match the given regular-expression.
include-title | Include only items with title matching the given regular-expression.
notify        | Comma-delimited list of emails to send notifications to (if set,
              | replaces the emails set in the cron/daemon command).
retry         | The maximum number of times to retry a failing HTTP-fetch.
sleep         | Sleep the specified number of seconds, before making the request.
tag           | Setup a tag for this feed, which can be accessed in the template.
template      | The path to a feed-specific email template to use.
user-agent    | Configure a specific User-Agent when making HTTP requests.


Regular Expression Tips
-----------------------

Regular expressions are case-sensitive by default, to make them ignore any
differences in case prefix them with "(?i)".

For example the following entry will ignore any feed-items containing the
word "cake" in their titles regardless of whether it is written as "cake",
"Cake", or some other combination of upper and lower-cased letters:

      https://example.com/feed/path/here
       - exclude-title: (?i)cake

`
	return name, doc
}

// Execute is invoked if the user specifies `add` as the subcommand.
func (c *configCmd) Execute(args []string) int {

	_, help := c.Info()
	fmt.Fprintf(out, "%s", help)

	// All done, with no errors.
	return 0
}
