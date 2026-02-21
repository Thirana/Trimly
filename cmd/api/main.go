package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/thirana/url-shortener/internal/httpapi"
)

func main() {
	// Production default: release mode reduces debug noise and overhead.
	// You can also set: export GIN_MODE=release
	gin.SetMode(gin.ReleaseMode)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := httpapi.NewRouter()

	log.Printf("starting api on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
