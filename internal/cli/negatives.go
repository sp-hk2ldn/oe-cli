package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"oe-cli/internal/appleads"
)

func RunNegatives(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("negatives", jsonOut, err)
		return
	}
	action := actionFromArgs(args, "list")
	switch action {
	case "list":
		runNegativesList(ctx, client, args, jsonOut)
	case "add":
		runNegativesAdd(ctx, client, args, jsonOut)
	case "remove", "delete":
		runNegativesRemove(ctx, client, args, jsonOut)
	case "pause", "activate":
		runNegativesUpdateStatus(ctx, client, args, action, jsonOut)
	default:
		respondCommandError("negatives", jsonOut, fmt.Errorf("Unsupported negatives action: %s. Use: list|add|remove|pause|activate", action))
	}
}

func runNegativesList(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	adGroupID := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--adGroupId")); raw != "" {
		if _, err := fmt.Sscanf(raw, "%d", &adGroupID); err != nil || adGroupID <= 0 {
			respondCommandError("negatives", jsonOut, fmt.Errorf("Invalid --adGroupId %q", raw))
			return
		}
	}
	if adGroupID > 0 {
		campaignID, err := requiredIntFlag(args, "--campaignId")
		if err != nil {
			respondCommandError("negatives", jsonOut, fmt.Errorf("--campaignId is required when using --adGroupId"))
			return
		}
		negatives, err := client.FetchNegativeKeywords(ctx, campaignID, adGroupID)
		if err != nil {
			respondCommandError("negatives", jsonOut, err)
			return
		}
		sort.Slice(negatives, func(i, j int) bool { return negatives[i].ID < negatives[j].ID })
		if jsonOut {
			printJSON(negatives)
			return
		}
		fmt.Printf("scope=adgroup campaignId=%d adGroupId=%d negativeCount=%d\n", campaignID, adGroupID, len(negatives))
		for _, item := range negatives {
			fmt.Printf("%d\t%s\t%s\t%s\n", item.ID, item.Status, item.MatchType, item.Text)
		}
		return
	}

	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("negatives", jsonOut, err)
		return
	}
	negatives, err := client.FetchCampaignNegativeKeywords(ctx, campaignID)
	if err != nil {
		respondCommandError("negatives", jsonOut, err)
		return
	}
	sort.Slice(negatives, func(i, j int) bool { return negatives[i].ID < negatives[j].ID })
	if jsonOut {
		printJSON(negatives)
		return
	}
	fmt.Printf("scope=campaign campaignId=%d negativeCount=%d\n", campaignID, len(negatives))
	for _, item := range negatives {
		fmt.Printf("%d\t%s\t%s\t%s\n", item.ID, item.Status, item.MatchType, item.Text)
	}
}

func runNegativesAdd(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	texts := make([]string, 0, 4)
	for _, text := range valuesForFlag(args, "--text") {
		trimmed := strings.TrimSpace(text)
		if trimmed != "" {
			texts = append(texts, trimmed)
		}
	}
	if len(texts) == 0 {
		respondCommandError("negatives", jsonOut, fmt.Errorf("Provide at least one --text <negative keyword>"))
		return
	}
	matchType := strings.ToUpper(firstNonEmptyString(valueForFlag(args, "--matchType"), "EXACT"))
	payload := make([]appleads.NegativeKeywordSummary, 0, len(texts))
	for _, text := range texts {
		payload = append(payload, appleads.NegativeKeywordSummary{Text: text, MatchType: matchType})
	}

	adGroupID := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--adGroupId")); raw != "" {
		if _, err := fmt.Sscanf(raw, "%d", &adGroupID); err != nil || adGroupID <= 0 {
			respondCommandError("negatives", jsonOut, fmt.Errorf("Invalid --adGroupId %q", raw))
			return
		}
	}
	if adGroupID > 0 {
		campaignID, err := requiredIntFlag(args, "--campaignId")
		if err != nil {
			respondCommandError("negatives", jsonOut, fmt.Errorf("--campaignId is required when using --adGroupId"))
			return
		}
		if err := client.AddNegativeKeywords(ctx, campaignID, adGroupID, payload); err != nil {
			respondCommandError("negatives", jsonOut, err)
			return
		}
		if jsonOut {
			printJSON(map[string]any{"ok": true, "scope": "adgroup", "campaignId": campaignID, "adGroupId": adGroupID, "added": len(texts), "matchType": matchType})
			return
		}
		fmt.Printf("ok scope=adgroup campaignId=%d adGroupId=%d added=%d matchType=%s\n", campaignID, adGroupID, len(texts), matchType)
		return
	}

	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("negatives", jsonOut, err)
		return
	}
	if err := client.AddCampaignNegativeKeywords(ctx, campaignID, payload); err != nil {
		respondCommandError("negatives", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "scope": "campaign", "campaignId": campaignID, "added": len(texts), "matchType": matchType})
		return
	}
	fmt.Printf("ok scope=campaign campaignId=%d added=%d matchType=%s\n", campaignID, len(texts), matchType)
}

