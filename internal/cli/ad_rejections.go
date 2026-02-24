package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"oe-cli/internal/appleads"
)

func RunAdRejections(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("ad-rejections", jsonOut, err)
		return
	}

	action := actionFromArgs(args, "find")
	switch action {
	case "find", "list":
		runAdRejectionsFind(ctx, client, args, jsonOut)
	case "get", "show":
		runAdRejectionsGet(ctx, client, args, jsonOut)
	case "assets":
		runAdRejectionsAssets(ctx, client, args, jsonOut)
	default:
		respondCommandError("ad-rejections", jsonOut, fmt.Errorf("Unsupported ad-rejections action: %s. Use: find|get|assets", action))
	}
}

func runAdRejectionsFind(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	offset := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--offset")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &offset)
		if offset < 0 {
			offset = 0
		}
	}
	limit := 200
	if raw := strings.TrimSpace(valueForFlag(args, "--limit")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &limit)
		if limit <= 0 {
			limit = 200
		}
	}

	conditions := make([]any, 0, 8)
	if rawAdamIDs := splitCSVValues(valuesForFlag(args, "--adamId")); len(rawAdamIDs) > 0 {
		conditions = append(conditions, selectorCondition("adamId", normalizeNonEmptyStrings(rawAdamIDs)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--productPageId")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("productPageId", normalizeNonEmptyStrings(values)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--reasonType")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("reasonType", normalizeUpperValues(values)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--reasonLevel")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("reasonLevel", normalizeUpperValues(values)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--reasonCode")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("reasonCode", normalizeNonEmptyStrings(values)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--countryOrRegion")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("countryOrRegion", normalizeUpperValues(values)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--languageCode")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("languageCode", normalizeNonEmptyStrings(values)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--supplySource")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("supplySource", normalizeUpperValues(values)))
	}

	selector := map[string]any{
		"conditions": conditions,
		"fields":     nil,
		"orderBy":    []any{map[string]any{"field": "id", "sortOrder": "ASCENDING"}},
		"pagination": map[string]any{"offset": offset, "limit": limit},
	}

	items, err := client.FindAdRejections(ctx, selector)
	if err != nil {
		respondCommandError("ad-rejections", jsonOut, err)
		return
	}
	if commentContains := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--commentContains"))); commentContains != "" {
		filtered := make([]appleads.AdRejectionSummary, 0, len(items))
		for _, item := range items {
			if item.Comment != nil && strings.Contains(strings.ToLower(*item.Comment), commentContains) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	respondAdRejections(jsonOut, items)
}

func runAdRejectionsGet(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	reasonID, err := requiredIntFlag(args, "--reasonId")
	if err != nil {
		respondCommandError("ad-rejections", jsonOut, err)
		return
	}
	item, err := client.FetchAdRejection(ctx, reasonID)
	if err != nil {
		respondCommandError("ad-rejections", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(item)
		return
	}
	fmt.Printf("reasonId=%d\n", item.ID)
	fmt.Printf("reasonType=%s reasonLevel=%s reasonCode=%s\n", item.ReasonType, item.ReasonLevel, item.ReasonCode)
	if item.ProductPageID != nil {
		fmt.Printf("productPageId=%s\n", *item.ProductPageID)
	}
	if item.Comment != nil {
		fmt.Printf("comment=%s\n", *item.Comment)
	}
}

func runAdRejectionsAssets(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	adamID, err := requiredIntFlag(args, "--adamId")
	if err != nil {
		respondCommandError("ad-rejections", jsonOut, err)
		return
	}
	offset := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--offset")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &offset)
		if offset < 0 {
			offset = 0
		}
	}
	limit := 200
	if raw := strings.TrimSpace(valueForFlag(args, "--limit")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &limit)
		if limit <= 0 {
			limit = 200
		}
	}

	conditions := make([]any, 0, 5)
	if values := splitCSVValues(valuesForFlag(args, "--assetType")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("assetType", normalizeUpperValues(values)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--orientation")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("orientation", normalizeUpperValues(values)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--appPreviewDevice")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("appPreviewDevice", normalizeNonEmptyStrings(values)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--assetGenId")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("assetGenId", normalizeNonEmptyStrings(values)))
	}

	selector := map[string]any{
		"conditions": conditions,
		"fields":     nil,
		"orderBy":    []any{map[string]any{"field": "assetGenId", "sortOrder": "ASCENDING"}},
		"pagination": map[string]any{"offset": offset, "limit": limit},
	}

	items, err := client.FindAppAssets(ctx, adamID, selector)
	if err != nil {
		respondCommandError("ad-rejections", jsonOut, err)
		return
	}
	if !hasFlag(args, "--includeDeleted") {
		filtered := make([]appleads.AppAssetSummary, 0, len(items))
		for _, item := range items {
			if !item.Deleted {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	sort.Slice(items, func(i, j int) bool {
		left := strings.TrimSpace(firstNonEmptyString(assetID(items[i]), assetURL(items[i])))
		right := strings.TrimSpace(firstNonEmptyString(assetID(items[j]), assetURL(items[j])))
		return left < right
	})
	respondAppAssets(jsonOut, items)
}

func respondAdRejections(jsonOut bool, items []appleads.AdRejectionSummary) {
	if jsonOut {
		printJSON(items)
		return
	}
	fmt.Printf("adRejectionCount=%d\n", len(items))
	for _, item := range items {
		productPageID := "-"
		if item.ProductPageID != nil {
			productPageID = *item.ProductPageID
		}
		comment := "-"
		if item.Comment != nil {
			comment = *item.Comment
		}
		fmt.Printf("%d\t%s\t%s\t%s\t%s\t%s\t%s\n", item.ID, item.ReasonType, item.ReasonLevel, item.ReasonCode, item.CountryOrRegion, productPageID, comment)
	}
}

func respondAppAssets(jsonOut bool, items []appleads.AppAssetSummary) {
	if jsonOut {
		printJSON(items)
		return
	}
	fmt.Printf("assetCount=%d\n", len(items))
	for _, item := range items {
		fmt.Printf("%s\t%s\t%s\t%s\t%t\n", assetID(item), item.AssetType, item.Orientation, assetURL(item), item.Deleted)
	}
}

func assetID(item appleads.AppAssetSummary) string {
	if item.AssetGenID != nil {
		return *item.AssetGenID
	}
	return "-"
}

func assetURL(item appleads.AppAssetSummary) string {
	if item.AssetURL != nil {
		return *item.AssetURL
	}
	if item.AssetVideoURL != nil {
		return *item.AssetVideoURL
	}
	return "-"
}

func normalizeNonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
