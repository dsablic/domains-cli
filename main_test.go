package main

import (
	"testing"
)

type mockRegistrarLookup struct {
	registrar string
	calls     int
}

func (m *mockRegistrarLookup) LookupRegistrar(domain string) string {
	m.calls++
	return m.registrar
}

func TestResolveRegistrar_Route53(t *testing.T) {
	mock := &mockRegistrarLookup{registrar: "some-registrar"}
	result := resolveRegistrar("example.com", nil, mock)
	if result != "some-registrar" {
		t.Errorf("expected %q, got %q", "some-registrar", result)
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 lookup call, got %d", mock.calls)
	}
}

func TestResolveRegistrars_Caching(t *testing.T) {
	records := []Record{
		{Domain: "test1.example", Name: "a.test1.example"},
		{Domain: "test1.example", Name: "b.test1.example"},
		{Domain: "test2.example", Name: "a.test2.example"},
	}

	mock := &mockRegistrarLookup{registrar: "mock-registrar"}
	resolveRegistrars(records, nil, mock)

	if records[0].Registrar != records[1].Registrar {
		t.Errorf("same domain got different registrars: %q vs %q", records[0].Registrar, records[1].Registrar)
	}

	for i, r := range records {
		if r.Registrar != "mock-registrar" {
			t.Errorf("record %d: expected %q, got %q", i, "mock-registrar", r.Registrar)
		}
	}

	if mock.calls != 2 {
		t.Errorf("expected 2 lookup calls (one per unique domain), got %d", mock.calls)
	}
}
