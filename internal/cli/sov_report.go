package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"oe-cli/internal/appleads"
)

type sovOptions struct {
	adamID     string
	countries  []string
	dateRange  string
	name       string
	outputRoot string
	jsonOut    bool
}

func RunSovReport(ctx context.Context, client *appleads.Client, args []string) {
	jsonOut := hasFlag(args, "--json")
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("sov-report", jsonOut, err)
		return
	}
	options, err := parseSovOptions(args)
	if err != nil {
		respondCommandError("sov-report", jsonOut, err)
		return
	}

	report, err := createSovReportWithRetry(ctx, client, options)
	if err != nil {
		respondCommandError("sov-report", jsonOut, err)
		return
	}

	deadline := time.Now().Add(120 * time.Second)
	for strings.ToUpper(report.State) != "COMPLETED" && strings.ToUpper(report.State) != "FAILED" && time.Now().Before(deadline) {
		time.Sleep(4 * time.Second)
		report, err = client.FetchImpressionShareReport(ctx, report.ID)
		if err != nil {
			if isRateLimitedError(err) {
				continue
			}
			respondCommandError("sov-report", jsonOut, err)
			return
		}
	}
	if strings.ToUpper(report.State) != "COMPLETED" {
		respondCommandError("sov-report", jsonOut, fmt.Errorf("Report did not complete in time. state=%s", report.State))
		return
	}
	if report.DownloadURI == nil || strings.TrimSpace(*report.DownloadURI) == "" {
		respondCommandError("sov-report", jsonOut, fmt.Errorf("Completed report missing downloadUri"))
		return
	}

	csvData, err := downloadSovReportWithRetry(ctx, client, *report.DownloadURI)
	if err != nil {
		respondCommandError("sov-report", jsonOut, err)
		return
	}
	csvText := string(csvData)
	normalizedRows := parseAndNormalizeSOV(csvText)
	decisionTable := buildSOVDecisionTable(normalizedRows)

	day := time.Now().UTC().Format("2006-01-02")
	outDir := filepath.Join(options.outputRoot, options.adamID, day)
	if err := os.MkdirAll(outDir, 0o700); err != nil {
		respondCommandError("sov-report", jsonOut, err)
		return
	}

	csvPath := filepath.Join(outDir, "sov-report.csv")
	if err := os.WriteFile(csvPath, csvData, 0o600); err != nil {
		respondCommandError("sov-report", jsonOut, err)
		return
	}

	normalizedPath := filepath.Join(outDir, "sov-report.normalized.json")
	normalizedPayload := map[string]any{
		"reportId":  report.ID,
		"state":     report.State,
		"adamId":    options.adamID,
		"countries": options.countries,
		"dateRange": options.dateRange,
		"rows":      normalizedRows,
	}
	normalizedData, err := json.MarshalIndent(normalizedPayload, "", "  ")
	if err != nil {
		respondCommandError("sov-report", jsonOut, err)
		return
	}
	if err := os.WriteFile(normalizedPath, normalizedData, 0o600); err != nil {
		respondCommandError("sov-report", jsonOut, err)
		return
	}

	decisionPath := filepath.Join(outDir, "sov-decision-table.json")
	decisionData, err := json.MarshalIndent(decisionTable, "", "  ")
	if err != nil {
		respondCommandError("sov-report", jsonOut, err)
		return
	}
	if err := os.WriteFile(decisionPath, decisionData, 0o600); err != nil {
		respondCommandError("sov-report", jsonOut, err)
		return
	}

	if jsonOut {
		printJSON(map[string]any{
			"ok":                 true,
			"reportId":           report.ID,
			"state":              report.State,
			"csvPath":            csvPath,
			"normalizedJsonPath": normalizedPath,
			"decisionTablePath":  decisionPath,
			"rowCount":           len(normalizedRows),
		})
		return
	}
	fmt.Printf("sovReportId=%d\n", report.ID)
	fmt.Printf("state=%s\n", report.State)
	fmt.Printf("rowCount=%d\n", len(normalizedRows))
	fmt.Printf("csvPath=%s\n", csvPath)
	fmt.Printf("normalizedJsonPath=%s\n", normalizedPath)
	fmt.Printf("decisionTablePath=%s\n", decisionPath)
}

func createSovReportWithRetry(ctx context.Context, client *appleads.Client, options *sovOptions) (*appleads.CustomReport, error) {
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		report, err := client.CreateImpressionShareReport(
			ctx,
			options.name,
			"",
			"",
			options.dateRange,
			"WEEKLY",
			options.countries,
			[]string{options.adamID},
			nil,
		)
		if err == nil {
			return report, nil
		}
		lastErr = err
		if !isRateLimitedError(err) {
			return nil, err
		}
		time.Sleep(rateLimitBackoff(attempt))
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("Failed to create SOV report")
}

func downloadSovReportWithRetry(ctx context.Context, client *appleads.Client, downloadURI string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		data, err := client.DownloadCustomReport(ctx, downloadURI)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if !isRateLimitedError(err) {
			return nil, err
		}
		time.Sleep(rateLimitBackoff(attempt))
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("Failed to download SOV report")
}

