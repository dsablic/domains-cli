# domains-cli

CLI tool to list DNS records from Cloudflare and AWS Route53 with registrar detection.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap dsablic/tap
brew install domains
```

### Go Install

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

# Include TLS certificate info (issuer, expiry, TLS version)
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
| tls_version | Negotiated TLS version (e.g., "TLS 1.3") |
| cert_error | Error if cert lookup failed (e.g., "timeout", "connection refused") |

Certificate lookups are performed for A, AAAA, and CNAME records. Other record types show "n/a".

## Configuration

### Environment Variables

```bash
export CLOUDFLARE_API_TOKEN="your-api-token"
```

AWS credentials use the standard SDK chain (environment variables, `~/.aws/credentials`, IAM roles).

### Config File

Create `~/.config/domains/config.yaml`:

```yaml
cloudflare:
  api_token: "your-api-token"
```

Environment variables take precedence over the config file.

## Registrar Detection

The tool determines domain registrars using:

1. Route53 hosted zone comments containing "Route53 Registrar"
2. WHOIS lookup as fallback for all other domains

## Development

### Releasing

1. Tag the release: `git tag v1.x.y`
2. Push the tag: `git push origin v1.x.y`
3. GitHub Actions builds and publishes the release via goreleaser

## License

MIT License - see [LICENSE](LICENSE) for details.
