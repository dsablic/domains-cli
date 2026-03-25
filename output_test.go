package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeTypes(t *testing.T) {
	tests := []struct {
		input []string
		want  []string
	}{
		{nil, nil},
		{[]string{"a", "cname", "Mx"}, []string{"A", "CNAME", "MX"}},
	}
	for _, tt := range tests {
		got := NormalizeTypes(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("NormalizeTypes(%v) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("NormalizeTypes(%v)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestMatchesType(t *testing.T) {
	tests := []struct {
		recordType string
		types      []string
		want       bool
	}{
		{"A", nil, true},
		{"A", []string{"A"}, true},
		{"a", []string{"A"}, true},
		{"CNAME", []string{"A", "CNAME"}, true},
		{"MX", []string{"A", "CNAME"}, false},
	}
	for _, tt := range tests {
		got := matchesType(tt.recordType, tt.types)
		if got != tt.want {
			t.Errorf("matchesType(%q, %v) = %v, want %v", tt.recordType, tt.types, got, tt.want)
		}
	}
}

func TestOutputTSV_NoCert(t *testing.T) {
	records := []Record{
		{Domain: "example.com", Name: "example.com", Value: "1.2.3.4", Type: "A", Source: "cloudflare", Registrar: "test"},
	}
	var buf bytes.Buffer
	if err := OutputTSV(&buf, records, false); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "domain") {
		t.Error("missing header")
	}
	if strings.Contains(lines[0], "cert_issuer") {
		t.Error("cert columns should not appear when hasCert is false")
	}
}

func TestOutputTSV_WithCert(t *testing.T) {
	records := []Record{
		{Domain: "example.com", Name: "example.com", Value: "1.2.3.4", Type: "A", Source: "cloudflare", Registrar: "test",
			CertIssuer: "Let's Encrypt", CertExpires: "2025-01-01", TLSVersion: "TLS 1.3"},
	}
	var buf bytes.Buffer
	if err := OutputTSV(&buf, records, true); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if !strings.Contains(lines[0], "tls_version") {
		t.Error("missing tls_version header")
	}
	if !strings.Contains(lines[1], "TLS 1.3") {
		t.Error("missing TLS version in data")
	}
}

func TestOutputJSON(t *testing.T) {
	records := []Record{
		{Domain: "example.com", Name: "example.com", Value: "1.2.3.4", Type: "A", Source: "cloudflare", Registrar: "test"},
	}
	var buf bytes.Buffer
	if err := OutputJSON(&buf, records); err != nil {
		t.Fatal(err)
	}
	var parsed []Record
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(parsed) != 1 || parsed[0].Domain != "example.com" {
		t.Errorf("unexpected parsed result: %+v", parsed)
	}
}
