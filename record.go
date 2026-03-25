package main

type Record struct {
	Domain      string `json:"domain"`
	Name        string `json:"record"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	Registrar   string `json:"registrar"`
	CertIssuer  string `json:"cert_issuer,omitempty"`
	CertExpires string `json:"cert_expires,omitempty"`
	TLSMinVersion string `json:"tls_min_version,omitempty"`
	TLSMaxVersion string `json:"tls_max_version,omitempty"`
	CertError   string `json:"cert_error,omitempty"`
}
