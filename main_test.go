// main_test.go - just setup a logger so the test-cases have
// one available.

package main

import (
	"log/slog"
	"os"
)

// init runs at test-time.
func init() {

	// setup logging-level
	lvl := &slog.LevelVar{}
	lvl.Set(slog.LevelWarn)

	// create a handler
	opts := &slog.HandlerOptions{Level: lvl}
	handler := slog.NewTextHandler(os.Stderr, opts)

	// ensure the global-variable is set.
	logger = slog.New(handler)
}
