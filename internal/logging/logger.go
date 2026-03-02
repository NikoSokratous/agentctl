package logging

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes structured logging.
func Init(level string, pretty bool) {
	// Set global log level
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Configure output format
	var output io.Writer = os.Stderr
	if pretty {
		output = zerolog.ConsoleWriter{Out: os.Stderr}
	}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}

// Logger returns the global logger
func Logger() *zerolog.Logger {
	return &log.Logger
}
