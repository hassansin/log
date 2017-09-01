package log

import "time"
import "github.com/rs/zerolog"

type RequestFields struct {
	RemoteAddr    string        `json:"remote_addr"`
	BodyBytesSent int           `json:"body_bytes_sent"`
	RequestTime   time.Duration `json:"request_time"`
	Status        int           `json:"status"`
	Request       string        `json:"request"`
	Method        string        `json:"request_method"`
	Referrer      string        `json:"http_referrer"`
	UserAgent     string        `json:"http_user_agent"`
}

func (f *RequestFields) MarshalZerologObject(e *zerolog.Event) {
	e.Str("remote_addr", f.RemoteAddr).
		Str("request", f.Request).
		Float64("request_time", f.RequestTime.Seconds()).
		Int("body_bytes_sent", f.BodyBytesSent).
		Int("status", f.Status).
		Str("request_method", f.Method).
		Str("http_referrer", f.Referrer).
		Str("http_user_agent", f.UserAgent)
}
