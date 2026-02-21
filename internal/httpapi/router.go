package httpapi

import "github.com/gin-gonic/gin"

func NewRouter() *gin.Engine {
	// gin.New() creates a router without default middleware.
	// We attach only what we need.
	r := gin.New()

	// Logger middleware logs requests.
	// Recovery middleware prevents panics from crashing the process.
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Routes
	r.GET("/health", Health)

	return r
}