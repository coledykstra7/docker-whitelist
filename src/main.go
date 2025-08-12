package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := setupRouter()
	r.Run(ServerPort)
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	registerRoutes(r)
	return r
}


