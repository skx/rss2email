package emailer

import (
	"strings"
	"testing"
)

func TestMakeListIdHeader(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "http://example.com/feed",
			expected: "example.com.feed.localhost",
		},
		{
			input:    "https://example.com/feed",
			expected: "example.com.feed.localhost",
		},
		{
			input:    "https://example.com/foo/bar?baz=qux",
			expected: "example.com.foo.bar.baz=qux.localhost",
		},
		{
			input:    "example.com/feed",
			expected: "example.com.feed.localhost",
		},
		{
			input:    "https://example.com/foo//bar///baz",
			expected: "example.com.foo.bar.baz.localhost",
		},
		{
			input:    "https://example.com/foo@bar#baz",
			expected: "example.com.foo.bar.baz.localhost",
		},
	}

	for _, tt := range tests {
		got := makeListIdHeader(tt.input)
		if got != tt.expected {
			t.Errorf("makeListIdHeader(%q) = %q; want %q", tt.input, got, tt.expected)
		}
        // Check that every character in got is allowed
		// from https://datatracker.ietf.org/doc/html/rfc2822#section-3.2.4
        allowed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!#$%&'*+-/=?^_`{|}~."
        for i, c := range got {
            if !strings.ContainsRune(allowed, c) {
                t.Errorf("makeListIdHeader(%q) produced invalid char %q at position %d", tt.input, c, i)
            }
        }
	}
}