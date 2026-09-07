package cache

import (
	"bytes"
	"net/http"
)

type recorder struct {
	http.ResponseWriter
	status    int
	buf       bytes.Buffer
	wroteHead bool
}

func newRecorder(w http.ResponseWriter) *recorder {
	return &recorder{ResponseWriter: w, status: http.StatusOK}
}

func (rec *recorder) WriteHeader(code int) {
	if rec.wroteHead {
		return
	}
	rec.wroteHead = true
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *recorder) Write(p []byte) (int, error) {
	if !rec.wroteHead {
		rec.WriteHeader(http.StatusOK)
	}
	rec.buf.Write(p)
	return rec.ResponseWriter.Write(p)
}
