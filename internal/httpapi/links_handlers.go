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

	link, err := h.svc.Create(c.Request.Context(), req.LongURL)
	if err != nil {
		if err == shortener.ErrInvalidURL {
			jsonError(c, http.StatusBadRequest, "invalid_url", "URL must be a valid http/https URL")
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
		Code:     link.Code,
		ShortURL: baseURL + "/" + link.Code,
		LongURL:  link.LongURL,
	}
	c.JSON(http.StatusCreated, resp)
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
		if err == shortener.ErrNotFound {
			jsonError(c, http.StatusNotFound, "not_found", "code not found")
			return
		}
		jsonError(c, http.StatusInternalServerError, "internal", "something went wrong")
		return
	}

	// For now choose 302 Temporary Redirect.
	// net/http documents standard redirect codes; we can choose 301/308 later for “permanent”. ([pkg.go.dev/net/http](https://pkg.go.dev/net/http?utm_source=chatgpt.com))
	c.Redirect(http.StatusFound, link.LongURL)
}
