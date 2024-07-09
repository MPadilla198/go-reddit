package reddit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// MultiService handles communication with the multireddit
// related methods of the Reddit API.
// Reddit API docs: https://www.reddit.com/dev/api#section_multis
type MultiService struct {
	client *Client
}

type MultiPostCopyOptions struct {
	DescriptionMarkdown string // raw Markdown text
	DisplayName         string // a string no longer than 50 characters
	ExpandSubreddits    bool
	From                string // multireddit url path
	To                  string // destination multireddit url path
}

func (opts *MultiPostCopyOptions) Params() url.Values {
	result := url.Values{}

	result.Add("description_md", opts.DescriptionMarkdown)
	result.Add("display_name", opts.DisplayName)
	result.Add("expand_srs", strconv.FormatBool(opts.ExpandSubreddits))
	result.Add("from", opts.From)
	result.Add("to", opts.To)

	return result
}

// PostCopy Copy a multi.
// Responds with 409 Conflict if the target already exists.
// A "copied from ..." line will automatically be appended to the description.
func (s *MultiService) PostCopy(ctx context.Context, modHash string, opts *MultiPostCopyOptions) (*http.Response, error) {

	path := "api/multi/copy" + opts.Params().Encode()

	req, err := s.client.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetMine Fetch a list of multis belonging to the current user.
func (s *MultiService) GetMine(ctx context.Context, expandSubreddits bool) (*http.Response, error) {
	path := fmt.Sprintf("api/multi/mine?expand_srs=%t", expandSubreddits)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)

}

// GetMultiOfUser Fetch a list of public multis belonging to username
func (s *MultiService) GetMultiOfUser(ctx context.Context, username string, expandSubreddits bool) (*http.Response, error) {
	path := fmt.Sprintf("api/multi/user/%s?expand_srs=%t", username, expandSubreddits)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// DeleteMulti Delete a multireddit.
func (s *MultiService) DeleteMulti(ctx context.Context, modHash, multiPath string, expandSubreddits, isFilter bool) (*http.Response, error) {
	name := "multi"
	if isFilter {
		name = "filter"
	}

	path := fmt.Sprintf("api/%s/%s?expand_srs=%t", name, multiPath, expandSubreddits)

	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetMulti Fetch a multis data and subreddit list by name.
func (s *MultiService) GetMulti(ctx context.Context, multiPath string, expandSubreddits, isFilter bool) (*http.Response, error) {
	name := "multi"
	if isFilter {
		name = "filter"
	}

	path := fmt.Sprintf("api/%s/%s?expand_srs=%t", name, multiPath, expandSubreddits)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type MultiIconImageType string

const (
	MultiIcoImagePNG  MultiIconImageType = "png"
	MultiIcoImageJPG  MultiIconImageType = "jpg"
	MultiIcoImageJPEG MultiIconImageType = "jpeg"
)

type MultiVisibilityType string

const (
	MultiVisibilityPrivate MultiVisibilityType = "private"
	MultiVisibilityPublic  MultiVisibilityType = "public"
	MultiVisibilityHidden  MultiVisibilityType = "hidden"
)

type Multi struct { // todo change name to Multi once all erroneous references to Multi are removed from project
	DescriptionMarkdown string             `json:"description_md"` // raw Markdown text
	DisplayName         string             `json:"display_name"`   // A string no longer than 50 characters
	IconIMG             MultiIconImageType `json:"icon_img"`
	KeyColor            string             `json:"key_color"` // a 6-digit rgb hex color, e.g. `#AABBCC`
	Subreddits          []struct {
		Name string `json:"name"` // subreddit name
	} `json:"subreddits"`
	Visibility MultiVisibilityType `json:"visibility"`
}

type MultiPathOptions struct {
	Model            Multi  `json:"model"`
	MultiPath        string `json:"multipath"` // multireddit url path
	ExpandSubreddits bool   `json:"expand_srs"`
}

// PostMulti Create a multi. Responds with 409 Conflict if it already exists.
func (s *MultiService) PostMulti(ctx context.Context, modHash string, isFilter bool, opts *MultiPathOptions) (*http.Response, error) {
	name := "multi"
	if isFilter {
		name = "filter"
	}
	path := fmt.Sprintf("api/%s/%s", name, opts.MultiPath)
	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PutMulti Create or update a multi.
func (s *MultiService) PutMulti(ctx context.Context, modHash string, isFilter bool, opts *MultiPathOptions) (*http.Response, error) {
	name := "multi"
	if isFilter {
		name = "filter"
	}

	path := fmt.Sprintf("api/%s/%s", name, opts.MultiPath)

	req, err := s.client.NewJSONRequest(http.MethodPut, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetMultiDescription get a multireddit's description.
func (s *MultiService) GetMultiDescription(ctx context.Context, multiPath string) (*http.Response, error) {
	path := fmt.Sprintf("api/multi/%s/description", multiPath)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PutMultiDescription Change a multi's markdown description.
func (s *MultiService) PutMultiDescription(ctx context.Context, modHash, multiPath, description string) (*http.Response, error) {
	data := struct {
		BodyMarkdown string `json:"body_md"` // raw Markdown text,
	}{BodyMarkdown: description}

	path := fmt.Sprintf("api/multi/%s/description", multiPath)

	req, err := s.client.NewJSONRequest(http.MethodPut, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// DeleteMultiSubreddit Remove a subreddit from a multi.
func (s *MultiService) DeleteMultiSubreddit(ctx context.Context, modHash, multiPath, subreddit string, isFilter bool) (*http.Response, error) {
	name := "multi"
	if isFilter {
		name = "filter"
	}

	path := fmt.Sprintf("api/%s/%s/r/%s", name, multiPath, subreddit)

	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetMultiSubreddit Get data about a subreddit in a multi.
func (s *MultiService) GetMultiSubreddit(ctx context.Context, multiPath, subreddit string, isFilter bool) (*http.Response, error) {
	name := "multi"
	if isFilter {
		name = "filter"
	}

	path := fmt.Sprintf("api/%s/%s/r/%s", name, multiPath, subreddit)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PutMultiSubreddit Add a subreddit to a multi.
func (s *MultiService) PutMultiSubreddit(ctx context.Context, modHash, multiPath, subreddit string, isFilter bool) (*http.Response, error) {
	data := struct {
		Name string `json:"name"`
	}{Name: subreddit}

	name := "multi"
	if isFilter {
		name = "filter"
	}

	path := fmt.Sprintf("api/%s/%s/r/%s", name, multiPath, subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPut, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}
