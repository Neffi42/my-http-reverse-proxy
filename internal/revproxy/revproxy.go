package revproxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type RevProxy struct {
	upstream  *url.URL
	transport http.RoundTripper
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
	rp := &RevProxy{upstream: u, transport: http.DefaultTransport}
	return rp, nil
}

func (rp *RevProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	outReq := r.Clone(r.Context())
	outReq.RequestURI = ""
	outReq.URL.Scheme = rp.upstream.Scheme
	outReq.URL.Host = rp.upstream.Host

	resp, err := rp.transport.RoundTrip(outReq)
	if err != nil {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
