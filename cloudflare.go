package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

type CloudflareClient struct {
	api   *cloudflare.API
	zones []cloudflare.Zone
}

func NewCloudflareClient(cfg CloudflareConfig) (*CloudflareClient, error) {
	if cfg.APIKey == "" || cfg.Email == "" {
		return nil, nil
	}

	api, err := cloudflare.New(cfg.APIKey, cfg.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudflare client: %w", err)
	}

	return &CloudflareClient{api: api}, nil
}

func (c *CloudflareClient) FetchRecords(ctx context.Context, types []string) ([]Record, error) {
	zones, err := c.api.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list cloudflare zones: %w", err)
	}
	c.zones = zones

	var records []Record
	for _, zone := range zones {
		zoneRecords, err := c.fetchZoneRecords(ctx, zone, types)
		if err != nil {
			return nil, err
		}
		records = append(records, zoneRecords...)
	}

	return records, nil
}

func (c *CloudflareClient) fetchZoneRecords(ctx context.Context, zone cloudflare.Zone, types []string) ([]Record, error) {
	var records []Record

	dnsRecords, _, err := c.api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zone.ID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list dns records for zone %s: %w", zone.Name, err)
	}

	for _, r := range dnsRecords {
		if !matchesType(r.Type, types) {
			continue
		}
		records = append(records, Record{
			Domain: zone.Name,
			Name:   r.Name,
			Value:  r.Content,
			Type:   r.Type,
			Source: "cloudflare",
		})
	}

	if matchesType("NS", types) {
		for _, ns := range zone.NameServers {
			records = append(records, Record{
				Domain: zone.Name,
				Name:   zone.Name,
				Value:  ns,
				Type:   "NS",
				Source: "cloudflare",
			})
		}
	}

	return records, nil
}

func (c *CloudflareClient) Zones() []cloudflare.Zone {
	return c.zones
}

func (c *CloudflareClient) IsCloudflareRegistrar(domain string) bool {
	for _, zone := range c.zones {
		if zone.Name == domain {
			for _, ns := range zone.NameServers {
				if strings.HasSuffix(ns, ".ns.cloudflare.com") {
					return true
				}
			}
			return false
		}
	}
	return false
}

func matchesType(recordType string, types []string) bool {
	if len(types) == 0 {
		return true
	}
	for _, t := range types {
		if strings.EqualFold(recordType, t) {
			return true
		}
	}
	return false
}
