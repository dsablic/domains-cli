package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func OutputTSV(w io.Writer, records []Record, hasCert bool) error {
	if hasCert {
		fmt.Fprintln(w, "domain\trecord\tvalue\ttype\tsource\tregistrar\tcert_issuer\tcert_expires\ttls_min\ttls_max\tcert_error")
		for _, r := range records {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				sanitizeTSV(r.Domain), sanitizeTSV(r.Name), sanitizeTSV(r.Value), sanitizeTSV(r.Type),
				sanitizeTSV(r.Source), sanitizeTSV(r.Registrar), sanitizeTSV(r.CertIssuer), sanitizeTSV(r.CertExpires),
				sanitizeTSV(r.TLSMinVersion), sanitizeTSV(r.TLSMaxVersion), sanitizeTSV(r.CertError))
		}
	} else {
		fmt.Fprintln(w, "domain\trecord\tvalue\ttype\tsource\tregistrar")
		for _, r := range records {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				sanitizeTSV(r.Domain), sanitizeTSV(r.Name), sanitizeTSV(r.Value), sanitizeTSV(r.Type),
				sanitizeTSV(r.Source), sanitizeTSV(r.Registrar))
		}
	}
	return nil
}

var tsvReplacer = strings.NewReplacer("\t", " ", "\n", " ", "\r", "")

func sanitizeTSV(s string) string {
	return tsvReplacer.Replace(s)
}

func OutputJSON(w io.Writer, records []Record) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(records)
}

func NormalizeTypes(args []string) []string {
	var types []string
	for _, arg := range args {
		types = append(types, strings.ToUpper(arg))
	}
	return types
}
