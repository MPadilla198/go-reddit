package reddit

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// ListingsService handles communication with the listing
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_listings
type ListingsService struct {
	client *Client
}

func (s *ListingsService) GetBest(ctx context.Context, opts *ListingOptions) (*Listing, *http.Response, error) {
	return s.client.getListing(ctx, "best", opts)
}

// GetNamesByIDs Get a listing of links by fullname.
// names is a list of fullnames for links separated by commas or spaces.
func (s *ListingsService) GetNamesByIDs(ctx context.Context, fullnames ...string) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("by_id/%s", strings.Join(fullnames, ","))
	return s.client.getListing(ctx, path, nil)
}

type ListingsCommentThemeType string

const (
	ListingsCommentThemeDefault ListingsCommentThemeType = "default"
	ListingsCommentThemeDark    ListingsCommentThemeType = "dark"
)

type ListingsLinkCommentsOptions struct {
	Article   string                            `json:"article"`           // ID36 of a link
	Comment   string                            `json:"comment,omitempty"` // (optional) ID336 of a comment
	Context   int                               `json:"context"`           // an integer between 0 and 8
	Depth     int                               `json:"depth,omitempty"`
	Limit     int                               `json:"limit,omitempty"`
	ShowEdits bool                              `json:"showedits"`
	ShowMedia bool                              `json:"showmedia"`
	ShowMore  bool                              `json:"showmore"`
	ShowTitle bool                              `json:"showtitle"`
	Sort      SubredditSuggestedCommentSortType `json:"sort"`
	SRDetail  string                            `json:"sr_detail,omitempty"` // expand subreddits
	Theme     ListingsCommentThemeType          `json:"theme"`
	Threaded  bool                              `json:"threaded"`
	Truncate  int                               `json:"truncate"` // an integer between 0 and 50
}

// GetSubredditCommentsForLink Get the comment tree for a given Link article.
// If supplied, comment is the ID36 of a comment in the comment tree for article.
// This comment will be the (highlighted) focal point of the returned view and context will be the number of parents shown.
// depth is the maximum depth of subtrees in the thread.
// limit is the maximum number of comments to return.
// See also: /api/morechildren and /api/comment.
func (s *ListingsService) GetSubredditCommentsForLink(ctx context.Context, subreddit, article string, opts *ListingsLinkCommentsOptions) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/comments/%s", subreddit, article)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetDuplicateLinks Return a list of other submissions of the same URL
func (s *ListingsService) GetDuplicateLinks(ctx context.Context, article string, opts *ListingDuplicateOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("duplicates/%s", article)
	return s.client.getListing(ctx, path, opts)
}

type ListingsSubredditSortType string

const (
	ListingsSubredditSortHot           ListingsSubredditSortType = "hot"
	ListingsSubredditSortNew           ListingsSubredditSortType = "new"
	ListingsSubredditSortRising        ListingsSubredditSortType = "rising"
	ListingsSubredditSortTop           ListingsSubredditSortType = "top"
	ListingsSubredditSortControversial ListingsSubredditSortType = "controversial"
)

func (s *ListingsService) GetSubredditSorted(ctx context.Context, subreddit string, sort ListingsSubredditSortType, opts *ListingSubredditSortOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("r/%s/%s", subreddit, sort)

	return s.client.getListing(ctx, path, opts)
}
