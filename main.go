package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"
)

func main() {
	format := flag.String("format", "tsv", "Output format: tsv or json")
	flag.StringVar(format, "f", "tsv", "Output format: tsv or json (shorthand)")
	cert := flag.Bool("cert", false, "Fetch TLS certificate info (issuer, expiry)")
	flag.BoolVar(cert, "c", false, "Fetch TLS certificate info (shorthand)")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.BoolVar(showVersion, "v", false, "Print version and exit (shorthand)")
	flag.Parse()

	if *showVersion {
		fmt.Println(Version())
		return
	}

	types := NormalizeTypes(flag.Args())

	if err := run(context.Background(), types, *format, *cert); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, types []string, format string, fetchCert bool) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	var records []Record
	var cfClient *CloudflareClient
	var r53Client *Route53Client
	var cfErr, r53Err error
	var cfRecords, r53Records []Record

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		cfClient, cfErr = NewCloudflareClient(cfg.Cloudflare)
		if cfErr != nil {
			fmt.Fprintf(os.Stderr, "warning: cloudflare: %v\n", cfErr)
			return
		}
		if cfClient == nil {
			fmt.Fprintln(os.Stderr, "warning: cloudflare credentials not configured, skipping")
			return
		}
		cfRecords, cfErr = cfClient.FetchRecords(ctx, types)
		if cfErr != nil {
			fmt.Fprintf(os.Stderr, "warning: cloudflare: %v\n", cfErr)
		}
	}()

	go func() {
		defer wg.Done()
		r53Client, r53Err = NewRoute53Client(ctx)
		if r53Err != nil {
			fmt.Fprintf(os.Stderr, "warning: route53: %v\n", r53Err)
			return
		}
		r53Records, r53Err = r53Client.FetchRecords(ctx, types)
		if r53Err != nil {
			fmt.Fprintf(os.Stderr, "warning: route53: %v\n", r53Err)
		}
	}()

	wg.Wait()

	if cfClient == nil && r53Client == nil {
		return fmt.Errorf("no credentials configured for cloudflare or aws")
	}

	records = append(records, cfRecords...)
	records = append(records, r53Records...)

	whoisClient := NewWhoisClient()
	resolveRegistrars(records, cfClient, r53Client, whoisClient)

	sort.Slice(records, func(i, j int) bool {
		if records[i].Domain != records[j].Domain {
			return records[i].Domain < records[j].Domain
		}
		return records[i].Name < records[j].Name
	})

	if fetchCert {
		FetchCertificates(records)
	}

	switch format {
	case "json":
		return OutputJSON(os.Stdout, records)
	default:
		return OutputTSV(os.Stdout, records)
	}
}

func resolveRegistrars(records []Record, cfClient *CloudflareClient, r53Client *Route53Client, whoisClient *WhoisClient) {
	resolved := make(map[string]string)

	for i := range records {
		domain := records[i].Domain

		if registrar, ok := resolved[domain]; ok {
			records[i].Registrar = registrar
			continue
		}

		registrar := resolveRegistrar(domain, cfClient, r53Client, whoisClient)
		resolved[domain] = registrar
		records[i].Registrar = registrar
	}
}

func resolveRegistrar(domain string, cfClient *CloudflareClient, r53Client *Route53Client, whoisClient *WhoisClient) string {
	if cfClient != nil && cfClient.IsCloudflareRegistrar(domain) {
		return "cloudflare"
	}

	if r53Client != nil && r53Client.IsRoute53Registrar(domain) {
		return "route53"
	}

	return whoisClient.LookupRegistrar(domain)
}
