# domains-cli

CLI tool to list DNS records from Cloudflare and AWS Route53 with registrar detection.

## Installation

```bash
go install github.com/dsablic/domains-cli@latest
```

Or download a binary from the [releases page](https://github.com/dsablic/domains-cli/releases).

## Usage

```bash
# List all DNS records
domains

# Filter by record type (case insensitive)
domains A CNAME
domains ns

# Output as JSON
domains -f json
domains --format json A CNAME
```

## Output

TSV (default) or JSON with columns:

| Column | Description |
|--------|-------------|
| domain | Zone/domain name |
| record | Record name (FQDN) |
| value | Record value (IP, hostname, etc.) |
| type | Record type (A, CNAME, NS, etc.) |
| source | DNS provider (cloudflare or route53) |
| registrar | Domain registrar |

## Configuration

### Environment Variables

```bash
export CLOUDFLARE_API_KEY="your-api-key"
export CLOUDFLARE_EMAIL="your-email@example.com"
```

AWS credentials use the standard SDK chain (environment variables, `~/.aws/credentials`, IAM roles).

### Config File

Create `~/.config/domains/config.yaml`:

```yaml
cloudflare:
  api_key: "your-api-key"
  email: "your-email@example.com"
```

Environment variables take precedence over the config file.

## Registrar Detection

The tool determines domain registrars using:

1. Cloudflare nameservers (`*.ns.cloudflare.com`) indicate Cloudflare registration
2. Route53 hosted zone comments containing "Route53 Registrar"
3. WHOIS lookup as fallback

## License

MIT License - see [LICENSE](LICENSE) for details.
