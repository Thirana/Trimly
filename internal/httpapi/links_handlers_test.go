package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thirana/url-shortener/internal/shortener"
	"github.com/thirana/url-shortener/internal/store"
)

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type failingStore struct {
	saveErr      error
	getErr       error
	findByIntent error
}

func (s *failingStore) Save(_ context.Context, _ store.Link) error {
	return s.saveErr
}

func (s *failingStore) Get(_ context.Context, _ string) (store.Link, bool, error) {
	return store.Link{}, false, s.getErr
}

func (s *failingStore) FindByIntent(_ context.Context, _ string, _ *time.Time) (store.Link, bool, error) {
	return store.Link{}, false, s.findByIntent
}

func testRouterWithService(svc *shortener.Service) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	links := NewLinksHandler(svc)
	v1 := r.Group("/v1")
	v1.POST("/links", links.Create)
	r.GET("/:code", links.Redirect)

	return r
}

func TestCreateLink_IdempotentDuplicateReturns200(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	mem := store.NewMemoryStore()
	svc := shortener.NewService(mem)
	r := testRouterWithService(svc)

	body := []byte(`{"long_url":"https://example.com/path"}`)

	firstReq := httptest.NewRequest(http.MethodPost, "/v1/links", bytes.NewReader(body))
	firstReq.Header.Set("Content-Type", "application/json")
	firstRec := httptest.NewRecorder()
	r.ServeHTTP(firstRec, firstReq)

	if firstRec.Code != http.StatusCreated {
		t.Fatalf("first create status = %d, want %d", firstRec.Code, http.StatusCreated)
	}

	var first CreateLinkResponse
	if err := json.Unmarshal(firstRec.Body.Bytes(), &first); err != nil {
		t.Fatalf("decode first response: %v", err)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/v1/links", bytes.NewReader(body))
	secondReq.Header.Set("Content-Type", "application/json")
	secondRec := httptest.NewRecorder()
	r.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusOK {
		t.Fatalf("duplicate create status = %d, want %d", secondRec.Code, http.StatusOK)
	}

	var second CreateLinkResponse
	if err := json.Unmarshal(secondRec.Body.Bytes(), &second); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if second.Code != first.Code {
		t.Fatalf("duplicate create should return same code, got %q and %q", first.Code, second.Code)
	}
}

func TestCreateLink_InvalidURLErrorShape(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	mem := store.NewMemoryStore()
	svc := shortener.NewService(mem)
	r := testRouterWithService(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/links",
		bytes.NewReader([]byte(`{"long_url":"ftp://example.com/file"}`)),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Code != "invalid_url" {
		t.Fatalf("error.code = %q, want %q", resp.Error.Code, "invalid_url")
	}
	if resp.Error.Message == "" {
		t.Fatalf("error.message should not be empty")
	}
}

func TestCreateLink_ExpiredRequestReturns400(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	mem := store.NewMemoryStore()
	svc := shortener.NewService(mem)
	r := testRouterWithService(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/links",
		bytes.NewReader([]byte(`{"long_url":"https://example.com","expires_at":"2020-01-01T00:00:00Z"}`)),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Code != "invalid_expiry" {
		t.Fatalf("error.code = %q, want %q", resp.Error.Code, "invalid_expiry")
	}
}

func TestRedirect_ExpiredLinkReturns404(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	mem := store.NewMemoryStore()
	svc := shortener.NewService(mem)
	r := testRouterWithService(svc)

	expiresAt := time.Now().UTC().Add(-time.Minute)
	if err := mem.Save(context.Background(), store.Link{
		Code:      "expiredCode",
		LongURL:   "https://example.com/expired",
		ExpiresAt: &expiresAt,
	}); err != nil {
		t.Fatalf("seed save failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/expiredCode", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Code != "not_found" {
		t.Fatalf("error.code = %q, want %q", resp.Error.Code, "not_found")
	}
}

func TestCreateLink_UnexpectedErrorReturnsGeneric500(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	svc := shortener.NewService(&failingStore{
		findByIntent: errors.New("db unavailable"),
	})
	r := testRouterWithService(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/links",
		bytes.NewReader([]byte(`{"long_url":"https://example.com/path"}`)),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Code != "internal" {
		t.Fatalf("error.code = %q, want %q", resp.Error.Code, "internal")
	}
	if resp.Error.Message != "something went wrong" {
		t.Fatalf("error.message = %q, want %q", resp.Error.Message, "something went wrong")
	}
}

func TestRedirect_UnexpectedErrorReturnsGeneric500(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	svc := shortener.NewService(&failingStore{
		getErr: errors.New("db unavailable"),
	})
	r := testRouterWithService(svc)

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Error.Code != "internal" {
		t.Fatalf("error.code = %q, want %q", resp.Error.Code, "internal")
	}
	if resp.Error.Message != "something went wrong" {
		t.Fatalf("error.message = %q, want %q", resp.Error.Message, "something went wrong")
	}
}
