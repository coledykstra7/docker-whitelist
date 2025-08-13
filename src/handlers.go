package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// noCacheMiddleware adds strict no-cache headers to all responses
func noCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// registerRoutes sets up all HTTP routes
func registerRoutes(r *gin.Engine) {
	// Apply no-cache middleware to all routes
	r.Use(noCacheMiddleware())
	
	r.Static("/static", "html")
	r.POST("/save", handleSave)
	r.POST("/reload", handleReload)
	r.POST("/clear-all-logs", handleClearAllLogs)
	r.POST("/move-domain", handleMoveDomain)
	r.GET("/", handleHome)
	r.GET("/summary", handleSummary)
	r.GET("/summary-data", handleSummaryData)
	r.GET("/log", handleLog)
	r.GET("/lists", handleLists)
}

// handleSave processes whitelist/blacklist updates
func handleSave(c *gin.Context) {
	wl := c.PostForm("whitelist")
	bl := c.PostForm("blacklist")
	err1 := writeFile(whitelistPath, wl)
	err2 := writeFile(blacklistPath, bl)
	if err1 != nil || err2 != nil {
		c.String(http.StatusInternalServerError, "save error: %v %v", err1, err2)
		return
	}
	// Optionally reload squid after save
	_ = reloadSquid()
	c.Redirect(http.StatusSeeOther, "/")
}

// handleReload triggers squid reconfiguration
func handleReload(c *gin.Context) {
	if err := reloadSquid(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERROR", "error": err.Error()})
		return
	}
	// brief delay to allow reconfigure
	time.Sleep(300 * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{"status": squidStatus()})
}

// handleClearAllLogs clears all access logs (whitelist, blacklist, and regular)
func handleClearAllLogs(c *gin.Context) {
	err1 := writeFile(accessLogWhitelistPath, "")
	err2 := writeFile(accessLogBlacklistPath, "")
	err3 := writeFile(accessLogRegularPath, "")
	
	if err1 != nil || err2 != nil || err3 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("errors: %v %v %v", err1, err2, err3)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "all logs cleared"})
}

// handleHome serves the main page with whitelist/blacklist editor
func handleHome(c *gin.Context) {
	wl := readFile(whitelistPath)
	bl := readFile(blacklistPath)
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFiles("html/template.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "template error: %v", err)
		return
	}
	data := struct {
		Whitelist string
		Blacklist string
	}{
		Whitelist: wl,
		Blacklist: bl,
	}
	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, "template exec error: %v", err)
	}
}

// handleSummary provides live summary box content
func handleSummary(c *gin.Context) {
	log := mergeLogFiles()
	summary := buildAccessLogSummaryFromLog(log)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, summary)
}

// handleSummaryData provides summary data as JSON for filtering
func handleSummaryData(c *gin.Context) {
	log := mergeLogFiles()
	rows := computeSummaryRows(log)
	c.JSON(http.StatusOK, gin.H{"rows": rows})
}

// handleLog provides live log tail
func handleLog(c *gin.Context) {
	log := mergeLogFiles()
	lines := strings.Split(log, "\n")
	if len(lines) > MaxLogLines {
		lines = lines[len(lines)-MaxLogLines:]
	}
	tail := strings.Join(lines, "\n")
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, tail)
}

// handleMoveDomain moves a domain between whitelist, blacklist, or unknown status
func handleMoveDomain(c *gin.Context) {
	domain := strings.TrimSpace(c.PostForm("domain"))
	target := strings.TrimSpace(c.PostForm("target"))
	note := strings.TrimSpace(c.PostForm("note"))
	
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "domain is required"})
		return
	}
	
	if target != "whitelist" && target != "blacklist" && target != "unknown" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "target must be whitelist, blacklist, or unknown"})
		return
	}
	
	// Read current whitelist and blacklist
	whitelistContent := readFile(whitelistPath)
	blacklistContent := readFile(blacklistPath)
	
	// Parse into slices
	whitelistDomains := parseDomainList(whitelistContent)
	blacklistDomains := parseDomainList(blacklistContent)
	
	// Remove domain from both lists first (strip any existing notes when removing)
	whitelistDomains = removeDomainFromList(whitelistDomains, domain)
	blacklistDomains = removeDomainFromList(blacklistDomains, domain)
	
	// Add to target list if not unknown
	switch target {
	case "whitelist":
		entry := domain
		if note != "" {
			entry = fmt.Sprintf("%s #%s", domain, note)
		}
		whitelistDomains = append(whitelistDomains, entry)
	case "blacklist":
		entry := domain
		if note != "" {
			entry = fmt.Sprintf("%s #%s", domain, note)
		}
		blacklistDomains = append(blacklistDomains, entry)
	// "unknown" means just remove from both lists (already done above)
	}
	
	// Write back to files
	newWhitelistContent := strings.Join(whitelistDomains, "\n")
	newBlacklistContent := strings.Join(blacklistDomains, "\n")
	
	err1 := writeFile(whitelistPath, newWhitelistContent)
	err2 := writeFile(blacklistPath, newBlacklistContent)
	
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("write error: %v %v", err1, err2)})
		return
	}
	
	// Reload squid configuration
	_ = reloadSquid()
	
	c.JSON(http.StatusOK, gin.H{"status": "success", "domain": domain, "target": target})
}

// parseDomainList parses a domain list content into a slice of domains
func parseDomainList(content string) []string {
	var domains []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			domains = append(domains, line)
		}
	}
	return domains
}

// removeDomainFromList removes a domain from a slice of domains, handling entries with notes
func removeDomainFromList(domains []string, target string) []string {
	var result []string
	for _, entry := range domains {
		// Extract just the domain part (before any # comment)
		domainPart := strings.TrimSpace(strings.Split(entry, "#")[0])
		if domainPart != target {
			result = append(result, entry)
		}
	}
	return result
}

// handleLists returns the current whitelist and blacklist content as JSON
func handleLists(c *gin.Context) {
	wl := readFile(whitelistPath)
	bl := readFile(blacklistPath)
	c.JSON(http.StatusOK, gin.H{
		"whitelist": wl,
		"blacklist": bl,
	})
}
