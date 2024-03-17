package heartbeat

import (
	"bytes"
	"net/http"
)

// response implements http.ResponseWriter
type response struct {
	buf    bytes.Buffer
	header http.Header
	status int
}

func (r *response) Header() http.Header {
	return r.header
}

func (r *response) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *response) Write(p []byte) (n int, err error) {
	return r.buf.Write(p)
}

func NewResponseWriter() *response { //nolint:revive
	return &response{header: map[string][]string{}, buf: bytes.Buffer{}}
}
