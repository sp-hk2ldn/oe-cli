package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"oe-cli/internal/appleads"
)

func RunProductPages(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("product-pages", jsonOut, err)
		return
	}

	action := actionFromArgs(args, "list")
	switch action {
	case "list", "find":
		runProductPagesList(ctx, client, args, jsonOut)
	case "get", "show":
		runProductPagesGet(ctx, client, args, jsonOut)
	case "locales":
		runProductPagesLocales(ctx, client, args, jsonOut)
	case "countries":
		runProductPagesCountries(ctx, client, args, jsonOut)
	case "devices", "device-sizes":
		runProductPagesDevices(ctx, client, args, jsonOut)
	default:
		respondCommandError("product-pages", jsonOut, fmt.Errorf("Unsupported product-pages action: %s. Use: list|get|locales|countries|devices", action))
	}
}

func runProductPagesList(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	adamID, err := requiredIntFlag(args, "--adamId")
	if err != nil {
		respondCommandError("product-pages", jsonOut, err)
		return
	}
	stateFilter := parseStringSet(splitCSVValues(valuesForFlag(args, "--state")), true)
	nameContains := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--nameContains")))

	items, err := client.FetchProductPages(ctx, adamID)
	if err != nil {
		respondCommandError("product-pages", jsonOut, err)
		return
	}
	filtered := make([]appleads.ProductPageSummary, 0, len(items))
	for _, item := range items {
		if len(stateFilter) > 0 {
			if _, ok := stateFilter[strings.ToUpper(strings.TrimSpace(item.State))]; !ok {
				continue
			}
		}
		if nameContains != "" && !strings.Contains(strings.ToLower(item.Name), nameContains) {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })
	respondProductPages(jsonOut, filtered)
}

func runProductPagesGet(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	adamID, err := requiredIntFlag(args, "--adamId")
	if err != nil {
		respondCommandError("product-pages", jsonOut, err)
		return
	}
	productPageID := strings.TrimSpace(valueForFlag(args, "--productPageId"))
	if productPageID == "" {
		respondCommandError("product-pages", jsonOut, fmt.Errorf("Missing required --productPageId <id>"))
		return
	}
	item, err := client.FetchProductPage(ctx, adamID, productPageID)
	if err != nil {
		respondCommandError("product-pages", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(item)
		return
	}
	fmt.Printf("productPageId=%s\n", item.ID)
	fmt.Printf("adamId=%d\n", item.AdamID)
	fmt.Printf("name=%s\n", item.Name)
	fmt.Printf("state=%s\n", item.State)
	if item.DeepLink != nil {
		fmt.Printf("deepLink=%s\n", *item.DeepLink)
	}
}

func runProductPagesLocales(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	adamID, err := requiredIntFlag(args, "--adamId")
	if err != nil {
		respondCommandError("product-pages", jsonOut, err)
		return
	}
	productPageID := strings.TrimSpace(valueForFlag(args, "--productPageId"))
	if productPageID == "" {
		respondCommandError("product-pages", jsonOut, fmt.Errorf("Missing required --productPageId <id>"))
		return
	}
	items, err := client.FetchProductPageLocales(ctx, adamID, productPageID, hasFlag(args, "--expand"))
	if err != nil {
		respondCommandError("product-pages", jsonOut, err)
		return
	}
	sort.Slice(items, func(i, j int) bool {
		return firstNonEmptyString(items[i].LanguageCode, items[i].Language) < firstNonEmptyString(items[j].LanguageCode, items[j].Language)
	})
	if jsonOut {
		printJSON(items)
		return
	}
	fmt.Printf("localeCount=%d\n", len(items))
	for _, item := range items {
		fmt.Printf("%s\t%s\t%s\t%s\n", item.LanguageCode, item.Language, item.ProductPageID, item.AppName)
	}
}

func runProductPagesCountries(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	items, err := client.FetchSupportedCountriesOrRegions(ctx)
	if err != nil {
		respondCommandError("product-pages", jsonOut, err)
		return
	}
	codeFilter := parseStringSet(normalizeUpperValues(splitCSVValues(valuesForFlag(args, "--code"))), true)
	nameContains := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--nameContains")))

	filtered := make([]appleads.CountryOrRegionSummary, 0, len(items))
	for _, item := range items {
		if len(codeFilter) > 0 {
			if _, ok := codeFilter[item.Code]; !ok {
				continue
			}
		}
		if nameContains != "" && !strings.Contains(strings.ToLower(item.DisplayName), nameContains) {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Code < filtered[j].Code })
	if jsonOut {
		printJSON(filtered)
		return
	}
	fmt.Printf("countryCount=%d\n", len(filtered))
	for _, item := range filtered {
		fmt.Printf("%s\t%s\n", item.Code, item.DisplayName)
	}
}

func runProductPagesDevices(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	items, err := client.FetchCreativeAppMappingDevices(ctx)
	if err != nil {
		respondCommandError("product-pages", jsonOut, err)
		return
	}
	classFilter := parseStringSet(normalizeUpperValues(splitCSVValues(valuesForFlag(args, "--deviceClass"))), true)
	nameContains := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--nameContains")))

	filtered := make([]appleads.DeviceSizeMapping, 0, len(items))
	for _, item := range items {
		key := strings.ToUpper(strings.TrimSpace(firstNonEmptyString(item.DeviceClass, item.DisplayName)))
		if len(classFilter) > 0 {
			if _, ok := classFilter[key]; !ok {
				continue
			}
		}
		if nameContains != "" && !strings.Contains(strings.ToLower(item.DisplayName), nameContains) {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.Slice(filtered, func(i, j int) bool {
		return firstNonEmptyString(filtered[i].DeviceClass, filtered[i].DisplayName) < firstNonEmptyString(filtered[j].DeviceClass, filtered[j].DisplayName)
	})
	if jsonOut {
		printJSON(filtered)
		return
	}
	fmt.Printf("deviceCount=%d\n", len(filtered))
	for _, item := range filtered {
		fmt.Printf("%s\t%s\n", item.DeviceClass, item.DisplayName)
	}
}

func respondProductPages(jsonOut bool, items []appleads.ProductPageSummary) {
	if jsonOut {
		printJSON(items)
		return
	}
	fmt.Printf("productPageCount=%d\n", len(items))
	for _, item := range items {
		deepLink := "-"
		if item.DeepLink != nil {
			deepLink = *item.DeepLink
		}
		fmt.Printf("%s\t%s\t%d\t%s\t%s\n", item.ID, item.State, item.AdamID, item.Name, deepLink)
	}
}
