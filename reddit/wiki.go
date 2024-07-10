package reddit

import (
	"context"
	"fmt"
	"net/http"
)

// WikiService handles communication with the wiki
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_wiki
type WikiService struct {
	client *Client
}

type WikiAllowEditorAct string

const (
	WikiAllowEditorActDelete WikiAllowEditorAct = "del"
	WikiAllowEditorActAdd    WikiAllowEditorAct = "add"
)

// PostAllowEditor Allow/deny username to edit this wiki page
func (s *WikiService) PostAllowEditor(ctx context.Context, modHash, subreddit, page, username string, act WikiAllowEditorAct) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/wiki/alloweditor/%s?page=%s&username=%s", subreddit, act, page, username)

	req, err := s.client.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Set("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type WikiPostEditOptions struct {
	Content  string    `json:"content"`
	Page     string    `json:"page"`     // the name of an existing page or a new page to create
	Previous string    `json:"previous"` // the starting point revision for this edit
	Reason   [256]byte `json:"reason"`   // a string up to 256 characters long, consisting of printable characters.
}

// PostEdit Edit a wiki page.
func (s *WikiService) PostEdit(ctx context.Context, modHash, subreddit string, opts *WikiPostEditOptions) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/wiki/edit", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Set("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostHide Toggle the public visibility of a wiki page revision
func (s *WikiService) PostHide(ctx context.Context, modHash, subreddit, page, revisionID string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/wiki/hide?page=%s&revision=%s", subreddit, page, revisionID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Set("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostRevert Revert a wiki page to revision
func (s *WikiService) PostRevert(ctx context.Context, modHash, subreddit, page, revisionID string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/wiki/revert?page=%s&revision=%s", subreddit, page, revisionID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Set("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetDiscussionsPage Retrieve a list of discussions about this wiki page
// This endpoint is a listing.
func (s *WikiService) GetDiscussionsPage(ctx context.Context, subreddit, page string, opts *ListingOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("r/%s/wiki/discussions/%s", subreddit, page)

	return s.client.getListing(ctx, path, opts)
}

// GetPages Retrieve a list of wiki pages in this subreddit
func (s *WikiService) GetPages(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/wiki/pages", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetRevisions Retrieve a list of recently changed wiki pages in this subreddit
func (s *WikiService) GetRevisions(ctx context.Context, subreddit string, opts *ListingOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("r/%s/wiki/revisions", subreddit)

	return s.client.getListing(ctx, path, opts)
}

// GetRevisionsPage Retrieve a list of revisions of this wiki page
// This endpoint is a listing.
func (s *WikiService) GetRevisionsPage(ctx context.Context, subreddit, page string, opts *ListingOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("r/%s/wiki/revisions/%s", subreddit, page)

	return s.client.getListing(ctx, path, opts)
}

// GetSettingsPage Retrieve the current permission settings for page
func (s *WikiService) GetSettingsPage(ctx context.Context, subreddit, page string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/wiki/settings/%s", subreddit, page)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

func (s *WikiService) PostSettingsPage(ctx context.Context, modHash, subreddit, page string, permLevel int, listed bool) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/wiki/settings/%s?permlevel=%d&listed=%t", subreddit, page, permLevel, listed)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Set("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetPage Return the content of a wiki page
// If v is given, show the wiki page as it was at that version If both v and v2 are given, show a diff of the two
func (s *WikiService) GetPage(ctx context.Context, subreddit, page, v, v2 string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/wiki/%s?v=%d&v2=%t", subreddit, page, v, v2)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}
