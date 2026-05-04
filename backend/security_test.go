package backend

import (
	"testing"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "''"},
		{"simple", "simple"},
		{"with spaces", "'with spaces'"},
		{"with'quote", "'with'\\''quote'"},
		{"$VAR", "'$VAR'"},
		{"`backtick`", "'`backtick`'"},
		{`"double quote"`, `'"double quote"'`},
	}

	for _, test := range tests {
		if got := shellQuote(test.input); got != test.expected {
			t.Errorf("ShellQuote(%q) = %q, want %q", test.input, got, test.expected)
		}
	}
}
