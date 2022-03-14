// Package state exists to provide a simple way of referring
// to the directory beneath which we store state:
//
// 1. The location of the configuration-file.
//
// 2. The location of the BoltDB database.
package state

import (
	"os"
	"os/user"
	"path/filepath"
)

// Directory returns the path to a directory which can be used
// for storing state.
//
// NOTE: This directory might not necessarily exist, we're just
// returning the prefix directory that should/would be used for
// persistent files.
func Directory() string {

	// Default to using $HOME
	home := os.Getenv("HOME")

	if home == "" {
		// Get the current user, and use their home if possible.
		usr, err := user.Current()
		if err == nil {
			home = usr.HomeDir
		}
	}

	// Return the path
	return filepath.Join(home, ".rss2email")
}
