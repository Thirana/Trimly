package httpapi

import (
	"github.com/gin-gonic/gin"
	"github.com/thirana/url-shortener/internal/shortener"
	"github.com/thirana/url-shortener/internal/store"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", Health)

	// Dependencies (manual DI)
	mem := store.NewMemoryStore()
	svc := shortener.NewService(mem)
	links := NewLinksHandler(svc)

	v1 := r.Group("/v1")
	{
		v1.POST("/links", links.Create)
	}

	// Redirect endpoint (public)
	// Bind URI uses uri tags; Gin shows this pattern in its docs. ([gin Bind URI](https://gin-gonic.com/en/docs/examples/bind-uri/?utm_source=chatgpt.com))
	r.GET("/:code", links.Redirect)

	return r
}
