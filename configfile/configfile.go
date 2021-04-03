// Package configfile contains the logic to read a list of source
// URLs, along with any (optional) configuration-directives.
//
// A configuration file looks like this:
//
//       https://example.com/
//        - foo:bar
//
//       https://example.org/
//       https://example.net/
//       # comment
//
// It is assumed lines contain URLs, but anything prefixed with a "-"
// is taken to be a parameter using a colon-deliminator.
//
package configfile

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Feed is an entry which is read from our configuration-file.
//
// A feed consists of an URL pointing to an Atom/RSS feed, as well as
// an optional set of parameters which are specific to that feed.
type Feed struct {
	// URL is the URL of an Atom/RSS feed.
	URL string

	// Options contains a collection of any optional parameters
	// which have been read after an URL
	Options map[string]string
}

// ConfigFile contains our state.
type ConfigFile struct {

	// Path contains the path to our config file
	path string

	// The entries we found.
	entries []Feed
}

// New creates a new configuration-file reader.
func New() *ConfigFile {
	return &ConfigFile{}
}

// Path returns the path to the configuration-file.
func (c *ConfigFile) Path() string {

	// If we've not calculated the path then do so now.
	if c.path == "" {

		// Default to using $HOME for our storage
		home := os.Getenv("HOME")

		// If that fails then get the current user, and use
		// their home if possible.
		if home == "" {
			usr, err := user.Current()
			if err == nil {
				home = usr.HomeDir
			}
		}

		// Now build up our file-path
		c.path = filepath.Join(home, ".rss2email", "feeds.lst")

	}
	return c.path
}

// Exists returns true if the configuration-file exists.
//
// This is useful as this configuration file was introduced in the 2.x
// release, previously we used a different configuration file, with
// a different format and name.
func (c *ConfigFile) Exists() bool {

	_, err := os.Stat(c.Path())
	if os.IsNotExist(err) {
		return false
	}
	return true
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
	tmp.Options = make(map[string]string)

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
				return c.entries, fmt.Errorf("Error: Option outside a URL: %s", scanner.Text())
			}

			// Remove the prefix and split by ":"
			line = strings.TrimPrefix(line, "-")
			vals := strings.Split(line, ":")

			// If we got key/val then save them awya
			if len(vals) > 1 {
				key := strings.TrimSpace(vals[0])
				val := strings.TrimSpace(vals[1])
				tmp.Options[key] = val
			}
		} else {

			// If we already have a URL stored then append
			// it and reset our temporary structure
			if tmp.URL != "" {
				// store it, and reset our map
				c.entries = append(c.entries, tmp)
				tmp.Options = make(map[string]string)
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

		for key, val := range entry.Options {
			fmt.Fprintf(file, " - %s:%s\n", key, val)
		}

	}

	err = file.Close()
	return err
}
