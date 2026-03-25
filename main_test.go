package main

import (
	"testing"
)

func TestResolveRegistrar_Route53(t *testing.T) {
	whois := NewWhoisClient()
	// nil r53Client should fall through to WHOIS
	result := resolveRegistrar("example.com", nil, whois)
	// We can't predict WHOIS result, but it shouldn't panic
	if result == "" {
		t.Error("expected non-empty registrar")
	}
}

func TestResolveRegistrars_Caching(t *testing.T) {
	records := []Record{
		{Domain: "test1.example", Name: "a.test1.example"},
		{Domain: "test1.example", Name: "b.test1.example"},
		{Domain: "test2.example", Name: "a.test2.example"},
	}

	whois := NewWhoisClient()
	resolveRegistrars(records, nil, whois)

	// All records for the same domain should have the same registrar
	if records[0].Registrar != records[1].Registrar {
		t.Errorf("same domain got different registrars: %q vs %q", records[0].Registrar, records[1].Registrar)
	}

	// All records should have some registrar set
	for i, r := range records {
		if r.Registrar == "" {
			t.Errorf("record %d has empty registrar", i)
		}
	}
}
