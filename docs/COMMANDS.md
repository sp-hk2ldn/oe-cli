# Command Reference

All commands support `--json` unless otherwise noted.

## status
- `oe-ads status`

## campaigns
- `oe-ads campaigns list`
- `oe-ads campaigns find [--campaignId <id> ...] [--status ENABLED,PAUSED] [--nameContains text]`
- `oe-ads campaigns create --name <name> --budgetAmount <number> [--budgetCurrency GBP] [--budgetType DAILY] [--status ENABLED] [--adamId <id>] [--countries GB,US] [--startTime RFC3339] [--endTime RFC3339]`
- `oe-ads campaigns pause --campaignId <id>`
- `oe-ads campaigns activate --campaignId <id>`
- `oe-ads campaigns delete --campaignId <id>`
- `oe-ads campaigns update-budget --campaignId <id> --budgetAmount <number> [--budgetCurrency GBP]`
- `oe-ads campaigns set-budget --campaignId <id> --budgetAmount <number> [--budgetCurrency GBP]`
- `oe-ads campaigns report --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--nameIncludes text] [--nameExcludes text] [--includePaused]`

## adgroups
- `oe-ads adgroups list --campaignId <id>`
- `oe-ads adgroups find --campaignId <id> [--adGroupId <id> ...] [--status ENABLED,PAUSED] [--nameContains text]`
- `oe-ads adgroups create --campaignId <id> --name <name> --defaultBid <number> [--currency GBP] [--status ENABLED] [--automatedKeywordsOptIn]`
- `oe-ads adgroups pause --campaignId <id> --adGroupId <id>`
- `oe-ads adgroups activate --campaignId <id> --adGroupId <id>`
- `oe-ads adgroups delete --campaignId <id> --adGroupId <id>`
- `oe-ads adgroups report --campaignId <id> --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--adGroupId <id>]`

## ads
- `oe-ads ads list --campaignId <id> --adGroupId <id>`
- `oe-ads ads find [--campaignId <id>] [--adGroupId <id>] [--status ENABLED,PAUSED] [--creativeType CUSTOM_PRODUCT_PAGE,DEFAULT_PRODUCT_PAGE] [--nameContains text] [--offset N] [--limit N]`
- `oe-ads ads get --campaignId <id> --adGroupId <id> --adId <id>`
- `oe-ads ads create --campaignId <id> --adGroupId <id> --creativeId <id> [--name text] [--status ENABLED|PAUSED]`
- `oe-ads ads update --campaignId <id> --adGroupId <id> --adId <id> [--name text] [--status ENABLED|PAUSED]`
- `oe-ads ads pause --campaignId <id> --adGroupId <id> --adId <id>`
- `oe-ads ads activate --campaignId <id> --adGroupId <id> --adId <id>`
- `oe-ads ads delete --campaignId <id> --adGroupId <id> --adId <id>`

## creatives
- `oe-ads creatives list`
- `oe-ads creatives find [--nameContains text] [--type CUSTOM_PRODUCT_PAGE,DEFAULT_PRODUCT_PAGE,CREATIVE_SET] [--state VALID,INVALID] [--adamId <id> ...] [--offset N] [--limit N]`
- `oe-ads creatives get --creativeId <id>`
- `oe-ads creatives create --adamId <id> --name <creative name> [--type CUSTOM_PRODUCT_PAGE|DEFAULT_PRODUCT_PAGE] [--productPageId <uuid>]`

Notes:
- `--productPageId` is required for `CUSTOM_PRODUCT_PAGE` creates.

## product-pages
- `oe-ads product-pages list --adamId <id> [--state VISIBLE,HIDDEN] [--nameContains text]`
- `oe-ads product-pages get --adamId <id> --productPageId <id>`
- `oe-ads product-pages locales --adamId <id> --productPageId <id> [--expand]`
- `oe-ads product-pages countries [--code GB,US] [--nameContains text]`
- `oe-ads product-pages devices [--deviceClass IPHONE,IPAD] [--nameContains text]`

## apps
- `oe-ads apps search --query <text> [--returnOwnedApps] [--limit N] [--offset N]`
- `oe-ads apps get --adamId <id>`
- `oe-ads apps localized-details --adamId <id>`
- `oe-ads apps eligibility [--adamId <id> ...] [--countryOrRegion GB,US] [--supplySource APPSTORE_SEARCH_RESULTS] [--state ELIGIBLE,INELIGIBLE] [--eligible true|false] [--appNameContains text] [--offset N] [--limit N]`

