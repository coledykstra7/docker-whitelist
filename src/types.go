package main

import (
	"fmt"
	"strings"
)

// Row represents a domain entry with access statistics
type Row struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
	Status string `json:"status"`
	Url    string `json:"url"`
}

// LogEntry represents a parsed squid log line
// Format: timestamp tag method status_code host url
// Example: "1712175100.000 WL GET 200 example.com example.com:80"
type LogEntry struct {
	Timestamp   string // Field 0: Unix timestamp with milliseconds
	Tag         string // Field 1: WL/BL/RG (injected by mergeLogFiles)
	Method      string // Field 2: HTTP method (GET, POST, etc.)
	StatusCode  string // Field 3: HTTP status code (200, 404, etc.)
	Host        string // Field 4: Domain/hostname
	URL         string // Field 5: Full URL
}

// ParseLogEntry parses a space-separated log line into a LogEntry struct
func ParseLogEntry(line string) (*LogEntry, error) {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return nil, fmt.Errorf("insufficient fields: expected 6, got %d", len(fields))
	}
	
	return &LogEntry{
		Timestamp:  fields[0],
		Tag:        fields[1],
		Method:     fields[2],
		StatusCode: fields[3],
		Host:       fields[4],
		URL:        fields[5],
	}, nil
}

// File paths for configuration and logs
const (
	whitelistPath          = "/data/whitelist.txt"
	blacklistPath          = "/data/blacklist.txt"
	accessLogRegularPath   = "/data/access-regular.log"
	accessLogWhitelistPath = "/data/access-whitelist.log"
	accessLogBlacklistPath = "/data/access-blacklist.log"
	// Note: Access to individual log files should go through mergeLogFiles() function
)

// Application constants
const (
	MaxLogLines     = 50
	MaxFileSize     = 10 * 1024 * 1024 // 10MB
	FilePermissions = 0644
	ServerPort      = ":8080"
	SquidHost       = "squid-whitelist-proxy"
	SquidPort       = "3128"
	ConnectionTimeout = 400 // milliseconds
)
