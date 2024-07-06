package reddit

import (
	"context"
	"net/http"
	"strings"
)

// CollectionService handles communication with the collection
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_collections
type CollectionService struct {
	client *Client
}

// PostAddLinkToCollection Add a post to a collection
func (s *CollectionService) PostAddLinkToCollection(ctx context.Context, collectionID, linkFullname, modHash string) (*http.Response, error) {
	data := struct {
		CollectionID string `json:"collection_id"`
		LinkFullname string `json:"link_fullname"`
	}{CollectionID: collectionID, LinkFullname: linkFullname}

	path := "api/v1/collections/add_post_to_collection"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetCollection Fetch a collection including all the links
func (s *CollectionService) GetCollection(ctx context.Context, collectionID string, includeLinks bool) (*http.Response, error) {
	data := struct {
		CollectionID string `json:"collection_id"`
		IncludeLinks bool   `json:"include_links"`
	}{CollectionID: collectionID, IncludeLinks: includeLinks}

	path := "api/v1/collections/collection"

	req, err := s.client.NewJSONRequest(http.MethodGet, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type CollectionDisplayLayout string

const (
	CollectionDisplayLayoutTimeline CollectionDisplayLayout = "timeline"
	CollectionDisplayLayoutGallery  CollectionDisplayLayout = "gallery"
)

type CreateCollectionOptions struct {
	Description   string                  `json:"description"`    // a string no longer than 500 characters
	DisplayLayout CollectionDisplayLayout `json:"display_layout"` //
	SRFullname    string                  `json:"sr_fullname"`    // a fullname of a subreddit
	Title         string                  `json:"title"`          // title of the submission. up to 300 characters long
}

// PostCreateCollection Create a collection.
func (s *CollectionService) PostCreateCollection(ctx context.Context, modHash string, createRequest *CreateCollectionOptions) (*http.Response, error) {
	path := "api/v1/collections/create_collection"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, createRequest)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostDeleteCollection Delete a collection via its id.
func (s *CollectionService) PostDeleteCollection(ctx context.Context, modHash string, collectionID string) (*http.Response, error) {
	data := struct {
		CollectionID string `json:"collection_id"`
	}{CollectionID: collectionID}

	path := "api/v1/collections/delete_collection"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostFollowCollection Follow a collection.
func (s *CollectionService) PostFollowCollection(ctx context.Context, modHash, collectionID string, follow bool) (*http.Response, error) {
	data := struct {
		CollectionID string `json:"collection_id"`
		Follow       bool   `json:"follow"`
	}{CollectionID: collectionID, Follow: follow}

	path := "api/v1/collections/follow_collection"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostRemoveLink Remove a post from a collection
func (s *CollectionService) PostRemoveLink(ctx context.Context, modHash, collectionID, linkFullname string) (*http.Response, error) {
	data := struct {
		CollectionID string `json:"collection_id"`
		LinkFullname string `json:"link_fullname"`
	}{CollectionID: collectionID, LinkFullname: linkFullname}

	path := "api/v1/collections/remove_post_in_collection"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// ReorderPosts reorders posts in a collection.
func (s *CollectionService) ReorderPosts(ctx context.Context, modHash, collectionID string, linkIDs ...string) (*http.Response, error) {
	data := struct {
		CollectionID string `json:"collection_id"`
		LinkIDs      string `json:"link_ids"`
	}{CollectionID: collectionID, LinkIDs: strings.Join(linkIDs, ",")}

	path := "api/v1/collections/reorder_collection"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetSubredditCollections Fetch collections for the subreddit
func (s *CollectionService) GetSubredditCollections(ctx context.Context, id string) (*http.Response, error) {
	data := struct {
		SRFullname string `json:"sr_fullname"`
	}{SRFullname: id}

	path := "api/v1/collections/subreddit_collections"

	req, err := s.client.NewJSONRequest(http.MethodGet, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostUpdateCollectionDescription updates a collection's description.
func (s *CollectionService) PostUpdateCollectionDescription(ctx context.Context, modHash, collectionID, description string) (*http.Response, error) {
	data := struct {
		CollectionID string `json:"collection_id"`
		Description  string `json:"description"`
	}{CollectionID: collectionID, Description: description}

	path := "api/v1/collections/update_collection_description"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostUpdateCollectionLayoutGallery updates a collection's layout to the gallery format.
func (s *CollectionService) PostUpdateCollectionLayoutGallery(ctx context.Context, modHash, collectionID string, layout CollectionDisplayLayout) (*http.Response, error) {
	data := struct {
		CollectionID  string                  `json:"collection_id"`
		DisplayLayout CollectionDisplayLayout `json:"display_layout"`
	}{CollectionID: collectionID, DisplayLayout: layout}

	path := "api/v1/collections/update_collection_display_layout"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostUpdateCollectionTitle updates a collection's title.
func (s *CollectionService) PostUpdateCollectionTitle(ctx context.Context, modHash, collectionID, title string) (*http.Response, error) {
	data := struct {
		CollectionID string `json:"collection_id"`
		Title        string `json:"title"`
	}{CollectionID: collectionID, Title: title}

	path := "api/v1/collections/update_collection_title"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}
