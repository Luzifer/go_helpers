package appauth

import "testing"

func TestIsLocalHost(t *testing.T) {
	for _, tc := range []struct {
		hostport string
		want     bool
	}{
		{hostport: "localhost", want: true},
		{hostport: "localhost:3000", want: true},
		{hostport: "127.0.0.1", want: true},
		{hostport: "127.0.0.1:8080", want: true},
		{hostport: "[::1]:8080", want: true},
		{hostport: "::1", want: true},
		{hostport: "example.com", want: false},
		{hostport: "example.com:443", want: false},
	} {
		if got := isLocalHost(tc.hostport); got != tc.want {
			t.Errorf("isLocalHost(%q) = %v, want %v", tc.hostport, got, tc.want)
		}
	}
}
