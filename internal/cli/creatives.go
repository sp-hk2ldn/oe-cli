package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"oe-cli/internal/appleads"
)

func RunCreatives(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("creatives", jsonOut, err)
		return
	}

	action := actionFromArgs(args, "list")
	switch action {
	case "list":
		runCreativesList(ctx, client, jsonOut)
	case "get", "show":
		runCreativesGet(ctx, client, args, jsonOut)
	case "find":
		runCreativesFind(ctx, client, args, jsonOut)
	case "create":
		runCreativesCreate(ctx, client, args, jsonOut)
	default:
		respondCommandError("creatives", jsonOut, fmt.Errorf("Unsupported creatives action: %s. Use: list|find|get|create", action))
	}
}

func runCreativesList(ctx context.Context, client *appleads.Client, jsonOut bool) {
	items, err := client.FetchCreatives(ctx)
	if err != nil {
		respondCommandError("creatives", jsonOut, err)
		return
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	respondCreatives(jsonOut, items)
}

func runCreativesGet(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	creativeID, err := requiredIntFlag(args, "--creativeId")
	if err != nil {
		respondCommandError("creatives", jsonOut, err)
		return
	}
	item, err := client.FetchCreative(ctx, creativeID)
	if err != nil {
		respondCommandError("creatives", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(item)
		return
	}
	fmt.Printf("creativeId=%d\n", item.ID)
	fmt.Printf("name=%s\n", item.Name)
	fmt.Printf("type=%s state=%s adamId=%d\n", item.Type, item.State, item.AdamID)
	if item.ProductPageID != nil {
		fmt.Printf("productPageId=%s\n", *item.ProductPageID)
	}
}

func runCreativesFind(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
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

	conditions := make([]any, 0, 4)
	if nameContains := strings.TrimSpace(valueForFlag(args, "--nameContains")); nameContains != "" {
		conditions = append(conditions, map[string]any{"field": "name", "operator": "CONTAINS", "values": []string{nameContains}})
	}
	if typeValues := normalizeUpperValues(splitCSVValues(valuesForFlag(args, "--type"))); len(typeValues) > 0 {
		conditions = append(conditions, selectorCondition("type", typeValues))
	}
	if stateValues := normalizeUpperValues(splitCSVValues(valuesForFlag(args, "--state"))); len(stateValues) > 0 {
		conditions = append(conditions, selectorCondition("state", stateValues))
	}
	if rawAdamIDs := splitCSVValues(valuesForFlag(args, "--adamId")); len(rawAdamIDs) > 0 {
		clean := make([]string, 0, len(rawAdamIDs))
		for _, value := range rawAdamIDs {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				clean = append(clean, trimmed)
			}
		}
		if len(clean) > 0 {
			conditions = append(conditions, selectorCondition("adamId", clean))
		}
	}

	selector := map[string]any{
		"conditions": conditions,
		"fields":     nil,
		"orderBy":    []any{map[string]any{"field": "id", "sortOrder": "ASCENDING"}},
		"pagination": map[string]any{"offset": offset, "limit": limit},
	}
	items, err := client.FindCreatives(ctx, selector)
	if err != nil {
		respondCommandError("creatives", jsonOut, err)
		return
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	respondCreatives(jsonOut, items)
}

func runCreativesCreate(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	adamID, err := requiredIntFlag(args, "--adamId")
	if err != nil {
		respondCommandError("creatives", jsonOut, err)
		return
	}
	name := strings.TrimSpace(valueForFlag(args, "--name"))
	if name == "" {
		respondCommandError("creatives", jsonOut, fmt.Errorf("Missing required --name <creative name>"))
		return
	}
	creativeType := strings.ToUpper(firstNonEmptyString(valueForFlag(args, "--type"), "CUSTOM_PRODUCT_PAGE"))
	var productPageID *string
	if raw := strings.TrimSpace(valueForFlag(args, "--productPageId")); raw != "" {
		productPageID = &raw
	}
	if creativeType == "CUSTOM_PRODUCT_PAGE" && productPageID == nil {
		respondCommandError("creatives", jsonOut, fmt.Errorf("--productPageId is required when --type=CUSTOM_PRODUCT_PAGE"))
		return
	}

	item, err := client.CreateCreative(ctx, adamID, name, creativeType, productPageID)
	if err != nil {
		respondCommandError("creatives", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "action": "create", "creative": item})
		return
	}
	fmt.Printf("ok action=create id=%d type=%s state=%s name=%s\n", item.ID, item.Type, item.State, item.Name)
}

func respondCreatives(jsonOut bool, items []appleads.CreativeSummary) {
	if jsonOut {
		printJSON(items)
		return
	}
	fmt.Printf("creativeCount=%d\n", len(items))
	for _, item := range items {
		productPageID := "-"
		if item.ProductPageID != nil {
			productPageID = *item.ProductPageID
		}
		fmt.Printf("%d\t%s\t%s\t%d\t%s\t%s\n", item.ID, item.State, item.Type, item.AdamID, productPageID, item.Name)
	}
}
