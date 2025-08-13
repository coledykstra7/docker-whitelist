package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestExtractDomain(t *testing.T) {
	cases := []struct{ in, want string }{
		{"http://example.com/path", "example.com"},
		{"https://example.com:443/", "example.com"},
		{"example.org", "example.org"},
		{"example.org:8080/foo", "example.org"},
	}
	for _, c := range cases {
		if got := extractDomain(c.in); got != c.want {
			t.Fatalf("extractDomain(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestBuildAccessLogSummaryFromLog(t *testing.T) {
	log := "" +
		"1712175100.000 WL 192.168.1.1 GET 200 example.com example.com:80\n" +
		"1712175101.000 BL 192.168.1.1 GET 200 blocked.com blocked.com:443\n" +
		"1712175102.000 WL 192.168.1.1 GET 200 example.com example.com:80\n"
	summary := buildAccessLogSummaryFromLog(log)
	if !containsAll(summary, []string{"example.com", "blocked.com", EmojiWhitelist, EmojiBlacklist}) {
		t.Fatalf("summary missing expected tokens:\n%s", summary)
	}
}

func TestIsDomainLikeRejectsNumeric(t *testing.T) {
	bad := []string{"1754845426.256", "123.456", "10.20.30.40"}
	for _, v := range bad {
		if isDomainLike(v) {
			t.Fatalf("expected isDomainLike(%q) == false", v)
		}
	}
	good := []string{"example.com", "sub.domain-1.org", "a.b"}
	for _, v := range good {
		if !isDomainLike(v) {
			t.Fatalf("expected isDomainLike(%q) == true", v)
		}
	}
}

func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

// setupTestRouter creates a test router with all routes
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerRoutes(router)
	return router
}

// setupTestFiles creates temporary test files and backs up original data
func setupTestFiles(t *testing.T) (cleanup func()) {
	// Create a temporary data directory for testing
	testDataDir := t.TempDir() + "/data"
	os.MkdirAll(testDataDir, 0755)
	
	// Store original paths
	origWhitelistPath := whitelistPath
	origBlacklistPath := blacklistPath
	origAccessLogRegularPath := accessLogRegularPath
	origAccessLogWhitelistPath := accessLogWhitelistPath
	origAccessLogBlacklistPath := accessLogBlacklistPath
	
	// Set test paths
	setTestDataDir(testDataDir)
	
	// Create backup directory
	backupDir := t.TempDir() + "/backup"
	os.MkdirAll(backupDir, 0755)
	
	// Backup original data files from the relative data directory if they exist
	originalDataDir := "../data"
	originalData := make(map[string]string)
	originalPaths := []string{
		filepath.Join(originalDataDir, "whitelist.txt"),
		filepath.Join(originalDataDir, "blacklist.txt"),
		filepath.Join(originalDataDir, "access-whitelist.log"),
		filepath.Join(originalDataDir, "access-blacklist.log"),
		filepath.Join(originalDataDir, "access-regular.log"),
	}
	
	for _, path := range originalPaths {
		if content := readFile(path); content != "" {
			originalData[path] = content
			// Write backup
			backupPath := filepath.Join(backupDir, filepath.Base(path))
			writeFile(backupPath, content)
		}
	}
	
	// Create test data in temp directory
	writeFile(whitelistPath, "example.com\nallowed.org #test note")
	writeFile(blacklistPath, "blocked.com\nbad.site #spam")
	writeFile(accessLogWhitelistPath, "1712175100.000 192.168.1.1 GET 200 example.com example.com:80")
	writeFile(accessLogBlacklistPath, "1712175101.000 192.168.1.1 GET 200 blocked.com blocked.com:443")
	writeFile(accessLogRegularPath, "1712175102.000 192.168.1.1 GET 200 unknown.com unknown.com:80")
	
	return func() {
		// Restore original paths
		whitelistPath = origWhitelistPath
		blacklistPath = origBlacklistPath
		accessLogRegularPath = origAccessLogRegularPath
		accessLogWhitelistPath = origAccessLogWhitelistPath
		accessLogBlacklistPath = origAccessLogBlacklistPath
		
		// Restore original data
		for path, content := range originalData {
			writeFile(path, content)
		}
		
		// Clean up backup directory (temp dirs clean themselves up)
		os.RemoveAll(backupDir)
	}
}

func TestGetHome(t *testing.T) {
	cleanup := setupTestFiles(t)
	defer cleanup()
	
	// Change to parent directory for template access
	originalWd, _ := os.Getwd()
	os.Chdir("..")
	defer os.Chdir(originalWd)
	
	// Verify test files were created at the paths the app expects
	t.Logf("Current working directory: %s", func() string { wd, _ := os.Getwd(); return wd }())
	t.Logf("Whitelist file exists: %v, content: %s", fileExists(whitelistPath), readFile(whitelistPath))
	t.Logf("Blacklist file exists: %v, content: %s", fileExists(blacklistPath), readFile(blacklistPath))
	
	router := setupTestRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		return
	}
	
	body := w.Body.String()
	t.Logf("Response body length: %d", len(body))
	
	if !strings.Contains(body, "Squid Proxy List Editor") {
		t.Error("Expected page title in response")
	}
	if !strings.Contains(body, "example.com") {
		t.Errorf("Expected whitelist content in response. Body contains: %s", body[:min(500, len(body))])
	}
	if !strings.Contains(body, "blocked.com") {
		t.Errorf("Expected blacklist content in response. Body contains: %s", body[:min(500, len(body))])
	}
}

func TestPostSave(t *testing.T) {
	cleanup := setupTestFiles(t)
	defer cleanup()
	
	router := setupTestRouter()
	
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("whitelist", "new-allowed.com\nexample.org")
	writer.WriteField("blacklist", "new-blocked.com\nbad-site.net")
	writer.Close()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/save", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303, got %d", w.Code)
	}
	
	// Verify files were updated
	wlContent := readFile(whitelistPath)
	blContent := readFile(blacklistPath)
	
	if !strings.Contains(wlContent, "new-allowed.com") {
		t.Error("Whitelist was not updated correctly")
	}
	if !strings.Contains(blContent, "new-blocked.com") {
		t.Error("Blacklist was not updated correctly")
	}
}

