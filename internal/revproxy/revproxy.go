package revproxy

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type RevProxy struct {
	upstream  *url.URL
	transport http.RoundTripper
	logger    *slog.Logger
}

func New(rawUpstream string) (*RevProxy, error) {
	u, err := url.Parse(rawUpstream)
	if err != nil {
		return nil, fmt.Errorf("parse upstream: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("upstream scheme must be http or https, got: %q", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("upstream must include a host")
	}
	rp := &RevProxy{upstream: u, transport: http.DefaultTransport, logger: slog.Default()}
	return rp, nil
}

func (rp *RevProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	outReq := r.Clone(r.Context())
	outReq.RequestURI = ""
	outReq.URL.Scheme = rp.upstream.Scheme
	outReq.URL.Host = rp.upstream.Host
	outReq.URL.Path = joinPaths(rp.upstream.Path, r.URL.Path)

	deleteHopByHopHeaders(outReq.Header)
	setForwardedHeaders(r, outReq)

	resp, err := rp.transport.RoundTrip(outReq)
	if err != nil {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	deleteHopByHopHeaders(w.Header())
	w.WriteHeader(resp.StatusCode)

	bytes, err := io.Copy(w, resp.Body)
	if err != nil {
		rp.logger.WarnContext(r.Context(), "response body copy failed",
			"err", err,
			"bytes_written", bytes,
			"path", r.URL.Path,
			"upstream", rp.upstream.Host,
		)
	}
}

var hopByHopHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Keep-Alive",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

func deleteHopByHopHeaders(h http.Header) {
	for _, v := range h.Values("Connection") {
		for _, key := range strings.Split(v, ",") {
			key = strings.TrimSpace(key)
			if key != "" {
				h.Del(key)
			}

		}
	}

	for _, key := range hopByHopHeaders {
		h.Del(key)
	}
}

func copyHeader(dest, src http.Header) {
	for key, vals := range src {
		for _, val := range vals {
			dest.Add(key, val)
		}
	}
}

func setForwardedHeaders(inReq, outReq *http.Request) {
	xForwardedFor := "X-Forwarded-For"
	xForwardedProto := "X-Forwarded-Proto"
	xForwardedHost := "X-Forwarded-Host"

	ip, _, err := net.SplitHostPort(inReq.RemoteAddr)
	if err == nil {
		previousProxies := inReq.Header[xForwardedFor]
		if len(previousProxies) > 0 {
			ip = strings.Join(previousProxies, ", ") + ", " + ip
		}
		outReq.Header.Set(xForwardedFor, ip)
	} else {
		outReq.Header.Del(xForwardedFor)
	}

	if inReq.TLS == nil {
		outReq.Header.Set(xForwardedProto, "http")
	} else {
		outReq.Header.Set(xForwardedProto, "https")
	}

	outReq.Header.Set(xForwardedHost, inReq.Host)
}

func joinPaths(base, req string) string {
	if base == "" || base == "/" {
		return req
	} else if req == "" || req == "/" {
		return base
	}
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(req, "/")
}
