package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func OutputTSV(w io.Writer, records []Record) error {
	hasCert := len(records) > 0 && (records[0].CertIssuer != "" || records[0].CertExpires != "" || records[0].CertError != "")

	if hasCert {
		fmt.Fprintln(w, "domain\trecord\tvalue\ttype\tsource\tregistrar\tcert_issuer\tcert_expires\tcert_error")
		for _, r := range records {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				r.Domain, r.Name, r.Value, r.Type, r.Source, r.Registrar, r.CertIssuer, r.CertExpires, r.CertError)
		}
	} else {
		fmt.Fprintln(w, "domain\trecord\tvalue\ttype\tsource\tregistrar")
		for _, r := range records {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				r.Domain, r.Name, r.Value, r.Type, r.Source, r.Registrar)
		}
	}
	return nil
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
