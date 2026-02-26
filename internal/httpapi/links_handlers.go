package httpapi

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/thirana/url-shortener/internal/shortener"
)

type LinksHandler struct {
	svc *shortener.Service
}

func NewLinksHandler(svc *shortener.Service) *LinksHandler {
	return &LinksHandler{svc: svc}
}

func (h *LinksHandler) Create(c *gin.Context) {
	var req CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		jsonError(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	link, created, err := h.svc.Create(c.Request.Context(), req.LongURL, req.ExpiresAt)
	if err != nil {
		if mapCreateDomainError(c, err) {
			return
		}
		jsonError(c, http.StatusInternalServerError, "internal", "something went wrong")
		return
	}

	baseURL := os.Getenv("BASE_URL") // e.g. https://sho.rt (later)
	if baseURL == "" {
		// Local default: you can set BASE_URL in prod
		baseURL = "http://localhost:8080"
	}

	resp := CreateLinkResponse{
		Code:      link.Code,
		ShortURL:  baseURL + "/" + link.Code,
		LongURL:   link.LongURL,
		ExpiresAt: link.ExpiresAt,
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, resp)
}

func (h *LinksHandler) Redirect(c *gin.Context) {
	// Bind URI param cleanly (Gin example uses ShouldBindUri)
	var uri RedirectURI
	// We’ll mount this route as "/:code" so uri tag should match.
	if err := c.ShouldBindUri(&uri); err != nil {
		jsonError(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	link, err := h.svc.Resolve(c.Request.Context(), uri.Code)
	if err != nil {
		if mapResolveDomainError(c, err) {
			return
		}
		jsonError(c, http.StatusInternalServerError, "internal", "something went wrong")
		return
	}

	// For now choose 302 Temporary Redirect.
	// net/http documents standard redirect codes; we can choose 301/308 later for “permanent”. ([pkg.go.dev/net/http](https://pkg.go.dev/net/http?utm_source=chatgpt.com))
	c.Redirect(http.StatusFound, link.LongURL)
}
