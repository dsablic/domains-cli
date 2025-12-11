package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"
)

type CertInfo struct {
	Issuer  string
	Expires string
	Error   string
}

func FetchCertificates(records []Record) {
	lookupTargets := make(map[string][]int)

	for i, r := range records {
		if shouldCheckCert(r.Type) {
			name := r.Name
			lookupTargets[name] = append(lookupTargets[name], i)
		}
	}

	results := make(map[string]CertInfo)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name := range lookupTargets {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			info := lookupCert(name)
			mu.Lock()
			results[name] = info
			mu.Unlock()
		}(name)
	}

	wg.Wait()

	for i := range records {
		if !shouldCheckCert(records[i].Type) {
			records[i].CertIssuer = "n/a"
			records[i].CertExpires = "n/a"
			continue
		}

		info := results[records[i].Name]
		records[i].CertIssuer = info.Issuer
		records[i].CertExpires = info.Expires
		records[i].CertError = info.Error
	}
}

func shouldCheckCert(recordType string) bool {
	switch recordType {
	case "A", "AAAA", "CNAME":
		return true
	default:
		return false
	}
}

func lookupCert(hostname string) CertInfo {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", hostname+":443", &tls.Config{
		ServerName:         hostname,
		InsecureSkipVerify: true,
	})
	if err != nil {
		return CertInfo{Error: formatCertError(err)}
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return CertInfo{Error: "no certificate"}
	}

	cert := certs[0]
	issuer := cert.Issuer.Organization
	issuerStr := ""
	if len(issuer) > 0 {
		issuerStr = issuer[0]
	} else if cert.Issuer.CommonName != "" {
		issuerStr = cert.Issuer.CommonName
	} else {
		issuerStr = "unknown"
	}

	return CertInfo{
		Issuer:  issuerStr,
		Expires: cert.NotAfter.Format("2006-01-02"),
	}
}

func formatCertError(err error) string {
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return "timeout"
	}
	if opErr, ok := err.(*net.OpError); ok {
		if opErr.Op == "dial" {
			return "connection refused"
		}
	}
	return fmt.Sprintf("%v", err)
}
