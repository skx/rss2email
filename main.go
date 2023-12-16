//
// Entry-point for our application.
//

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/skx/subcommands"
)

var (
	// logger contains a shared logging handle, used by our sub-commands.
	logger *slog.Logger
)

// Recovery is good
func recoverPanic() {
	if r := recover(); r != nil {
		logger.Error("recovered from a panic", slog.String("error", fmt.Sprintf("%s", r)))
	}
}

// Register the subcommands, and run the one the user chose.
func main() {

	//
	// Setup our default logging level
	//
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelWarn)

	//
	// Allow showing "all the logs"
	//
	if os.Getenv("LOG_ALL") != "" {
		lvl.Set(slog.LevelDebug)
	}

	// Those handler options
	opts := &slog.HandlerOptions{
		Level: lvl,
	}

	//
	// Default to showing to STDERR in text.
	//
	var handler slog.Handler = slog.NewTextHandler(os.Stderr, opts)

	//
	// But allow JSON formatting too.
	//
	if os.Getenv("LOG_JSON") != "" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	}

	//
	// Create our logging handler, using the level we've just setup
	//
	logger = slog.New(handler)

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
	subcommands.Register(&unseeCmd{})
	subcommands.Register(&versionCmd{})

	//
	// Execute the one the user chose.
	//
	os.Exit(subcommands.Execute())
}
