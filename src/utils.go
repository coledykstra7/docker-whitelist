package main

import (
	"fmt"
	"sort"
	"strings"
)

// isDomainLike does a lightweight check for domain style tokens (example.com or example.com:443)
func isDomainLike(s string) bool {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return true
	}
	// Trim trailing punctuation
	s = strings.Trim(s, ",.;")
	// Remove port
	if i := strings.LastIndex(s, ":"); i > 0 && i < len(s)-1 {
		s = s[:i]
	}
	if strings.Count(s, ".") == 0 {
		return false
	}
	// Reject pure numeric dotted tokens (e.g., timestamps like 1754845426.256 or IP-like 1.2.3.4 unless you want IPs counted)
	pureNumeric := true
	hasAlpha := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			hasAlpha = true
			pureNumeric = false
		case r >= 'A' && r <= 'Z':
			hasAlpha = true
			pureNumeric = false
		case r >= '0' && r <= '9':
			// still possibly numeric
		case r == '.' || r == '-':
			// allowed separators
		}
	}
	if pureNumeric {
		return false
	}
	if !hasAlpha {
		return false
	}
	// Basic label validation
	labels := strings.Split(s, ".")
	for _, lab := range labels {
		if lab == "" {
			return false
		}
		if len(lab) > 63 {
			return false
		}
	}
	return true
}

// extractDomain extracts domain from URL for test compatibility
func extractDomain(url string) string {
	url = strings.TrimSpace(url)
	if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	}
	if i := strings.Index(url, "/"); i >= 0 {
		url = url[:i]
	}
	if i := strings.Index(url, ":"); i >= 0 {
		url = url[:i]
	}
	return url
}

// DomainEntry represents a domain list entry with domain and optional note
type DomainEntry struct {
	Domain string
	Note   string
	Full   string // The full line including note
}

// parseDomainEntry parses a domain list line into domain and note parts
func parseDomainEntry(line string) DomainEntry {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return DomainEntry{}
	}
	
	parts := strings.SplitN(line, "#", 2)
	domain := strings.TrimSpace(parts[0])
	note := ""
	if len(parts) > 1 {
		// Handle both formats: "domain#note" and "domain # note" and "domain  # note"
		note = strings.TrimSpace(parts[1])
	}
	
	return DomainEntry{
		Domain: domain,
		Note:   note,
		Full:   line,
	}
}

// sortDomainsByParts sorts domains by reverse domain parts, ignoring TLD
func sortDomainsByParts(a, b string) bool {
	split := func(domain string) ([]string, string) {
		parts := strings.Split(domain, ".")
		if len(parts) < 2 {
			return parts, ""
		}
		tld := parts[len(parts)-1]
		sub := parts[:len(parts)-1]
		for l, r := 0, len(sub)-1; l < r; l, r = l+1, r-1 {
			sub[l], sub[r] = sub[r], sub[l]
		}
		return sub, tld
	}
	
	aSub, aTld := split(a)
	bSub, bTld := split(b)
	for x := 0; x < len(aSub) && x < len(bSub); x++ {
		if aSub[x] != bSub[x] {
			return aSub[x] < bSub[x]
		}
	}
	if len(aSub) != len(bSub) {
		return len(aSub) < len(bSub)
	}
	return aTld < bTld
}

// sortDomainEntries sorts domain entries by note first, then by domain parts
func sortDomainEntries(entries []DomainEntry) {
	sort.Slice(entries, func(i, j int) bool {
		// First sort by note (alphabetically)
		if entries[i].Note != entries[j].Note {
			return entries[i].Note < entries[j].Note
		}
		// Then sort by domain parts in reverse order
		return sortDomainsByParts(entries[i].Domain, entries[j].Domain)
	})
}

// sortAndJoinDomainList takes a slice of domain strings, sorts them by note then domain parts, and joins them
func sortAndJoinDomainList(domains []string) string {
	if len(domains) == 0 {
		return ""
	}
	
	// Parse entries
	entries := make([]DomainEntry, 0, len(domains))
	maxDomainLen := 0
	for _, domain := range domains {
		if entry := parseDomainEntry(domain); entry.Domain != "" {
			entries = append(entries, entry)
			if len(entry.Domain) > maxDomainLen {
				maxDomainLen = len(entry.Domain)
			}
		}
	}
	
	// Sort entries
	sortDomainEntries(entries)
	
	// Convert back to strings with column alignment
	result := make([]string, len(entries))
	for i, entry := range entries {
		if entry.Note != "" {
			// Pad domain to align notes in a column
			padding := maxDomainLen - len(entry.Domain) + 2 // 2 extra spaces
			result[i] = entry.Domain + strings.Repeat(" ", padding) + "# " + entry.Note
		} else {
			result[i] = entry.Domain
		}
	}
	
	return strings.Join(result, "\n")
}

// writeDomainList handles all whitelist/blacklist file writes with consistent sorting
func writeDomainList(listType string, domains []string) error {
	var filePath string
	switch listType {
	case "whitelist":
		filePath = whitelistPath
	case "blacklist":
		filePath = blacklistPath
	default:
		return fmt.Errorf("invalid list type: %s", listType)
	}
	
	sortedContent := sortAndJoinDomainList(domains)
	err := writeFile(filePath, sortedContent)
	if err == nil {
		_ = reloadSquid() // Always reload after successful write
	}
	return err
}
