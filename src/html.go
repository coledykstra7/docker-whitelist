package main

import (
	"fmt"
	"html"
	"strings"
)

// buildTableHTML builds HTML table for initial render
func buildTableHTML(rows []Row) string {
	var sb strings.Builder
	sb.WriteString("<table class=\"summary-table\"><tr><th>Domain</th><th>Count</th><th>Status</th></tr>")
	for _, r := range rows {
		cls := "unknown"
		if r.Status == "✅" {
			cls = "whitelist"
		} else if r.Status == "❌" {
			cls = "blacklist"
		}
		sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td><td class=\"status %s\">%s</td></tr>", 
			html.EscapeString(r.Domain), r.Count, cls, r.Status))
	}
	sb.WriteString("</table>")
	return sb.String()
}

// buildAccessLogSummaryFromLog builds summary from log for test compatibility
func buildAccessLogSummaryFromLog(log, whitelist, blacklist string) string {
	rows := computeSummaryRows(log, whitelist, blacklist)
	return buildTableHTML(rows)
}
