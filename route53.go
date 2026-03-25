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
	client  *route53.Client
	zones   []types.HostedZone
	zoneMap map[string]types.HostedZone
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
	paginator := route53.NewListHostedZonesPaginator(c.client, &route53.ListHostedZonesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list route53 hosted zones: %w", err)
		}
		c.zones = append(c.zones, page.HostedZones...)
	}

	c.zoneMap = make(map[string]types.HostedZone, len(c.zones))
	for _, zone := range c.zones {
		domain := strings.TrimSuffix(*zone.Name, ".")
		c.zoneMap[domain] = zone
	}

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

	var allRecordSets []types.ResourceRecordSet
	input := &route53.ListResourceRecordSetsInput{HostedZoneId: zone.Id}
	for {
		out, err := c.client.ListResourceRecordSets(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list record sets for zone %s: %w", domain, err)
		}
		allRecordSets = append(allRecordSets, out.ResourceRecordSets...)
		if !out.IsTruncated {
			break
		}
		input.StartRecordName = out.NextRecordName
		input.StartRecordType = out.NextRecordType
		input.StartRecordIdentifier = out.NextRecordIdentifier
	}

	for _, rs := range allRecordSets {
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
	zone, ok := c.zoneMap[domain]
	if !ok {
		return false
	}
	if zone.Config != nil && zone.Config.Comment != nil {
		return strings.Contains(*zone.Config.Comment, "Route53 Registrar")
	}
	return false
}
