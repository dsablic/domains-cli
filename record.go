package main

type Record struct {
	Domain    string `json:"domain"`
	Name      string `json:"record"`
	Value     string `json:"value"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	Registrar string `json:"registrar"`
}
