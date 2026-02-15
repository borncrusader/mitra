package util

import (
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"
)

func NewLogger(out io.Writer) zerolog.Logger {
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	return zerolog.New(zerolog.ConsoleWriter{
		Out: out,
		FormatTimestamp: func(i interface{}) string {
			if t, ok := i.(string); ok {
				parsed, err := time.Parse(time.RFC3339Nano, t)
				if err == nil {
					return parsed.UTC().Format("2006-01-02T15:04:05.000Z")
				}
			}
			return fmt.Sprintf("%v", i)
		},
	}).
		With().
		Timestamp().
		Logger()
}
