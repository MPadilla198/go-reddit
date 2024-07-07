package reddit

import (
	"context"
	"fmt"
	"net/http"
)

// MessageService handles communication with the message
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_messages
type MessageService struct {
	client *Client
}

// SendMessageRequest represents a request to send a message.
type SendMessageRequest struct {
	// Username, or /r/name for that subreddit's moderators.
	To      string `url:"to"`
	Subject string `url:"subject"`
	Text    string `url:"text"`
	// Optional. If specified, the message will look like it came from the subreddit.
	FromSubreddit string `url:"from_sr,omitempty"`
}

// PostBlock For blocking the author of a thing via inbox. Only accessible to approved OAuth applications
func (s *MessageService) PostBlock(ctx context.Context, modHash, fullname string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // Fullname of a thing
	}{ID: fullname}

	path := "api/block"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostMessageCollapse Collapse a message
// See also: /api/uncollapse_message
func (s *MessageService) PostMessageCollapse(ctx context.Context, modHash string, ids ...string) (*http.Response, error) {
	data := struct {
		IDs []string `json:"id"` // A comma-separated list of thing fullnames
	}{IDs: ids}

	path := "api/collapse_message"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type MessageComposeOptions struct {
	APIType            string `json:"api_type"`
	FromSR             string `json:"from_sr"` // A subreddit name
	GRecaptchaResponse string `json:"g_recaptcha_response"`
	Subject            string `json:"subject"` //
	Text               string `json:"text"`    // raw Markdown text
	To                 string `json:"to"`      // the name of an existing user
}

// PostMessageCompose Handles message composition under /message/compose.
func (s *MessageService) PostMessageCompose(ctx context.Context, modHash string, opts *MessageComposeOptions) (*http.Response, error) {
	path := "api/compose"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostMessageDelete Delete messages from the recipient's view of their inbox.
func (s *MessageService) PostMessageDelete(ctx context.Context, modHash, id string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // A thing fullnames
	}{ID: id}

	path := "api/del_msg"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostReadAllMessages Queue up marking all messages for a user as read.
// This may take some time, and returns 202 to acknowledge acceptance of the request.
func (s *MessageService) PostReadAllMessages(ctx context.Context, modHash string, filterTypes ...string) (*http.Response, error) {
	data := struct {
		FilterTypes []string `json:"filter_types"` // A comma-separated list of items
	}{FilterTypes: filterTypes}

	path := "api/read_all_messages"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostReadMessages Read marks a message/comment as read via its full ID.
func (s *MessageService) PostReadMessages(ctx context.Context, modHash string, ids ...string) (*http.Response, error) {
	data := struct {
		ID []string `json:"id"` // A comma-separated list of thing fullnames
	}{ID: ids}

	path := "api/read_message"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

func (s *MessageService) PostUnblock(ctx context.Context, modHash, fullname string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // A thing fullname
	}{ID: fullname}

	path := "api/unblock_subreddit"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostUncollapseMessages Uncollapse a message
// See also: /api/collapse_message
func (s *MessageService) PostUncollapseMessages(ctx context.Context, modHash string, ids ...string) (*http.Response, error) {
	data := struct {
		ID []string `json:"id"` // A comma-separated list of thing fullnames
	}{ID: ids}

	path := "api/uncollapse_message"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostUnreadMessages marks a message/comment as unread via its full ID.
func (s *MessageService) PostUnreadMessages(ctx context.Context, modHash string, ids ...string) (*http.Response, error) {
	data := struct {
		ID []string `json:"id"` // A comma-separated list of thing fullnames
	}{ID: ids}

	path := "api/unread_message"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type MessagesWhereType string

const (
	MessagesWhereInbox  MessagesWhereType = "inbox"
	MessagesWhereUnread MessagesWhereType = "unread"
	MessagesWhereSent   MessagesWhereType = "sent"
)

func (s *MessageService) GetMessageWhere(ctx context.Context, where MessagesWhereType, opts *ListingMessageOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("message/%s", where)

	return s.client.getListing(ctx, path, opts)
}