func TestPostReload(t *testing.T) {
	cleanup := setupTestFiles(t)
	defer cleanup()
	
	router := setupTestRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/reload", nil)
	router.ServeHTTP(w, req)
	
	// In test environment, docker commands will fail, so we expect 500
	// but the response should still be valid JSON with error info
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 (due to no Docker in test), got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Expected valid JSON response, got error: %v", err)
	}
	
	if response["status"] != "ERROR" {
		t.Error("Expected status field to be ERROR in test environment")
	}
	
	if response["error"] == nil {
		t.Error("Expected error field in response")
	}
}

func TestPostClearAllLogs(t *testing.T) {
	cleanup := setupTestFiles(t)
	defer cleanup()
	
	router := setupTestRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/clear-all-logs", nil)
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	if response["status"] != "all logs cleared" {
		t.Errorf("Expected 'all logs cleared', got %v", response["status"])
	}
	
	// Verify logs were cleared
	if readFile(accessLogWhitelistPath) != "" {
		t.Error("Whitelist log was not cleared")
	}
	if readFile(accessLogBlacklistPath) != "" {
		t.Error("Blacklist log was not cleared")
	}
	if readFile(accessLogRegularPath) != "" {
		t.Error("Regular log was not cleared")
	}
}

func TestPostMoveDomain(t *testing.T) {
	cleanup := setupTestFiles(t)
	defer cleanup()
	
	router := setupTestRouter()
	
	tests := []struct {
		name   string
		domain string
		target string
		note   string
		expect string
	}{
		{"Move to whitelist", "newdomain.com", "whitelist", "test note", "whitelist"},
		{"Move to blacklist", "baddomain.com", "blacklist", "spam", "blacklist"},
		{"Move to unknown", "example.com", "unknown", "", "unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			writer.WriteField("domain", tt.domain)
			writer.WriteField("target", tt.target)
			writer.WriteField("note", tt.note)
			writer.Close()
			
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/move-domain", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
			
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			
			if response["status"] != "success" {
				t.Errorf("Expected success, got %v", response["status"])
			}
			
			// Verify domain was moved correctly
			wlContent := readFile(whitelistPath)
			blContent := readFile(blacklistPath)
			
			switch tt.target {
			case "whitelist":
				expectedEntry := tt.domain
				if tt.note != "" {
					expectedEntry = tt.domain + " #" + tt.note
				}
				if !strings.Contains(wlContent, expectedEntry) {
					t.Errorf("Domain not found in whitelist: %s", expectedEntry)
				}
			case "blacklist":
				expectedEntry := tt.domain
				if tt.note != "" {
					expectedEntry = tt.domain + " #" + tt.note
				}
				if !strings.Contains(blContent, expectedEntry) {
					t.Errorf("Domain not found in blacklist: %s", expectedEntry)
				}
			case "unknown":
				if strings.Contains(wlContent, tt.domain) || strings.Contains(blContent, tt.domain) {
					t.Error("Domain should have been removed from both lists")
				}
			}
		})
	}
}

func TestGetSummary(t *testing.T) {
	cleanup := setupTestFiles(t)
	defer cleanup()
	
	router := setupTestRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/summary", nil)
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "<table") {
		t.Error("Expected HTML table in response")
	}
}

func TestGetSummaryData(t *testing.T) {
	cleanup := setupTestFiles(t)
	defer cleanup()
	
	router := setupTestRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/summary-data", nil)
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	if response["rows"] == nil {
		t.Error("Expected rows field in response")
	}
}

func TestGetLog(t *testing.T) {
	cleanup := setupTestFiles(t)
	defer cleanup()
	
	router := setupTestRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/log", nil)
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "WL") && !strings.Contains(body, "BL") && !strings.Contains(body, "RG") {
		t.Error("Expected log entries with tags in response")
	}
}

func TestGetLists(t *testing.T) {
	cleanup := setupTestFiles(t)
	defer cleanup()
	
	router := setupTestRouter()
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lists", nil)
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	if response["whitelist"] == nil || response["blacklist"] == nil {
		t.Error("Expected whitelist and blacklist fields in response")
	}
	
	wl := response["whitelist"].(string)
	bl := response["blacklist"].(string)
	
	if !strings.Contains(wl, "example.com") {
		t.Error("Expected example.com in whitelist")
	}
	if !strings.Contains(bl, "blocked.com") {
		t.Error("Expected blocked.com in blacklist")
	}
}

func TestMoveDomainEdgeCases(t *testing.T) {
	cleanup := setupTestFiles(t)
	defer cleanup()
	
	router := setupTestRouter()
	
	tests := []struct {
		name           string
		domain         string
		target         string
		expectedStatus int
	}{
		{"Empty domain", "", "whitelist", http.StatusBadRequest},
		{"Invalid target", "test.com", "invalid", http.StatusBadRequest},
		{"Valid move", "valid.com", "whitelist", http.StatusOK},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			writer.WriteField("domain", tt.domain)
			writer.WriteField("target", tt.target)
			writer.WriteField("note", "")
			writer.Close()
			
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/move-domain", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			router.ServeHTTP(w, req)
			
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
