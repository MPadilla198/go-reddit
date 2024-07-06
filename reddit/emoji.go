package reddit

import (
	"context"
	"fmt"
	"net/http"
)

// EmojiService handles communication with the emoji
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_emoji
type EmojiService struct {
	client *Client
}

type EmojiSubredditOptions struct {
	ModFlairOnly     bool   `json:"mod_flair_only"`     //
	Name             string `json:"name"`               // Name of the emoji to be created. It can be alphanumeric without any special characters except '-' & '_' and cannot exceed 24 characters.
	PostFlairAllowed bool   `json:"post_flair_allowed"` //
	S3Key            string `json:"s3_key"`             // S3 key of the uploaded image which can be obtained from the S3 url. This is of the form subreddit/hash_value.
	UserFlairAllowed bool   `json:"user_flair_allowed"` //
}

// PostSubredditEmoji Add an emoji to the DB by posting a message on emoji_upload_q.
// A job processor that listens on a queue, uses the s3_key provided in the request to locate the image in S3 Temp Bucket and moves it to the PERM bucket.
// It also adds it to the DB using name as the column and sr_fullname as the key and sends the status on the websocket URL that is provided as part of this response.
// This endpoint should also be used to update custom subreddit emojis with new images.
// If only the permissions on an emoji require updating the POST_emoji_permissions endpoint should be requested, instead.
func (s *EmojiService) PostSubredditEmoji(ctx context.Context, subreddit string, opts EmojiSubredditOptions) (*http.Response, error) {
	path := fmt.Sprintf("api/v1/%s/emoji.json", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// DeleteSubredditEmoji Delete a subreddit emoji. Remove the emoji from Cassandra and purge the assets from S3 and the image resizing provider.
func (s *EmojiService) DeleteSubredditEmoji(ctx context.Context, subreddit, emojiName string) (*http.Response, error) {
	path := fmt.Sprintf("api/v1/%s/emoji/%s", subreddit, emojiName)

	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostSubredditEmojiAssetUploadS3 Acquire and return an upload lease to s3 temp bucket.
// The return value of this function is a json object containing credentials for uploading assets to S3 bucket, S3 url for upload request and the key to use for uploading.
// Using this lease the client will upload the emoji image to S3 temp bucket (included as part of the S3 URL).
// This lease is used by S3 to verify that the upload is authorized.
func (s *EmojiService) PostSubredditEmojiAssetUploadS3(ctx context.Context, subreddit, filePath, mimeType string) (*http.Response, error) {
	data := struct {
		Filepath string `json:"filepath"`
		MIMEType string `json:"mimetype"`
	}{Filepath: filePath, MIMEType: mimeType}

	path := fmt.Sprintf("api/v1/%s/emoji_asset_upload_s3.json", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostSubredditEmojiCustomSize Set custom emoji size.
// Omitting width or height will disable custom emoji sizing.
func (s *EmojiService) PostSubredditEmojiCustomSize(ctx context.Context, subreddit string, height, width int) (*http.Response, error) {
	data := struct {
		Height int `json:"height"` // an integer between 1 and 40 (default: 0)
		Width  int `json:"width"`  // an integer between 1 and 40 (default: 0)
	}{Height: height, Width: width}

	path := fmt.Sprintf("api/v1/%s/emoji_custom_size", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetSubredditEmojiAll Get all emojis for a SR.
// The response includes snoomojis as well as emojis for the SR specified in the request.
// The response has 2 keys: - snoomojis - SR emojis
func (s *EmojiService) GetSubredditEmojiAll(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("api/v1/%s/emojis/all", subreddit)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}
