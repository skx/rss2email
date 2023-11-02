// Package configfile contains the logic to read a list of source
// URLs, along with any (optional) configuration-directives.
//
// A configuration file looks like this:
//
//	https://example.com/
//	 - foo:bar
//
//	https://example.org/
//	https://example.net/
//	# comment
//
// It is assumed lines contain URLs, but anything prefixed with a "-"
// is taken to be a parameter using a colon-deliminator.
package configfile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/skx/rss2email/state"
)

// Option contain options which are used on a per-feed basis.
//
// We could use a map, but that would mean that each named option could
// only be used once - and we want to allow multiple "exclude" values
// for example.
type Option struct {
	// Name holds the name of the configuration option.
	Name string

	// Value contains the specified value of the configuration option.
	Value string
}

// Feed is an entry which is read from our configuration-file.
//
// A feed consists of an URL pointing to an Atom/RSS feed, as well as
// an optional set of parameters which are specific to that feed.
type Feed struct {
	// URL is the URL of an Atom/RSS feed.
	URL string

	// Options contains a collection of any optional parameters
	// which have been read after an URL
	Options []Option
}

// ConfigFile contains our state.
type ConfigFile struct {

	// Path contains the path to our config file
	path string

	// The entries we found.
	entries []Feed

	// Key:value regular expression
	re *regexp.Regexp
}

// New creates a new configuration-file reader.
func New() *ConfigFile {
	return &ConfigFile{
		re: regexp.MustCompile(`^([^:]+):(.*)$`),
	}
}

// NewWithPath creates a configuration-file reader, using the given file as
// a source.  This is primarily used for testing.
func NewWithPath(file string) *ConfigFile {

	// Create new object - to avoid having to repeat our regexp
	// initialization.
	x := New()

	// Setup the path, and return the updated object.
	x.path = file
	return x
}

// Path returns the path to the configuration-file.
func (c *ConfigFile) Path() string {

	// If we've not calculated the path then do so now.
	if c.path == "" {
		c.path = filepath.Join(state.Directory(), "feeds.txt")
	}

	return c.path
}

// Parse returns the entries from the config-file
func (c *ConfigFile) Parse() ([]Feed, error) {

	// Remove all existing entries
	c.entries = []Feed{}

	// Open the file
	file, err := os.Open(c.Path())
	if err != nil {
		return c.entries, err
	}
	defer file.Close()

	// Temporary entry
	var tmp Feed
	tmp.Options = []Option{}

	// Create a scanner to process the file.
	scanner := bufio.NewScanner(file)

	// Scan line by line
	for scanner.Scan() {

		// Get the line, and strip leading/trailing space
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// optional params have "-" prefix
		if strings.HasPrefix(line, "-") {

			// options go AFTER the URL to which they refer
			if tmp.URL == "" {
				return c.entries, fmt.Errorf("error: option outside a URL: %s", scanner.Text())
			}

			// Remove the prefix and split by ":"
			line = strings.TrimPrefix(line, "-")

			// Look for "foo:bar"
			fields := c.re.FindStringSubmatch(line)

			// If we got key/val then save them away
			if len(fields) == 3 {
				key := strings.TrimSpace(fields[1])
				val := strings.TrimSpace(fields[2])
				tmp.Options = append(tmp.Options, Option{Name: key, Value: val})
			} else {
				// If we have an URL show it, to help identify the section which is broken
				if tmp.URL != "" {
					return c.entries, fmt.Errorf("options should be of the form 'key:value', bogus entry found '%s', beneath feed %s", line, tmp.URL)
				}
				return c.entries, fmt.Errorf("options should be of the form 'key:value', bogus entry found '%s'", line)

			}
		} else {

			// If we already have a URL stored then append
			// it and reset our temporary structure
			if tmp.URL != "" {
				// store it, and reset our map
				c.entries = append(c.entries, tmp)
				tmp.Options = []Option{}
			}

			// set the url
			tmp.URL = line
		}
	}

	// Ensure we don't forget about the last item in the file.
	if tmp.URL != "" {
		c.entries = append(c.entries, tmp)
	}

	// Look for scanner-errors
	if err := scanner.Err(); err != nil {
		return c.entries, err
	}

	return c.entries, nil
}

// Add appends the given URIs to the config-file
//
// You must call `Save` if you wish this removal to be persisted.
func (c *ConfigFile) Add(uris ...string) {

	for _, uri := range uris {

		// Look to see if it is already-present.
		found := false
		for _, ent := range c.entries {
			if ent.URL == uri {
				found = true
			}
		}

		// Not found?  Then we can add it.
		if !found {
			f := Feed{URL: uri}
			c.entries = append(c.entries, f)
		}
	}
}

// Delete removes an entry from our list of feeds.
//
// You must call `Save` if you wish this removal to be persisted.
func (c *ConfigFile) Delete(url string) {

	var keep []Feed

	for _, ent := range c.entries {
		if ent.URL != url {
			keep = append(keep, ent)
		}
	}

	c.entries = keep
}

// Save persists our list of feeds/options to disk.
func (c *ConfigFile) Save() error {

	// Open the file
	file, err := os.Create(c.Path())
	if err != nil {
		return err
	}

	// For each entry do the necessary
	for _, entry := range c.entries {

		fmt.Fprintf(file, "%s\n", entry.URL)

		for _, opt := range entry.Options {
			fmt.Fprintf(file, " - %s:%s\n", opt.Name, opt.Value)
		}

	}

	err = file.Close()
	return err
}
