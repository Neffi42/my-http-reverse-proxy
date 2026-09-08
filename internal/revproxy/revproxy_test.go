package revproxy

import (
	"net/url"
	"testing"
)

func TestJoinPaths(t *testing.T) {
	tests := []struct {
		name            string
		base, req, want string
	}{
		{"empty base", "", "/users", "/users"},
		{"root base", "/", "/users", "/users"},
		{"simple join", "/api", "/users", "/api/users"},
		{"base trailing slash", "/api/", "/users", "/api/users"},
		{"req trailing slash preserved", "/api", "/users/", "/api/users/"},
		{"req is root", "/api", "/", "/api/"},
		{"both slashes", "/api/", "/", "/api/"},
		{"empty req", "/api", "", "/api/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinPaths(tt.base, tt.req); got != tt.want {
				t.Errorf("joinPaths(%q, %q) = %q, want %q", tt.base, tt.req, got, tt.want)
			}
		})
	}
}

func TestJoinURLPaths(t *testing.T) {
	mustParse := func(t *testing.T, raw string) *url.URL {
		t.Helper()
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("url.Parse(%q): %v", raw, err)
		}
		return u
	}

	tests := []struct {
		name        string
		base, req   string
		wantPath    string
		wantRawPath string
	}{
		{
			name:        "no escaping either side",
			base:        "http://up/api",
			req:         "http://client/users",
			wantPath:    "/api/users",
			wantRawPath: "",
		},
		{
			name:        "encoded slash in request is preserved",
			base:        "http://up/api",
			req:         "http://client/foo%2Fbar",
			wantPath:    "/api/foo/bar",
			wantRawPath: "/api/foo%2Fbar",
		},
		{
			name:        "encoded slash in upstream base path is preserved",
			base:        "http://up/svc%2Fv1",
			req:         "http://client/users",
			wantPath:    "/svc/v1/users",
			wantRawPath: "/svc%2Fv1/users",
		},
		{
			name:        "other percent-encoding does not force a RawPath",
			base:        "http://up/api",
			req:         "http://client/caf%C3%A9",
			wantPath:    "/api/café",
			wantRawPath: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := mustParse(t, tt.base)
			req := mustParse(t, tt.req)

			gotPath, gotRawPath := joinURLPaths(base, req)
			if gotPath != tt.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tt.wantPath)
			}
			if gotRawPath != tt.wantRawPath {
				t.Errorf("rawPath = %q, want %q", gotRawPath, tt.wantRawPath)
			}
		})
	}
}
