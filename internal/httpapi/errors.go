package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thirana/url-shortener/internal/shortener"
)

func jsonError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

func mapCreateDomainError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, shortener.ErrInvalidURL):
		jsonError(c, http.StatusBadRequest, "invalid_url", "URL must be a valid http/https URL")
	case errors.Is(err, shortener.ErrExpired):
		jsonError(c, http.StatusBadRequest, "invalid_expiry", "expires_at must be in the future")
	case errors.Is(err, shortener.ErrCollision):
		jsonError(c, http.StatusConflict, "conflict", "could not allocate a unique short code")
	default:
		return false
	}
	return true
}

func mapResolveDomainError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, shortener.ErrNotFound), errors.Is(err, shortener.ErrExpired):
		// Expired links intentionally look like not found.
		jsonError(c, http.StatusNotFound, "not_found", "code not found")
	default:
		return false
	}
	return true
}
