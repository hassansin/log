package log

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

var CorrelationIdKey = "CorrelationId"
var Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

func init() {
	zerolog.TimestampFieldName = "@timestamp"
	zerolog.LevelFieldName = "@level"
	zerolog.MessageFieldName = "@message"
}
func Init(app string, debug bool) {
	if debug {
		Logger = Logger.Level(zerolog.DebugLevel)
	}
	Logger = Logger.With().Str("@source", app).
		Logger()
}

func WithLogger(handler http.Handler, exclude []string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rww := &responseWriter{ResponseWriter: rw}
		cid := xid.New().String()
		ctx := context.WithValue(r.Context(), CorrelationIdKey, cid)
		r = r.WithContext(ctx)
		handler.ServeHTTP(rww, r)

		for _, path := range exclude {
			if path == r.URL.Path {
				return
			}
		}

		Logger.Info().
			Str("correlation_id", cid).
			Object("@fields", &RequestFields{
				RemoteAddr:    r.RemoteAddr,
				BodyBytesSent: rww.written,
				RequestTime:   time.Since(start),
				Status:        rww.status,
				Method:        r.Method,
				Request:       r.RequestURI,
			}).
			Msg("")
	})
}

// Output duplicates the global logger and sets w as its output.
func Output(w io.Writer) zerolog.Logger {
	return Logger.Output(w)
}

// With creates a child logger with the field added to its context.
func With() zerolog.Context {
	return Logger.With()
}

// Level crestes a child logger with the minium accepted level set to level.
func Level(level zerolog.Level) zerolog.Logger {
	return Logger.Level(level)
}

// Sample returns a logger with the s sampler.
func Sample(s zerolog.Sampler) zerolog.Logger {
	return Logger.Sample(s)
}

// Debug starts a new message with debug level.
//
// You must call Msg on the returned event in order to send the event.
func Debug() *zerolog.Event {
	return Logger.Debug()
}

// Info starts a new message with info level.
//
// You must call Msg on the returned event in order to send the event.
func Info() *zerolog.Event {
	return Logger.Info()
}

// Warn starts a new message with warn level.
//
// You must call Msg on the returned event in order to send the event.
func Warn() *zerolog.Event {
	return Logger.Warn()
}

// Error starts a new message with error level.
//
// You must call Msg on the returned event in order to send the event.
func Error() *zerolog.Event {
	return Logger.Error()
}

// Fatal starts a new message with fatal level. The os.Exit(1) function
// is called by the Msg method.
//
// You must call Msg on the returned event in order to send the event.
func Fatal() *zerolog.Event {
	return Logger.Fatal()
}

// Panic starts a new message with panic level. The message is also sent
// to the panic function.
//
// You must call Msg on the returned event in order to send the event.
func Panic() *zerolog.Event {
	return Logger.Panic()
}

// Log starts a new message with no level. Setting zerolog.GlobalLevel to
// zerlog.Disabled will still disable events produced by this method.
//
// You must call Msg on the returned event in order to send the event.
func Log() *zerolog.Event {
	return Logger.Log()
}

// Ctx returns the Logger associated with the ctx. If no logger
// is associated, a disabled logger is returned.
func Ctx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
