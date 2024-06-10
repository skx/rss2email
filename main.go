//
// Entry-point for our application.
//

package main

import (
	"fmt"
	"io"
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
	// Setup our default logging level, which will show
	// both warnings and errors.
	//
	loggerLevel = &slog.LevelVar{}
	loggerLevel.Set(slog.LevelWarn)

	//
	// If the user wants a different level they can choose it.
	//
	level := os.Getenv("LOG_LEVEL")

	//
	// Legacy/Compatibility
	//
	if os.Getenv("LOG_ALL") != "" {
		level = "DEBUG"
	}

	// Simplify things by only caring about upper-case
	level = strings.ToUpper(level)

	switch level {
	case "DEBUG":
		loggerLevel.Set(slog.LevelDebug)
	case "WARN":
		loggerLevel.Set(slog.LevelWarn)
	case "ERROR":
		loggerLevel.Set(slog.LevelError)
	default:
		fmt.Printf("Unknown logging-level '%s'\n", level)
		return
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
	// Create a default writer, which the logger will use.
	// This will mostly go to STDERR, however it might also
	// be duplicated to a file.
	//
	multi := io.MultiWriter(os.Stderr)

	//
	// Default logfile path can be changed by LOG_FILE
	// environmental variable.
	//
	logPath := "rss2email.log"
	if os.Getenv("LOG_FILE_PATH") != "" {
		logPath = os.Getenv("LOG_FILE_PATH")
	}

	//
	// Create a logfile, if we can.
	//
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	//
	// No error?  Then update our writer to use it.
	//
	if err == nil {
		defer file.Close()

		//
		// Unless we've been disabled then update our
		// writer.
		//
		if os.Getenv("LOG_FILE_DISABLE") != "" {
			multi = io.MultiWriter(file, os.Stderr)
		}
	}

	//
	// Default to showing to STDERR [+file] in text.
	//
	var handler slog.Handler
	handler = slog.NewTextHandler(multi, opts)

	//
	// But allow JSON formatting too.
	//
	if os.Getenv("LOG_JSON") != "" {
		handler = slog.NewJSONHandler(multi, opts)
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
