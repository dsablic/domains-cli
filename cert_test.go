package main

import (
	"testing"
)

func TestShouldCheckCert(t *testing.T) {
	tests := []struct {
		recordType string
		want       bool
	}{
		{"A", true},
		{"AAAA", true},
		{"CNAME", true},
		{"MX", false},
		{"NS", false},
		{"TXT", false},
	}
	for _, tt := range tests {
		got := shouldCheckCert(tt.recordType)
		if got != tt.want {
			t.Errorf("shouldCheckCert(%q) = %v, want %v", tt.recordType, got, tt.want)
		}
	}
}

func TestTLSVersionName(t *testing.T) {
	tests := []struct {
		version uint16
		want    string
	}{
		{0x0301, "TLS 1.0"},
		{0x0302, "TLS 1.1"},
		{0x0303, "TLS 1.2"},
		{0x0304, "TLS 1.3"},
		{0x0000, "unknown (0x0000)"},
	}
	for _, tt := range tests {
		got := tlsVersionName(tt.version)
		if got != tt.want {
			t.Errorf("tlsVersionName(0x%04x) = %q, want %q", tt.version, got, tt.want)
		}
	}
}
