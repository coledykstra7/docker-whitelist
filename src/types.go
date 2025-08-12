package main

// Row represents a domain entry with access statistics
type Row struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
	Status string `json:"status"`
	Url    string `json:"url"`
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
