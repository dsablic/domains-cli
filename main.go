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
	resolveRegistrars(records, r53Client, whoisClient)

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
	case "tsv":
		return OutputTSV(os.Stdout, records, fetchCert)
	case "json":
		return OutputJSON(os.Stdout, records)
	default:
		return fmt.Errorf("unknown output format %q (supported: tsv, json)", format)
	}
}

func resolveRegistrars(records []Record, r53Client *Route53Client, whoisClient *WhoisClient) {
	domains := make(map[string]struct{})
	for _, r := range records {
		domains[r.Domain] = struct{}{}
	}

	resolved := make(map[string]string, len(domains))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for domain := range domains {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			sem <- struct{}{}
			registrar := resolveRegistrar(d, r53Client, whoisClient)
			<-sem
			mu.Lock()
			resolved[d] = registrar
			mu.Unlock()
		}(domain)
	}
	wg.Wait()

	for i := range records {
		records[i].Registrar = resolved[records[i].Domain]
	}
}

func resolveRegistrar(domain string, r53Client *Route53Client, whoisClient *WhoisClient) string {
	if r53Client != nil && r53Client.IsRoute53Registrar(domain) {
		return "route53"
	}

	return whoisClient.LookupRegistrar(domain)
}
