package main

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	// Ensure required files exist on startup
	ensureRequiredFilesExist()
	
	r := setupRouter()
	r.Run(ServerPort)
}

// ensureRequiredFilesExist creates whitelist.txt and blacklist.txt if they don't exist
func ensureRequiredFilesExist() {
	// Ensure the data directory exists
	dataDir := filepath.Dir(whitelistPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic("Failed to create data directory: " + err.Error())
	}
	
	// Create whitelist.txt if it doesn't exist
	if _, err := os.Stat(whitelistPath); os.IsNotExist(err) {
		if err := writeFile(whitelistPath, ""); err != nil {
			panic("Failed to create whitelist.txt: " + err.Error())
		}
	}
	
	// Create blacklist.txt if it doesn't exist
	if _, err := os.Stat(blacklistPath); os.IsNotExist(err) {
		if err := writeFile(blacklistPath, ""); err != nil {
			panic("Failed to create blacklist.txt: " + err.Error())
		}
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	registerRoutes(r)
	return r
}


