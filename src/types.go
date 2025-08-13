package main

import (
	"fmt"
	"path/filepath"
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
// Format: timestamp tag client-ip method status host url
// Example: "1712175100.000 WL 192.168.1.1 GET 200 example.com example.com:80"
type LogEntry struct {
	Timestamp   string // Field 0: Unix timestamp with milliseconds
	Tag         string // Field 1: WL/BL/RG (injected by mergeLogFiles)
	ClientIP    string // Field 2: Client IP address
	Method      string // Field 3: HTTP method (GET, POST, etc.)
	StatusCode  string // Field 4: HTTP status code (200, 404, etc.)
	Host        string // Field 5: Domain/hostname
	URL         string // Field 6: Full URL
}

// ParseLogEntry parses a space-separated log line into a LogEntry struct
func ParseLogEntry(line string) (*LogEntry, error) {
	fields := strings.Fields(line)
	if len(fields) < 7 {
		return nil, fmt.Errorf("insufficient fields: expected 7, got %d", len(fields))
	}
	
	return &LogEntry{
		Timestamp:  fields[0],
		Tag:        fields[1],
		ClientIP:   fields[2],
		Method:     fields[3],
		StatusCode: fields[4],
		Host:       fields[5],
		URL:        fields[6],
	}, nil
}

// File paths for configuration and logs
// These are variables so they can be overridden for testing
var (
	whitelistPath          = "/data/whitelist.txt"
	blacklistPath          = "/data/blacklist.txt"
	accessLogRegularPath   = "/data/access-regular.log"
	accessLogWhitelistPath = "/data/access-whitelist.log"
	accessLogBlacklistPath = "/data/access-blacklist.log"
	// Note: Access to individual log files should go through mergeLogFiles() function
)

// Helper function to set test data directory
func setTestDataDir(dataDir string) {
	whitelistPath = filepath.Join(dataDir, "whitelist.txt")
	blacklistPath = filepath.Join(dataDir, "blacklist.txt")
	accessLogRegularPath = filepath.Join(dataDir, "access-regular.log")
	accessLogWhitelistPath = filepath.Join(dataDir, "access-whitelist.log")
	accessLogBlacklistPath = filepath.Join(dataDir, "access-blacklist.log")
}

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
