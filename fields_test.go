package log

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRequestFields(t *testing.T) {
	t.Run("string-field", func(t *testing.T) {
		out := &bytes.Buffer{}
		l := Logger.Output(out)
		rf := RequestFields{
			RemoteAddr: "127.0.0.1",
		}
		l.Info().
			Object("@fields", &rf).
			Msg("")
		if got, want := out.String(), "\"remote_addr\":\"127.0.0.1\","; strings.Index(got, want) == -1 {
			t.Errorf("invalid log output:\ngot:  %v\nwant: %v", got, want)
		}
	})
	t.Run("int-field", func(t *testing.T) {
		out := &bytes.Buffer{}
		l := Logger.Output(out)
		rf := RequestFields{
			Status: 200,
		}
		l.Info().
			Object("@fields", &rf).
			Msg("")
		if got, want := out.String(), "\"status\":200,"; strings.Index(got, want) == -1 {
			t.Errorf("invalid log output:\ngot:  %v\nwant: %v", got, want)
		}
	})
	t.Run("duration-field", func(t *testing.T) {
		out := &bytes.Buffer{}
		l := Logger.Output(out)
		rf := RequestFields{
			RequestTime: time.Duration(500 * time.Millisecond),
		}
		l.Info().
			Object("@fields", &rf).
			Msg("")
		if got, want := out.String(), "\"request_time\":0.5,"; strings.Index(got, want) == -1 {
			t.Errorf("invalid log output:\ngot:  %v\nwant: %v", got, want)
		}
	})
}
