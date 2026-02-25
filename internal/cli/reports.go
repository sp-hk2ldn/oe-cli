package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"oe-cli/internal/appleads"
)

func RunReports(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("reports", jsonOut, err)
		return
	}

	action := actionFromArgs(args, "list")
	switch action {
	case "list":
		runReportsList(ctx, client, args, jsonOut)
	case "get", "show":
		runReportsGet(ctx, client, args, jsonOut)
	case "download":
		runReportsDownload(ctx, client, args, jsonOut)
	default:
		respondCommandError("reports", jsonOut, fmt.Errorf("Unsupported reports action: %s. Use: list|get|download", action))
	}
}

func runReportsList(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	reports, err := client.FetchCustomReports(ctx)
	if err != nil {
		respondCommandError("reports", jsonOut, err)
		return
	}

	stateFilters := parseStringSet(splitCSVValues(valuesForFlag(args, "--state")), true)
	nameContains := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--nameContains")))
	limit := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--limit")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &limit)
		if limit < 0 {
			limit = 0
		}
	}

	filtered := make([]appleads.CustomReport, 0, len(reports))
	for _, report := range reports {
		if len(stateFilters) > 0 {
			if _, ok := stateFilters[strings.ToUpper(strings.TrimSpace(report.State))]; !ok {
				continue
			}
		}
		if nameContains != "" && !strings.Contains(strings.ToLower(report.Name), nameContains) {
			continue
		}
		filtered = append(filtered, report)
	}
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID > filtered[j].ID })

	if jsonOut {
		printJSON(filtered)
		return
	}
	fmt.Printf("reportCount=%d\n", len(filtered))
	for _, report := range filtered {
		download := "-"
		if report.DownloadURI != nil {
			download = safeDisplayURL(*report.DownloadURI)
		}
		fmt.Printf("%d\t%s\t%s\t%s\t%s\n", report.ID, report.State, report.Granularity, report.Name, download)
	}
}

func runReportsGet(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	reportID, err := requiredInt64Flag(args, "--reportId")
	if err != nil {
		respondCommandError("reports", jsonOut, err)
		return
	}
	report, err := client.FetchImpressionShareReport(ctx, reportID)
	if err != nil {
		respondCommandError("reports", jsonOut, err)
		return
	}

	if jsonOut {
		printJSON(report)
		return
	}
	fmt.Printf("reportId=%d\n", report.ID)
	fmt.Printf("name=%s\n", report.Name)
	fmt.Printf("state=%s\n", report.State)
	fmt.Printf("granularity=%s\n", report.Granularity)
	if report.StartTime != nil {
		fmt.Printf("startTime=%s\n", *report.StartTime)
	}
	if report.EndTime != nil {
		fmt.Printf("endTime=%s\n", *report.EndTime)
	}
	if report.DateRange != nil {
		fmt.Printf("dateRange=%s\n", *report.DateRange)
	}
	if report.DownloadURI != nil {
		fmt.Printf("downloadUri=%s\n", safeDisplayURL(*report.DownloadURI))
	}
}

func runReportsDownload(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	reportID, err := requiredInt64Flag(args, "--reportId")
	if err != nil {
		respondCommandError("reports", jsonOut, err)
		return
	}
	report, err := client.FetchImpressionShareReport(ctx, reportID)
	if err != nil {
		respondCommandError("reports", jsonOut, err)
		return
	}
	if report.DownloadURI == nil || strings.TrimSpace(*report.DownloadURI) == "" {
		respondCommandError("reports", jsonOut, fmt.Errorf("Report %d does not have a downloadUri yet. Current state=%s", reportID, report.State))
		return
	}
	data, err := client.DownloadCustomReport(ctx, *report.DownloadURI)
	if err != nil {
		respondCommandError("reports", jsonOut, err)
		return
	}

	outPath := strings.TrimSpace(valueForFlag(args, "--out"))
	if outPath == "" {
		outPath = filepath.Join("reports", "custom", fmt.Sprintf("%d.csv", reportID))
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o700); err != nil {
		respondCommandError("reports", jsonOut, err)
		return
	}
	if err := os.WriteFile(outPath, data, 0o600); err != nil {
		respondCommandError("reports", jsonOut, err)
		return
	}

	if jsonOut {
		printJSON(map[string]any{"ok": true, "reportId": reportID, "state": report.State, "bytes": len(data), "out": outPath})
		return
	}
	fmt.Printf("ok reportId=%d state=%s bytes=%d out=%s\n", reportID, report.State, len(data), outPath)
}

func requiredInt64Flag(args []string, flag string) (int64, error) {
	raw := strings.TrimSpace(valueForFlag(args, flag))
	if raw == "" {
		return 0, fmt.Errorf("Missing required %s <id>", flag)
	}
	var id int64
	if _, err := fmt.Sscanf(raw, "%d", &id); err != nil || id <= 0 {
		return 0, fmt.Errorf("Invalid %s %q", flag, raw)
	}
	return id, nil
}
