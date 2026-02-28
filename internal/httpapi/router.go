package httpapi

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(links *LinksHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", Health)

	v1 := r.Group("/v1")
	{
		v1.POST("/links", links.Create)
	}

	// Redirect endpoint (public)
	// Bind URI uses uri tags; Gin shows this pattern in its docs. ([gin Bind URI](https://gin-gonic.com/en/docs/examples/bind-uri/?utm_source=chatgpt.com))
	r.GET("/:code", links.Redirect)

	return r
}
