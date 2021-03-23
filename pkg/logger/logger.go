package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func NewLogger() zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	logger := zerolog.New(output).With().Timestamp().Logger()
	return logger
}
