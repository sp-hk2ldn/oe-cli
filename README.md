# oe-cli (Go)

Lightweight Apple Search Ads CLI in Go.

## Command Surface
- `oe-ads status`
- `oe-ads campaigns [list|find|create|pause|activate|delete|update-budget|set-budget|report] [flags] [--json]`
- `oe-ads adgroups [list|find|create|pause|activate|delete|report] [flags] [--json]`
- `oe-ads ads [list|find|get|create|update|pause|activate|delete] [flags] [--json]`
- `oe-ads creatives [list|find|get|create] [flags] [--json]`
- `oe-ads product-pages [list|get|locales|countries|devices] [flags] [--json]`
- `oe-ads apps [search|get|localized-details|eligibility] [flags] [--json]`
- `oe-ads geo [search|get] [flags] [--json]`
- `oe-ads ad-rejections [find|get|assets] [flags] [--json]`
- `oe-ads keywords [list|find|report|add|pause|activate|remove|rebid|pause-by-text] --campaignId <id> --adGroupId <id> [flags] [--json]`
- `oe-ads searchterms report --campaignId <id> [--adGroupId <id>] --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--minTaps N] [--minSpend X] [--json]`
- `oe-ads negatives [list|add|remove|pause|activate] --campaignId <id> [--adGroupId <id>] [--negativeKeywordId <id> ...] [--text <kw> ...] [--matchType EXACT|BROAD] [--json]`
- `oe-ads sov-report --adamId <id> [--country GB,US] [--dateRange LAST_4_WEEKS] [--out reports/sov] [--json]`
- `oe-ads reports [list|get|download] [--reportId <id>] [--state COMPLETED] [--nameContains text] [--limit N] [--out reports/custom/id.csv] [--json]`

Full command and flag docs: [docs/COMMANDS.md](docs/COMMANDS.md)
Open source release checklist: [docs/OPEN_SOURCE_RELEASE_CHECKLIST.md](docs/OPEN_SOURCE_RELEASE_CHECKLIST.md)
Contributor guide: [CONTRIBUTING.md](CONTRIBUTING.md)
Security policy: [SECURITY.md](SECURITY.md)
Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## Credentials
The CLI supports either:
- `OE_ADS_CREDENTIALS_JSON` with JSON fields:
  - required: `clientId`, `teamId`, `keyId`, `privateKey`
  - optional: `orgId`, `popularityAdamId`, `popularityAdGroupId`, `popularityWebCookie`, `popularityXsrfToken`
- split env vars:
  - `OE_ADS_CLIENT_ID`, `OE_ADS_TEAM_ID`, `OE_ADS_KEY_ID`, `OE_ADS_PRIVATE_KEY`

## Build / Run
```bash
go build ./...
go run ./cmd/oe-ads --help
go run ./cmd/oe-ads status
```

## Tests
Golden command stability tests are in:
- `cmd/oe-ads/main_golden_test.go`
- `cmd/oe-ads/testdata/golden/*.json`

Run:
```bash
go test ./...
```

## Swift vs Go Parity Check
Run side-by-side parity against the Swift CLI:
```bash
./scripts/parity_check.sh
```

When credentials are present, you can deepen live checks by setting:
- `OE_ADS_PARITY_CAMPAIGN_ID`
- `OE_ADS_PARITY_ADGROUP_ID`

## Notes
- Auth flow: ES256 client-secret JWT, Apple token exchange, org discovery via `/api/v5/me`.
- Includes custom-report workflows for SOV and generic report download.
