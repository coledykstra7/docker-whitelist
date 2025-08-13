package main

import (
	"fmt"
	"html"
	"strings"
)

// buildTableHTML builds HTML table for initial render
func buildTableHTML(rows []Row) string {
	var sb strings.Builder
	sb.WriteString("<table class=\"summary-table\"><tr><th>Domain</th><th>Count</th><th>Status</th><th>Actions</th></tr>")
	for _, r := range rows {
		cls := "unknown"
		if r.Status == EmojiWhitelist {
			cls = "whitelist"
		} else if r.Status == EmojiBlacklist {
			cls = "blacklist"
		}
		
		// Generate action buttons based on current status
		var actions string
		domain := html.EscapeString(r.Domain)
		if r.Status == EmojiWhitelist {
			// Whitelisted: can move to blacklist or remove (make unknown)
			actions = fmt.Sprintf(`<button onclick="moveDomain('%s', 'blacklist')" class="action-btn bl">→ BL</button> <button onclick="moveDomain('%s', 'unknown')" class="action-btn unknown">→ ?</button>`, domain, domain)
		} else if r.Status == EmojiBlacklist {
			// Blacklisted: can move to whitelist or remove (make unknown)  
			actions = fmt.Sprintf(`<button onclick="moveDomain('%s', 'whitelist')" class="action-btn wl">→ WL</button> <button onclick="moveDomain('%s', 'unknown')" class="action-btn unknown">→ ?</button>`, domain, domain)
		} else {
			// Unknown: can move to whitelist or blacklist
			actions = fmt.Sprintf(`<button onclick="moveDomain('%s', 'whitelist')" class="action-btn wl">→ WL</button> <button onclick="moveDomain('%s', 'blacklist')" class="action-btn bl">→ BL</button>`, domain, domain)
		}
		
		sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td><td class=\"status %s\">%s</td><td>%s</td></tr>", 
			domain, r.Count, cls, r.Status, actions))
	}
	sb.WriteString("</table>")
	return sb.String()
}

// buildAccessLogSummaryFromLog builds summary from log for test compatibility
func buildAccessLogSummaryFromLog(log string) string {
	rows := computeSummaryRows(log)
	return buildTableHTML(rows)
}
