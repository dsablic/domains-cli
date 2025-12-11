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
domains-cli

# Filter by record type (case insensitive)
domains-cli A CNAME
domains-cli ns

# Output as JSON
domains-cli -f json
domains-cli --format json A CNAME

# Include TLS certificate info (issuer, expiry)
domains-cli --cert
domains-cli -c -f json A CNAME
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

With `--cert` flag, additional columns:

| Column | Description |
|--------|-------------|
| cert_issuer | Certificate issuer (e.g., "Let's Encrypt", "Amazon") |
| cert_expires | Certificate expiry date (YYYY-MM-DD) |
| cert_error | Error if cert lookup failed (e.g., "timeout", "connection refused") |

Certificate lookups are performed for A, AAAA, and CNAME records. Other record types show "n/a".

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

## Development

### Setup

After cloning, install git hooks:

```bash
./scripts/install-hooks.sh
```

### Releasing

1. Update `VERSION` file (e.g., `0.1.0` → `0.2.0`)
2. Commit the change
3. The post-commit hook automatically creates and pushes the tag
4. GitHub Actions builds and publishes the release

## License

MIT License - see [LICENSE](LICENSE) for details.
