package log

import "net/http"

type responseWriter struct {
	status  int
	written int
	http.ResponseWriter
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	rw.written += len(data)
	return rw.ResponseWriter.Write(data)
}