func isRateLimitedError(err error) bool {
	var apiErr *appleads.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 429
}

func rateLimitBackoff(attempt int) time.Duration {
	switch attempt {
	case 0:
		return 2 * time.Second
	case 1:
		return 5 * time.Second
	case 2:
		return 10 * time.Second
	default:
		return 15 * time.Second
	}
}

func parseSovOptions(args []string) (*sovOptions, error) {
	adamID := firstNonEmptyString(valueForFlag(args, "--adamId"), valueForFlag(args, "--appId"))
	if strings.TrimSpace(adamID) == "" {
		return nil, fmt.Errorf("Missing required --adamId <id>")
	}
	countries := []string{}
	for _, raw := range strings.Split(valueForFlag(args, "--country"), ",") {
		country := strings.ToUpper(strings.TrimSpace(raw))
		if country != "" {
			countries = append(countries, country)
		}
	}
	return &sovOptions{
		adamID:     strings.TrimSpace(adamID),
		countries:  countries,
		dateRange:  strings.ToUpper(firstNonEmptyString(valueForFlag(args, "--dateRange"), "LAST_4_WEEKS")),
		name:       strings.TrimSpace(valueForFlag(args, "--name")),
		outputRoot: firstNonEmptyString(valueForFlag(args, "--out"), "reports/sov"),
		jsonOut:    hasFlag(args, "--json"),
	}, nil
}

func parseAndNormalizeSOV(csvText string) []map[string]any {
	lines := splitLines(csvText)
	if len(lines) == 0 {
		return []map[string]any{}
	}
	headers := parseSOVCSVLine(lines[0])
	rows := make([]map[string]any, 0, len(lines)-1)
	for _, line := range lines[1:] {
		values := parseSOVCSVLine(line)
		row := map[string]string{}
		for idx, key := range headers {
			if idx < len(values) {
				row[key] = values[idx]
			}
		}
		rows = append(rows, map[string]any{
			"week":            pickSOV(row, []string{"week", "Week", "date", "Date"}),
			"appName":         pickSOV(row, []string{"appName", "App Name"}),
			"appId":           pickSOV(row, []string{"adamId", "App ID", "appId"}),
			"countryOrRegion": pickSOV(row, []string{"countryOrRegion", "Country or Region", "country"}),
			"keyword":         pickSOV(row, []string{"searchTerm", "Search Term", "search_term"}),
			"popularity":      toSOVFloat(pickSOV(row, []string{"searchPopularity", "Search Popularity", "popularity"})),
			"impressionShare": toSOVFloat(pickSOV(row, []string{"impressionShare", "Impression Share"})),
			"rank":            toSOVFloat(pickSOV(row, []string{"rank", "Rank"})),
			"impressions":     toSOVInt(pickSOV(row, []string{"impressions", "Impressions"})),
			"taps":            toSOVInt(pickSOV(row, []string{"taps", "Taps"})),
			"installs":        toSOVInt(pickSOV(row, []string{"installs", "Installs", "tapThroughInstalls"})),
			"spend":           toSOVFloat(pickSOV(row, []string{"spend", "Spend"})),
		})
	}
	return rows
}

func buildSOVDecisionTable(rows []map[string]any) []map[string]any {
	results := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		keyword, _ := row["keyword"].(string)
		popularity, _ := row["popularity"].(float64)
		share, _ := row["impressionShare"].(float64)
		rank, _ := row["rank"].(float64)
		installs, _ := row["installs"].(int)

		bucket := "volume-limited"
		action := "expand-keyword-coverage"
		reason := "Low popularity and/or constrained volume"
		if popularity >= 60 && share < 0.2 {
			bucket = "auction-limited"
			action = "increase-bid-or-budget"
			reason = "High popularity with low share"
		} else if share >= 0.2 && installs == 0 {
			bucket = "conversion-limited"
			action = "improve-asa-cvr-surface"
			reason = "Receiving share but no installs"
		}

		results = append(results, map[string]any{
			"keyword":    keyword,
			"popularity": popularity,
			"share":      share,
			"rank":       rank,
			"bucket":     bucket,
			"action":     action,
			"reason":     reason,
		})
	}
	return results
}

func parseSOVCSVLine(line string) []string {
	out := []string{}
	current := strings.Builder{}
	inQuotes := false
	chars := []rune(line)
	for i := 0; i < len(chars); i++ {
		ch := chars[i]
		if ch == '"' {
			if inQuotes && i+1 < len(chars) && chars[i+1] == '"' {
				current.WriteRune('"')
				i++
				continue
			}
			inQuotes = !inQuotes
			continue
		}
		if ch == ',' && !inQuotes {
			out = append(out, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(ch)
	}
	out = append(out, current.String())
	return out
}

func pickSOV(row map[string]string, keys []string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(row[key]); value != "" {
			return value
		}
	}
	return ""
}

func toSOVFloat(raw string) float64 {
	cleaned := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(raw), "%", ""), ",", "")
	if cleaned == "" {
		return 0
	}
	value := 0.0
	_, _ = fmt.Sscanf(cleaned, "%f", &value)
	return value
}

func toSOVInt(raw string) int {
	return int(toSOVFloat(raw))
}
