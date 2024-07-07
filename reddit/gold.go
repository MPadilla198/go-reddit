package reddit

import (
	"context"
	"fmt"
	"net/http"
)

// GoldService handles communication with the gold
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_gold
type GoldService struct {
	client *Client
}

// PostGild the post or comment via its full ID.
// This requires you to own Reddit coins and will consume them.
func (s *GoldService) PostGild(ctx context.Context, fullname string) (*http.Response, error) {
	path := fmt.Sprintf("api/v1/gold/gild/%s", fullname)

	req, err := s.client.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostGive the user between 1 and 36 (inclusive) months of gold.
// This requires you to own Reddit coins and will consume them.
func (s *GoldService) PostGive(ctx context.Context, username string, months int) (*http.Response, error) {
	data := struct {
		Username string `json:"username"` // A valid, existing reddit username
		Months   int    `json:"months"`   // an integer between 1 and 36
	}{Username: username, Months: months}

	path := fmt.Sprintf("api/v1/gold/give/%s", username)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}
