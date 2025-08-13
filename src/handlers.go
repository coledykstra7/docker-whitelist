package main

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// Global setpoint for filtering logs (Unix timestamp)
var logSetpoint float64 = 0

// registerRoutes sets up all HTTP routes
func registerRoutes(r *gin.Engine) {
	r.Static("/static", "html")
	r.POST("/save", handleSave)
	r.POST("/reload", handleReload)
	r.POST("/setpoint", handleSetpoint)
	r.POST("/clear", handleClear)
	r.GET("/", handleHome)
	r.GET("/summary", handleSummary)
	r.GET("/log", handleLog)
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

// handleSetpoint sets a setpoint to filter logs after current time
func handleSetpoint(c *gin.Context) {
	// Set setpoint to current Unix timestamp with milliseconds
	logSetpoint = float64(time.Now().UnixMilli()) / 1000.0
	
	c.JSON(http.StatusOK, gin.H{
		"status": "setpoint_set", 
		"setpoint": strconv.FormatFloat(logSetpoint, 'f', 3, 64),
	})
}

// handleClear clears the setpoint filter
func handleClear(c *gin.Context) {
	logSetpoint = 0
	
	c.JSON(http.StatusOK, gin.H{
		"status": "cleared", 
		"setpoint": "none",
	})
}

// handleReset clears the setpoint to show all logs
func handleReset(c *gin.Context) {
	logSetpoint = 0
	
	c.JSON(http.StatusOK, gin.H{
		"status": "reset", 
		"setpoint": "none",
	})
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
	// Apply setpoint filtering if set
	if logSetpoint > 0 {
		log = filterLogAfterSetpoint(log, logSetpoint)
	}
	summary := buildAccessLogSummaryFromLog(log)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, summary)
}

// handleLog provides live log tail
func handleLog(c *gin.Context) {
	log := mergeLogFiles()
	// Apply setpoint filtering if set
	if logSetpoint > 0 {
		log = filterLogAfterSetpoint(log, logSetpoint)
	}
	lines := strings.Split(log, "\n")
	if len(lines) > MaxLogLines {
		lines = lines[len(lines)-MaxLogLines:]
	}
	tail := strings.Join(lines, "\n")
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, tail)
}
