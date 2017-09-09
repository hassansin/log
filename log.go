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

// CorrelationIDKey is used as context key for corelation id
var CorrelationIDKey = "CorrelationID"

// Logger is the default logger instance
var Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

const logKey = "log"
const xCorrelationID = "x-correlation-id"

func init() {
	zerolog.TimestampFieldName = "@timestamp"
	zerolog.LevelFieldName = "@level"
	zerolog.MessageFieldName = "@message"
}

// Init initializes the default logger with application name and logging level
func Init(app string, debug bool) {
	if debug {
		Logger = Logger.Level(zerolog.DebugLevel)
	}
	Logger = Logger.With().Str("@source", app).
		Logger()
}

// getCorrelationID retrieves Correlation ID from HTTP request
// it looks up in the HTTP X-Correlation-ID header first
// if not found in header, it tries to retrieve it from context
// Returns empty string if not found.
func getCorrelationID(r *http.Request) string {
	cid := r.Header.Get(xCorrelationID)
	if cid != "" {
		return cid
	}
	ctx := r.Context()
	if ctx == nil {
		return ""
	}
	cid, ok := ctx.Value(CorrelationIDKey).(string)
	if !ok {
		return ""
	}
	return cid
}

// WithContextLogger adds a logger instance in the request context
func WithContextLogger(r *http.Request) *http.Request {
	cid := getCorrelationID(r)
	l := Logger.With().Str("correlation_id", cid).
		Object("@fields", &RequestFields{
			RemoteAddr: r.RemoteAddr,
			Method:     r.Method,
			Request:    r.RequestURI,
		}).
		Logger()
	r = r.WithContext(context.WithValue(r.Context(), logKey, l))
	return r
}

// FromRequest returns the per-request logger instance.
// If not available, it returns the default logger
func FromRequest(r *http.Request) zerolog.Logger {
	ctx := r.Context()
	if ctx == nil {
		return Logger
	}
	l, ok := ctx.Value(logKey).(zerolog.Logger)
	if !ok {
		return Logger
	}
	return l
}

// WithLogger is a HTTP logging middleware
func WithLogger(handler http.Handler, exclude ...string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		skipping := false
		for _, path := range exclude {
			if path == r.URL.Path {
				skipping = true
				break
			}
		}

		if skipping {
			handler.ServeHTTP(rw, r)
			return
		}

		start := time.Now()
		rww := &responseWriter{ResponseWriter: rw}

		cid := getCorrelationID(r)
		// generate a new correlation id if not available
		// store it in both context and request header
		if cid == "" {
			cid = xid.New().String()
			ctx := context.WithValue(r.Context(), CorrelationIDKey, cid)
			r.Header.Set(xCorrelationID, cid)
			r = r.WithContext(ctx)
		}
		r = WithContextLogger(r)

		handler.ServeHTTP(rww, r)

		Logger.Info().
			Str("correlation_id", cid).
			Object("@fields", &RequestFields{
				RemoteAddr:    r.RemoteAddr,
				BodyBytesSent: rww.written,
				RequestTime:   time.Since(start),
				Status:        rww.status,
				Method:        r.Method,
				Request:       r.RequestURI,
				UserAgent:     r.UserAgent(),
				Referrer:      r.Referer(),
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
