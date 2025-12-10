package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

type Route53Client struct {
	client *route53.Client
	zones  []types.HostedZone
}

func NewRoute53Client(ctx context.Context) (*Route53Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	return &Route53Client{
		client: route53.NewFromConfig(cfg),
	}, nil
}

func (c *Route53Client) FetchRecords(ctx context.Context, recordTypes []string) ([]Record, error) {
	zonesOut, err := c.client.ListHostedZones(ctx, &route53.ListHostedZonesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list route53 hosted zones: %w", err)
	}
	c.zones = zonesOut.HostedZones

	var records []Record
	for _, zone := range c.zones {
		zoneRecords, err := c.fetchZoneRecords(ctx, zone, recordTypes)
		if err != nil {
			return nil, err
		}
		records = append(records, zoneRecords...)
	}

	return records, nil
}

func (c *Route53Client) fetchZoneRecords(ctx context.Context, zone types.HostedZone, recordTypes []string) ([]Record, error) {
	var records []Record
	domain := strings.TrimSuffix(*zone.Name, ".")

	recordSets, err := c.client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
		HostedZoneId: zone.Id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list record sets for zone %s: %w", domain, err)
	}

	for _, rs := range recordSets.ResourceRecordSets {
		if !matchesType(string(rs.Type), recordTypes) {
			continue
		}

		values := recordValues(rs)
		for _, value := range values {
			records = append(records, Record{
				Domain: domain,
				Name:   strings.TrimSuffix(*rs.Name, "."),
				Value:  value,
				Type:   string(rs.Type),
				Source: "route53",
			})
		}
	}

	return records, nil
}

func recordValues(rs types.ResourceRecordSet) []string {
	if rs.AliasTarget != nil {
		return []string{strings.TrimSuffix(*rs.AliasTarget.DNSName, ".")}
	}

	var values []string
	for _, rr := range rs.ResourceRecords {
		values = append(values, strings.TrimSuffix(*rr.Value, "."))
	}
	return values
}

func (c *Route53Client) IsRoute53Registrar(domain string) bool {
	for _, zone := range c.zones {
		zoneDomain := strings.TrimSuffix(*zone.Name, ".")
		if zoneDomain == domain {
			comment := ""
			if zone.Config != nil && zone.Config.Comment != nil {
				comment = *zone.Config.Comment
			}
			return strings.Contains(comment, "Route53 Registrar")
		}
	}
	return false
}

func (c *Route53Client) HasZone(domain string) bool {
	for _, zone := range c.zones {
		if strings.TrimSuffix(*zone.Name, ".") == domain {
			return true
		}
	}
	return false
}
