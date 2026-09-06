package revproxy

import "testing"

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
