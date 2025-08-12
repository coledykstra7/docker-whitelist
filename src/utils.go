package main

import (
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
