package reddit

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// CommentService handles communication with the comment
// related methods of the Reddit API.
type CommentService service

func (s *CommentService) isCommentID(id string) bool {
	return strings.HasPrefix(id, kindComment+"_")
}

// Submit submits a comment as a reply to a post or to another comment.
func (s *CommentService) Submit(ctx context.Context, id string, text string) (*Comment, *Response, error) {
	path := "api/comment"

	form := url.Values{}
	form.Set("api_type", "json")
	form.Set("return_rtjson", "true")
	form.Set("parent", id)
	form.Set("text", text)

	req, err := s.client.NewPostForm(path, form)
	if err != nil {
		return nil, nil, err
	}

	root := new(Comment)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root, resp, nil
}

// Edit edits the comment with the id provided.
// todo: don't forget to do this for posts
func (s *CommentService) Edit(ctx context.Context, id string, text string) (*Comment, *Response, error) {
	if !s.isCommentID(id) {
		return nil, nil, fmt.Errorf("must provide comment id (starting with %s_); id provided: %q", kindComment, id)
	}

	path := "api/editusertext"

	form := url.Values{}
	form.Set("api_type", "json")
	form.Set("return_rtjson", "true")
	form.Set("thing_id", id)
	form.Set("text", text)

	req, err := s.client.NewPostForm(path, form)
	if err != nil {
		return nil, nil, err
	}

	root := new(Comment)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root, resp, nil
}

// Delete deletes a comment via the id.
// todo: don't forget to do this for posts.
func (s *CommentService) Delete(ctx context.Context, id string) (*Response, error) {
	if !s.isCommentID(id) {
		return nil, fmt.Errorf("must provide comment id (starting with %s_); id provided: %q", kindComment, id)
	}

	path := "api/del"

	form := url.Values{}
	form.Set("id", id)

	req, err := s.client.NewPostForm(path, form)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// Save saves a comment.
func (s *CommentService) Save(ctx context.Context, id string) (*Response, error) {
	if !s.isCommentID(id) {
		return nil, fmt.Errorf("must provide comment id (starting with %s_); id provided: %q", kindComment, id)
	}

	path := "api/save"

	form := url.Values{}
	form.Set("id", id)

	req, err := s.client.NewPostForm(path, form)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// Unsave unsaves a comment.
func (s *CommentService) Unsave(ctx context.Context, id string) (*Response, error) {
	if !s.isCommentID(id) {
		return nil, fmt.Errorf("must provide comment id (starting with t1_); id provided: %q", id)
	}

	path := "api/unsave"

	form := url.Values{}
	form.Set("id", id)

	req, err := s.client.NewPostForm(path, form)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