## geo
- `oe-ads geo search --query <text> [--countryCode GB] [--entity COUNTRY|ADMIN_AREA|LOCALITY] [--limit N]`
- `oe-ads geo get --geoId <id>`

## ad-rejections
- `oe-ads ad-rejections find [--adamId <id> ...] [--productPageId <id> ...] [--reasonType <value> ...] [--reasonLevel <value> ...] [--reasonCode <value> ...] [--countryOrRegion GB,US] [--languageCode en-GB] [--supplySource APPSTORE_SEARCH_RESULTS] [--commentContains text] [--offset N] [--limit N]`
- `oe-ads ad-rejections get --reasonId <id>`
- `oe-ads ad-rejections assets --adamId <id> [--assetType APP_PREVIEW,SCREENSHOT] [--orientation LANDSCAPE,PORTRAIT] [--appPreviewDevice <value> ...] [--assetGenId <value> ...] [--includeDeleted] [--offset N] [--limit N]`

## keywords
- `oe-ads keywords list --campaignId <id> --adGroupId <id>`
- `oe-ads keywords find --campaignId <id> --adGroupId <id> [--keywordId <id> ...] [--text <exactText> ...] [--textContains partial] [--status ACTIVE,PAUSED] [--matchType BROAD,EXACT]`
- `oe-ads keywords report --campaignId <id> --adGroupId <id> --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--minTaps N] [--minSpend X] [--keywordId <id> ...] [--text <exactText> ...] [--textContains partial] [--status ACTIVE,PAUSED] [--matchType BROAD,EXACT]`
- `oe-ads keywords add --campaignId <id> --adGroupId <id> --text <keyword> ... [--matchType BROAD|EXACT] [--status ACTIVE|PAUSED] [--bidAmount N] [--currency GBP]`
- `oe-ads keywords add --campaignId <id> --adGroupId <id> --file <csvOrJsonFile> [--matchType BROAD|EXACT] [--status ACTIVE|PAUSED] [--currency GBP]`
- `oe-ads keywords pause --campaignId <id> --adGroupId <id> (--keywordId <id> ... | --text <exactText> ...)`
- `oe-ads keywords activate --campaignId <id> --adGroupId <id> (--keywordId <id> ... | --text <exactText> ...)`
- `oe-ads keywords remove --campaignId <id> --adGroupId <id> (--keywordId <id> ... | --text <exactText> ...)`
- `oe-ads keywords rebid --campaignId <id> --adGroupId <id> --bidAmount <number> [--currency GBP] (--keywordId <id> ... | --text <exactText> ...)`
- `oe-ads keywords pause-by-text --campaignId <id> --adGroupId <id> --text <exactText> ...`

## searchterms
- `oe-ads searchterms report --campaignId <id> --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--adGroupId <id>] [--minTaps N] [--minSpend X]`

## negatives
- `oe-ads negatives list --campaignId <id> [--adGroupId <id>]`
- `oe-ads negatives add --campaignId <id> [--adGroupId <id>] --text <keyword> ... [--matchType EXACT|BROAD]`
- `oe-ads negatives remove --campaignId <id> [--adGroupId <id>] (--negativeKeywordId <id> ... | --text <exactText> ...)`
- `oe-ads negatives pause --campaignId <id> [--adGroupId <id>] (--negativeKeywordId <id> ... | --text <exactText> ...)`
- `oe-ads negatives activate --campaignId <id> [--adGroupId <id>] (--negativeKeywordId <id> ... | --text <exactText> ...)`

## sov-report
- `oe-ads sov-report --adamId <id> [--country GB,US] [--dateRange LAST_4_WEEKS] [--name report_name] [--out reports/sov]`
- `--appId` is accepted as an alias for `--adamId`.

Outputs:
- raw CSV
- normalized JSON
- decision table JSON

## reports (Custom Reports)
- `oe-ads reports list [--state COMPLETED,FAILED] [--nameContains text] [--limit N]`
- `oe-ads reports get --reportId <id>`
- `oe-ads reports download --reportId <id> [--out reports/custom/<id>.csv]`

## Useful examples
```bash
# Find paused campaigns
oe-ads campaigns find --status PAUSED --json

# Daily keyword report for one ad group
oe-ads keywords report \
  --campaignId 123 --adGroupId 456 \
  --startDate 2026-02-01 --endDate 2026-02-07 \
  --minTaps 5 --json

# Pause negatives by text at campaign level
oe-ads negatives pause --campaignId 123 --text "free" --text "cheap" --json

# Download a completed custom report
oe-ads reports download --reportId 987654 --out reports/custom/987654.csv --json
```
