package cache

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type Middleware struct {
	next   http.Handler
	store  *Store
	ttl    time.Duration
	logger *slog.Logger
}

func New(next http.Handler, store *Store, ttl time.Duration, logger *slog.Logger) *Middleware {
	if logger == nil {
		logger = slog.Default()
	}
	m := &Middleware{next: next, store: store, ttl: ttl, logger: logger}
	return m
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("X-Cache", "BYPASS")
		m.next.ServeHTTP(w, r)
		return
	}

	k := key(r)
	now := time.Now()

	if e, ok := m.store.Get(k); ok && e.Fresh(now) {
		m.writeEntry(w, r, e, now)
		return
	}

	rec := newRecorder(w)
	w.Header().Set("X-Cache", "MISS")
	m.next.ServeHTTP(rec, r)

	if rec.status == http.StatusOK {
		hdr := rec.Header().Clone()
		hdr.Del("X-Cache")
		m.store.Set(k, &Entry{
			StatusCode: rec.status,
			Header:     hdr,
			Body:       rec.buf.Bytes(),
			StoredAt:   now,
			ExpiresAt:  now.Add(m.ttl),
		})
	}
}

func (m *Middleware) writeEntry(w http.ResponseWriter, r *http.Request, e *Entry, now time.Time) {
	copyHeader(w.Header(), e.Header)
	w.Header().Set("X-Cache", "HIT")
	w.Header().Set("Age", strconv.Itoa(int(e.Age(now).Seconds())))
	w.WriteHeader(e.StatusCode)

	bytes, err := w.Write(e.Body)
	if err != nil {
		m.logger.WarnContext(r.Context(), "cached response write failed",
			"err", err,
			"bytes_written", bytes,
			"path", r.URL.Path,
		)
	}
}

func copyHeader(dest, src http.Header) {
	for key, vals := range src {
		for _, val := range vals {
			dest.Add(key, val)
		}
	}
}

func key(r *http.Request) string {
	return r.Method + " " + r.Host + r.URL.RequestURI()
}
