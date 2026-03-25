package main

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/likexian/whois"
)

type WhoisClient struct {
	client *whois.Client
	cache  map[string]string
	mu     sync.Mutex
}

func NewWhoisClient() *WhoisClient {
	client := whois.NewClient()
	client.SetTimeout(10 * time.Second)
	return &WhoisClient{
		client: client,
		cache:  make(map[string]string),
	}
}

var registrarPattern = regexp.MustCompile(`(?i)^Registrar:\s*(.+)$`)

func (c *WhoisClient) LookupRegistrar(domain string) string {
	c.mu.Lock()
	if registrar, ok := c.cache[domain]; ok {
		c.mu.Unlock()
		return registrar
	}
	c.mu.Unlock()

	registrar := c.fetchRegistrar(domain)

	c.mu.Lock()
	c.cache[domain] = registrar
	c.mu.Unlock()

	return registrar
}

func (c *WhoisClient) fetchRegistrar(domain string) string {
	result, err := c.client.Whois(domain)
	if err != nil {
		return "unknown"
	}

	for _, line := range strings.Split(result, "\n") {
		if matches := registrarPattern.FindStringSubmatch(line); len(matches) > 1 {
			return strings.ToLower(strings.TrimSpace(matches[1]))
		}
	}

	return "unknown"
}
