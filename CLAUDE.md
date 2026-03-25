# CLAUDE.md - Project Guidelines for Claude

## Project Overview

domains-cli is a CLI tool that lists DNS records from Cloudflare and AWS Route53 with registrar detection and TLS certificate info.

## Building and Running

```bash
go build .
./domains
./domains -c              # include TLS cert info
./domains -f json A CNAME # JSON output, filter by type
```

## Testing

```bash
go test ./...
go vet ./...
```

## Code Style

- Follow standard Go conventions (gofmt, go vet)
- Use `fmt.Errorf("context: %w", err)` for error wrapping
- No emojis in code, commit messages, or documentation
- No comments - write self-documenting code

## Releasing

Releases are automated via GitHub Actions + goreleaser. Version is injected at build time via `-ldflags -X main.version=...` from the git tag. There is no VERSION file.

```bash
git tag v1.x.y
git push origin v1.x.y
```

### Versioning Guidelines

Use semantic versioning (MAJOR.MINOR.PATCH):

- **PATCH** (v1.0.x): Bug fixes, minor enhancements
  - Fixing output formatting
  - Improving error messages
  - Performance improvements
- **MINOR** (v1.x.0): New features, new providers, new output fields
  - Adding a new DNS provider
  - Adding new output columns (e.g., TLS version)
  - Adding new output formats
- **MAJOR** (vX.0.0): Breaking changes
  - Changed CLI flags or arguments
  - Changed output format structure
  - Removed features
