package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"
)

type CertInfo struct {
	Issuer     string
	Expires    string
	TLSVersion string
	Error      string
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

	sem := make(chan struct{}, 10)
	for name := range lookupTargets {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			sem <- struct{}{}
			info := lookupCert(name)
			<-sem
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
			records[i].TLSVersion = "n/a"
			continue
		}

		info := results[records[i].Name]
		records[i].CertIssuer = info.Issuer
		records[i].CertExpires = info.Expires
		records[i].TLSVersion = info.TLSVersion
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

	state := conn.ConnectionState()
	certs := state.PeerCertificates
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
		Issuer:     issuerStr,
		Expires:    cert.NotAfter.Format("2006-01-02"),
		TLSVersion: tlsVersionName(state.Version),
	}
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("unknown (0x%04x)", version)
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
