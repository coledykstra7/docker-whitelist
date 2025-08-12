package main

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// registerRoutes sets up all HTTP routes
func registerRoutes(r *gin.Engine) {
	r.Static("/static", "html")
	r.POST("/save", handleSave)
	r.POST("/reload", handleReload)
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
	wl := readFile(whitelistPath)
	bl := readFile(blacklistPath)
	log := readFile(accessLogPath)
	summary := buildAccessLogSummaryFromLog(log, wl, bl)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, summary)
}

// handleLog provides live log tail
func handleLog(c *gin.Context) {
	log := readFile(accessLogPath)
	lines := strings.Split(log, "\n")
	if len(lines) > MaxLogLines {
		lines = lines[len(lines)-MaxLogLines:]
	}
	tail := strings.Join(lines, "\n")
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, tail)
}