func runNegativesRemove(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	keywordIDs := parseIntFlagSet(args, "--negativeKeywordId")
	textFilters := parseStringSet(valuesForFlag(args, "--text"), false)
	if len(keywordIDs) == 0 && len(textFilters) == 0 {
		respondCommandError("negatives", jsonOut, fmt.Errorf("Provide --negativeKeywordId <id> (repeatable) and/or --text <keyword> (repeatable)"))
		return
	}

	adGroupID := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--adGroupId")); raw != "" {
		if _, err := fmt.Sscanf(raw, "%d", &adGroupID); err != nil || adGroupID <= 0 {
			respondCommandError("negatives", jsonOut, fmt.Errorf("Invalid --adGroupId %q", raw))
			return
		}
	}
	if adGroupID > 0 {
		campaignID, err := requiredIntFlag(args, "--campaignId")
		if err != nil {
			respondCommandError("negatives", jsonOut, fmt.Errorf("--campaignId is required when using --adGroupId"))
			return
		}
		negatives, err := client.FetchNegativeKeywords(ctx, campaignID, adGroupID)
		if err != nil {
			respondCommandError("negatives", jsonOut, err)
			return
		}
		targetIDs := resolveNegativeTargets(keywordIDs, textFilters, negatives)
		for _, targetID := range targetIDs {
			if err := client.DeleteNegativeKeyword(ctx, campaignID, adGroupID, targetID); err != nil {
				respondCommandError("negatives", jsonOut, err)
				return
			}
		}
		if jsonOut {
			printJSON(map[string]any{"ok": true, "scope": "adgroup", "campaignId": campaignID, "adGroupId": adGroupID, "removed": len(targetIDs)})
			return
		}
		fmt.Printf("ok scope=adgroup campaignId=%d adGroupId=%d removed=%d\n", campaignID, adGroupID, len(targetIDs))
		return
	}

	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("negatives", jsonOut, err)
		return
	}
	negatives, err := client.FetchCampaignNegativeKeywords(ctx, campaignID)
	if err != nil {
		respondCommandError("negatives", jsonOut, err)
		return
	}
	targetIDs := resolveNegativeTargets(keywordIDs, textFilters, negatives)
	for _, targetID := range targetIDs {
		if err := client.DeleteCampaignNegativeKeyword(ctx, campaignID, targetID); err != nil {
			respondCommandError("negatives", jsonOut, err)
			return
		}
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "scope": "campaign", "campaignId": campaignID, "removed": len(targetIDs)})
		return
	}
	fmt.Printf("ok scope=campaign campaignId=%d removed=%d\n", campaignID, len(targetIDs))
}

func runNegativesUpdateStatus(ctx context.Context, client *appleads.Client, args []string, action string, jsonOut bool) {
	keywordIDs := parseIntFlagSet(args, "--negativeKeywordId")
	textFilters := parseStringSet(valuesForFlag(args, "--text"), false)
	if len(keywordIDs) == 0 && len(textFilters) == 0 {
		respondCommandError("negatives", jsonOut, fmt.Errorf("Provide --negativeKeywordId <id> (repeatable) and/or --text <keyword> (repeatable)"))
		return
	}

	status := "PAUSED"
	if action == "activate" {
		status = "ACTIVE"
	}

	adGroupID := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--adGroupId")); raw != "" {
		if _, err := fmt.Sscanf(raw, "%d", &adGroupID); err != nil || adGroupID <= 0 {
			respondCommandError("negatives", jsonOut, fmt.Errorf("Invalid --adGroupId %q", raw))
			return
		}
	}
	if adGroupID > 0 {
		campaignID, err := requiredIntFlag(args, "--campaignId")
		if err != nil {
			respondCommandError("negatives", jsonOut, fmt.Errorf("--campaignId is required when using --adGroupId"))
			return
		}
		negatives, err := client.FetchNegativeKeywords(ctx, campaignID, adGroupID)
		if err != nil {
			respondCommandError("negatives", jsonOut, err)
			return
		}
		targetIDs := resolveNegativeTargets(keywordIDs, textFilters, negatives)
		for _, targetID := range targetIDs {
			if err := client.UpdateNegativeKeywordStatus(ctx, campaignID, adGroupID, targetID, status); err != nil {
				respondCommandError("negatives", jsonOut, err)
				return
			}
		}
		if jsonOut {
			printJSON(map[string]any{"ok": true, "scope": "adgroup", "campaignId": campaignID, "adGroupId": adGroupID, "action": action, "status": status, "affected": len(targetIDs)})
			return
		}
		fmt.Printf("ok scope=adgroup campaignId=%d adGroupId=%d action=%s status=%s affected=%d\n", campaignID, adGroupID, action, status, len(targetIDs))
		return
	}

	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("negatives", jsonOut, err)
		return
	}
	negatives, err := client.FetchCampaignNegativeKeywords(ctx, campaignID)
	if err != nil {
		respondCommandError("negatives", jsonOut, err)
		return
	}
	targetIDs := resolveNegativeTargets(keywordIDs, textFilters, negatives)
	for _, targetID := range targetIDs {
		if err := client.UpdateCampaignNegativeKeywordStatus(ctx, campaignID, targetID, status); err != nil {
			respondCommandError("negatives", jsonOut, err)
			return
		}
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "scope": "campaign", "campaignId": campaignID, "action": action, "status": status, "affected": len(targetIDs)})
		return
	}
	fmt.Printf("ok scope=campaign campaignId=%d action=%s status=%s affected=%d\n", campaignID, action, status, len(targetIDs))
}

func resolveNegativeTargets(keywordIDs map[int]struct{}, textFilters map[string]struct{}, negatives []appleads.NegativeKeywordSummary) []int {
	ids := map[int]struct{}{}
	for id := range keywordIDs {
		ids[id] = struct{}{}
	}
	if len(textFilters) > 0 {
		for _, negative := range negatives {
			if _, ok := textFilters[strings.ToLower(strings.TrimSpace(negative.Text))]; ok {
				ids[negative.ID] = struct{}{}
			}
		}
	}
	resolved := make([]int, 0, len(ids))
	for id := range ids {
		resolved = append(resolved, id)
	}
	sort.Ints(resolved)
	return resolved
}
