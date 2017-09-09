package log

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetCorrelationID(t *testing.T) {
	t.Run("empty cid", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		if want, got := "", getCorrelationID(r); want != got {
			t.Errorf("Expected: %s\nGot: %s\n", want, got)
		}
	})
	t.Run("cid in header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Correlation-ID", "123")
		if want, got := "123", getCorrelationID(r); want != got {
			t.Errorf("Expected: %s\nGot: %s\n", want, got)
		}
	})
	t.Run("cid in context", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r = r.WithContext(context.WithValue(r.Context(), CorrelationIDKey, "234"))
		if want, got := "234", getCorrelationID(r); want != got {
			t.Errorf("Expected: %s\nGot: %s\n", want, got)
		}
	})
}

func TestWithContextLogger(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Correlation-ID", "123")
	r = WithContextLogger(r)
	l := FromRequest(r)
	var b bytes.Buffer
	l.Output(&b).Log().Msg("")
	if !strings.Contains(b.String(), "\"correlation_id\":\"123\"") {
		t.Errorf("Expected log to contain correlation_id field")
	}
}

func TestWithLogger(t *testing.T) {
	t.Run("test logging", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			w.Write([]byte("unexpecte error"))
			if _, ok := r.Context().Value(CorrelationIDKey).(string); !ok {
				t.Errorf("Got empty correlation ID in context")
			}
		}
		r := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		h := WithLogger(http.HandlerFunc(handler))
		var b bytes.Buffer
		Logger = Logger.Output(&b)
		h.ServeHTTP(w, r)
		if got := r.Header.Get("X-Correlation-ID"); got == "" {
			t.Errorf("Got empty correlation ID in header")
		}
		if !strings.Contains(b.String(), "\"status\":500") {
			t.Errorf("Expected 500 status code")
		}
		if !strings.Contains(b.String(), "\"body_bytes_sent\":15") {
			t.Errorf("Expected body_bytes_sent=15")
		}
		if strings.Contains(b.String(), "\"correlation_id\":\"\"") {
			t.Errorf("Expected non-empty correation_id field in the log")
		}
	})
	t.Run("test exclude", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {}
		r := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		h := WithLogger(http.HandlerFunc(handler), "/test")
		var b bytes.Buffer
		Logger = Logger.Output(&b)
		h.ServeHTTP(w, r)
		if "" != r.Header.Get("X-Correlation-ID") {
			t.Errorf("Expected empty X-Correlation-ID header")
		}
		if b.String() != "" {
			t.Errorf("Expected: %s\nGot: %s", "", b.String())
		}
	})
}
