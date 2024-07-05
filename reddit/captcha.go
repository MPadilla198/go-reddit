package reddit

import (
	"context"
	"net/http"
)

// CaptchaService services reddit's captcha services
type CaptchaService struct {
	client *Client
}

// GetNeedsCaptcha Check whether ReCAPTCHAs are needed for API methods
func (s *CaptchaService) GetNeedsCaptcha(ctx context.Context) (*http.Response, error) {
	path := "api/needs_captcha"

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}
