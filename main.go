//
// Entry-point for our application.
//

package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/skx/subcommands"
)

var (
	// logger contains a shared logging handle, used by our sub-commands.
	logger *slog.Logger

	// loggerLevel allows changing the log-level at runtime
	loggerLevel *slog.LevelVar
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
	loggerLevel = new(slog.LevelVar)
	loggerLevel.Set(slog.LevelWarn)

	//
	// Allow showing "all the logs"
	//
	if os.Getenv("LOG_ALL") != "" {
		loggerLevel.Set(slog.LevelDebug)
	}

	// Those handler options
	opts := &slog.HandlerOptions{
		Level:     loggerLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				s := a.Value.Any().(*slog.Source)

				// Assume we have a source-path containing "rss2email"
				// if we do strip everything before that out.
				start := strings.Index(s.File, "rss2email")
				if start > 0 {
					s.File = s.File[start:]
				}

				// Assume we have a function containing "rss2email"
				// if we do strip everything before that out.
				start = strings.Index(s.Function, "rss2email")
				if start > 0 {
					s.Function = s.Function[start:]
				}

			}
			return a
		},
	}

	//
	// Default to showing to STDERR in text.
	//
	var handler slog.Handler
	handler = slog.NewTextHandler(os.Stderr, opts)

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
