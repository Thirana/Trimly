package httpapi

type CreateLinkRequest struct {
	LongURL string `json:"long_url" binding:"required"`
}

type CreateLinkResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

type RedirectURI struct {
	Code string `uri:"code" binding:"required"`
}
