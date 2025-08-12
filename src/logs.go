package main

import (
	"sort"
	"strconv"
	"strings"
)

// computeSummaryRows builds summary rows from log/whitelist/blacklist
func computeSummaryRows(logText, whitelist, blacklist string) []Row {
	wset := make(map[string]struct{})
	for _, line := range strings.Split(whitelist, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		wset[line] = struct{}{}
	}
	
	bset := make(map[string]struct{})
	for _, line := range strings.Split(blacklist, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		bset[line] = struct{}{}
	}
	
	// Status by log tag: WL=✅, BL=❌, RG=❓
	type stat struct {
		count  int
		status string
	}
	counts := make(map[string]stat)
	latestUrl := make(map[string]string)
	
	for _, line := range strings.Split(logText, "\n") {
		entry, err := ParseLogEntry(line)
		if err != nil {
			continue // Skip malformed lines
		}
		
		host := entry.Host
		if host == "" {
			continue
		}
		
		var status string
		switch entry.Tag {
		case "WL":
			status = "✅"
		case "BL":
			status = "❌"
		default:
			status = "❓"
		}
		
		val := counts[host]
		// If status is more severe (❌ > ✅ > ❓), keep highest
		if val.status == "❌" || (val.status == "✅" && status == "❓") {
			counts[host] = stat{val.count + 1, val.status}
		} else {
			counts[host] = stat{val.count + 1, status}
		}
		latestUrl[host] = entry.URL
	}
	
	type kv struct {
		k string
		v stat
	}
	arr := make([]kv, 0, len(counts))
	for k, v := range counts {
		arr = append(arr, kv{k, v})
	}
	
	// Sort by domain parts in reverse order, ignoring TLD
	sort.Slice(arr, func(i, j int) bool {
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
		aSub, aTld := split(arr[i].k)
		bSub, bTld := split(arr[j].k)
		for x := 0; x < len(aSub) && x < len(bSub); x++ {
			if aSub[x] != bSub[x] {
				return aSub[x] < bSub[x]
			}
		}
		if len(aSub) != len(bSub) {
			return len(aSub) < len(bSub)
		}
		return aTld < bTld
	})
	
	rows := make([]Row, 0, len(arr))
	for _, kv := range arr {
		rows = append(rows, Row{Domain: kv.k, Count: kv.v.count, Status: kv.v.status, Url: latestUrl[kv.k]})
	}
	return rows
}

// filterLogAfterSetpoint keeps only lines whose first field (float or int) is > setpoint
func filterLogAfterSetpoint(logText string, setpoint float64) string {
	if setpoint <= 0 {
		return logText
	}
	var out []string
	for _, line := range strings.Split(logText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		ts, err := strconv.ParseFloat(f[0], 64)
		if err != nil {
			continue
		}
		if ts > setpoint {
			out = append(out, line)
		}
	}
	result := strings.Join(out, "\n")
	if len(out) > 0 {
		result += "\n"
	}
	return result
}

// mergeLogFiles reads the three categorized logs, parses first field as timestamp,
// merges and returns them in chronological order as a single string (ending with \n if any records).
func mergeLogFiles() string {
	type rec struct {
		ts   float64
		line string
	}
	var recs []rec
	type fileTag struct{ path, tag string }
	files := []fileTag{
		{accessLogWhitelistPath, "WL"}, 
		{accessLogBlacklistPath, "BL"}, 
		{accessLogRegularPath, "RG"},
	}
	
	for _, ft := range files {
		content := readFile(ft.path)
		if content == "" {
			continue
		}
		for _, line := range strings.Split(strings.TrimRight(content, "\n"), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			f := strings.Fields(line)
			if len(f) == 0 {
				continue
			}
			// Skip lines where the status code is "0" 
			// Original squid format: timestamp client-ip method status host url
			// Status code 0 indicates connection errors/failures, not actual requests
			if len(f) >= 4 && f[3] == "0" {
				continue
			}
			ts, err := strconv.ParseFloat(f[0], 64)
			if err != nil {
				recs = append(recs, rec{ts: 0, line: line})
				continue
			}
			// Inject tag as second field (after timestamp) so timestamp stays first for setpoint filtering
			// Preserve original spacing after first token
			rest := strings.TrimPrefix(line, f[0])
			taggedLine := f[0] + " " + ft.tag + rest
			recs = append(recs, rec{ts: ts, line: taggedLine})
		}
	}
	
	sort.Slice(recs, func(i, j int) bool { return recs[i].ts < recs[j].ts })
	var b strings.Builder
	for _, r := range recs {
		b.WriteString(r.line)
		b.WriteByte('\n')
	}
	return b.String()
}
