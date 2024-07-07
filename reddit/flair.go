package reddit

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
)

// FlairService handles communication with the flair
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_flair
type FlairService struct {
	client *Client
}

type FlairType string

const (
	FlairTypeUser FlairType = "USER_FLAIR"
	FlairTypeLink FlairType = "LINK_FLAIR"
)

// PostClearFlairTemplates deletes all user flair templates.
func (s *FlairService) PostClearFlairTemplates(ctx context.Context, modHash, subreddit string, flairType FlairType) (*http.Response, error) {
	data := struct {
		APIType string    `json:"api_type"`
		Type    FlairType `json:"flair_type"`
	}{APIType: "json", Type: flairType}

	path := fmt.Sprintf("r/%s/api/clearflairtemplates", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostSubredditDeleteFlair Delete the flair of the user.
func (s *FlairService) PostSubredditDeleteFlair(ctx context.Context, modHash, subreddit, username string) (*http.Response, error) {
	data := struct {
		APIType string `json:"api_type"`
		Name    string `json:"name"`
	}{APIType: "json", Name: username}

	path := fmt.Sprintf("r/%s/api/deleteflair", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostSubredditDeleteFlairTemplate DeleteTemplate deletes the flair template via its id.
func (s *FlairService) PostSubredditDeleteFlairTemplate(ctx context.Context, modHash, subreddit, flairTemplateID string) (*http.Response, error) {
	data := struct {
		APIType         string `json:"api_type"`
		FlairTemplateID string `json:"flair_template_id"`
	}{APIType: "json", FlairTemplateID: flairTemplateID}

	path := fmt.Sprintf("r/%s/api/deleteflairtemplate", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type FlairSubredditOptions struct {
	APIType  string `json:"api_type"`
	CSSClass string `json:"css_class"` // a valid subreddit image name
	Link     string `json:"link"`      // fullname of a link
	Name     string `json:"name"`      // a user by name
	Text     string `json:"text"`      // a string no longer than 64 characters
}

func (s *FlairService) PostSubredditFlair(ctx context.Context, modHash, subreddit string, opts FlairSubredditOptions) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/flair", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PatchSubredditFlairTemplateOrder Update the order of flair templates in the specified subreddit.
// Order should contain every single flair id for that flair type; omitting any id will result in a loss of data.
func (s *FlairService) PatchSubredditFlairTemplateOrder(ctx context.Context, modHash, subreddit string, flairType FlairType) (*http.Response, error) {
	data := struct {
		Type      FlairType `json:"flair_type"`
		Subreddit string    `json:"subreddit"`
	}{Type: flairType, Subreddit: subreddit}

	path := fmt.Sprintf("r/%s/api/flair_template_order", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPatch, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type FlairPosition string

const (
	FlairPositionLeft  FlairPosition = "left"
	FlairPositionRight FlairPosition = "right"
)

type FlairConfigOptions struct {
	APIType                    string        `json:"api_type"`
	FlairEnabled               bool          `json:"flair_enabled"`
	Position                   FlairPosition `json:"flair_position"`
	FlairSelfAssignEnabled     bool          `json:"flair_self_assign_enabled"`
	LinkFlairPosition          FlairPosition `json:"link_flair_position"`
	LinkFlairSelfAssignEnabled bool          `json:"link_flair_self_assign_enabled"`
}

func (s *FlairService) PostSubredditFlairConfig(ctx context.Context, modHash, subreddit string, opts FlairConfigOptions) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/flairconfig", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostSubredditFlairCSV Change the flair of multiple users in the same subreddit with a single API call.
// Requires a string 'flair_csv' which has up to 100 lines of the form 'user,flairtext,cssclass' (Lines beyond the 100th are ignored).
// If both cssclass and flairtext are the empty string for a given user, instead clears that user's flair.
// Returns an array of objects indicating if each flair setting was applied, or a reason for the failure.
func (s *FlairService) PostSubredditFlairCSV(ctx context.Context, modHash, subreddit string, csvData [][]string) (*http.Response, error) {
	var csvResult string

	w := csv.NewWriter(bytes.NewBufferString(csvResult))
	err := w.WriteAll(csvData)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	w.Flush()

	data := struct {
		FlairCSV string `json:"flair_csv"`
	}{FlairCSV: csvResult}

	path := fmt.Sprintf("r/%s/api/flaircsv", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

func (s *FlairService) GetSubredditFlairList(ctx context.Context, subreddit string, opts ListingOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("r/%s/api/flairlist", subreddit)

	return s.client.getListing(ctx, path, opts)
}

type FlairSelectorOptions struct {
	IsNewLink bool   `json:"is_newlink,omitempty"`
	Link      string `json:"link,omitempty"` // fullname of a link
	Name      string `json:"name,omitempty"` // a user by name
}

// PostSubredditFlairSelector Return information about a user's flair options.
// If link is given, return link flair options for an existing link.
// If is_newlink is True, return link flairs options for a new link submission.
// Otherwise, return user flair options for this subreddit.
// The logged-in user's flair is also returned.
// subreddit moderators may give a user by name to instead retrieve that user's flair.
func (s *FlairService) PostSubredditFlairSelector(ctx context.Context, subreddit string, opts FlairSelectorOptions) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/flairselector", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type FlairTemplateOptions struct {
	APIType         string    `json:"api_type"`
	CSSClass        string    `json:"css_class"` // a valid subreddit name
	FlairTemplateID string    `json:"flair_template_id"`
	FlairType       FlairType `json:"flair_type"`
	Text            string    `json:"text"` // a string no longer than 64 characters
	TextEditable    bool      `json:"text_editable"`
}

// PostSubredditFlairTemplate Modify flair template for a subreddit.
func (s *FlairService) PostSubredditFlairTemplate(ctx context.Context, modHash, subreddit string, opts FlairTemplateOptions) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/flairtemplate", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type FlairAllowableContentType string

const (
	FlairAllowableContentAll   FlairAllowableContentType = "all"
	FlairAllowableContentEmoji FlairAllowableContentType = "emoji"
	FlairAllowableContentText  FlairAllowableContentType = "text"
)

type FlairTextColorType string

const (
	FlairTextColorLight FlairTextColorType = "light"
	FlairTextColorDark  FlairTextColorType = "dark"
)

type FlairTemplateV2Options struct {
	AllowableContent FlairAllowableContentType `json:"allowable_content"`
	APIType          string                    `json:"api_type"`
	BackgroundColor  string                    `json:"background_color"` // a 6-digit rgb hex color, e.g. #AABBCC
	CSSClass         string                    `json:"css_class"`        // a valid subreddit name
	FlairTemplateID  string                    `json:"flair_template_id"`
	FlairType        FlairType                 `json:"flair_type"`
	MaxEmojis        int                       `json:"max_emojis"` // an integer between 1 and 10 (default: 10)
	ModOnly          bool                      `json:"mod_only"`
	OverrideCSS      bool                      `json:"override_css"`
	Text             string                    `json:"text"` // a string no longer than 64 characters
	TextColor        FlairTextColorType        `json:"text_color"`
	TextEditable     bool                      `json:"text_editable"`
}

// PostSubredditFlairTemplateV2 Create or update a flair template.
// This new endpoint is primarily used for the redesign.
func (s *FlairService) PostSubredditFlairTemplateV2(ctx context.Context, modHash, subreddit string, opts FlairTemplateV2Options) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/flairtemplate_v2", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetSubredditLinkFlair Return list of available link flair for the current subreddit.
// Will not return flair if the user cannot set their own link flair and is not a moderator that can set flair.
func (s *FlairService) GetSubredditLinkFlair(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/link_flair", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetSubredditLinkFlairV2 Return list of available link flair for the current subreddit.
// Will not return flair if the user cannot set their own link flair and is not a moderator that can set flair.
func (s *FlairService) GetSubredditLinkFlairV2(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/link_flair_v2", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type FlairReturnRtsonType string

const (
	FlairReturnRtsonAll  FlairReturnRtsonType = "all"
	FlairReturnRtsonOnly FlairReturnRtsonType = "only"
	FlairReturnRtsonNone FlairReturnRtsonType = "none"
)

type FlairSubredditSelectOptions struct {
	APIType         string               `json:"api_type"`
	BackgroundColor string               `json:"background_color"`  // a 6-digit rgb hex color, e.g. #AABBCC
	CSSClass        string               `json:"css_class"`         // a valid subreddit name
	FlairTemplateId string               `json:"flair_template_id"` //
	Link            string               `json:"link"`              // a fullname of a link
	Name            string               `json:"name"`              // a user by name
	ReturnRtson     FlairReturnRtsonType `json:"return_rtson"`      // [all|only|none]: "all" saves attributes and returns rtjson "only" only returns rtjson"none" only saves attributes
	Text            string               `json:"text"`              // a string no longer than 64 characters
	TextColor       FlairTextColorType   `json:"text_color"`
}

func (s *FlairService) PostSubredditSelectFlair(ctx context.Context, modHash, subreddit string, opts FlairSubredditSelectOptions) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/selectflair", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

func (s *FlairService) PostSubredditSetFlairEnabled(ctx context.Context, modHash, subreddit string, flairEnabled bool) (*http.Response, error) {
	data := struct {
		APIType      string `json:"api_type"`
		FlairEnabled bool   `json:"flair_enabled"`
	}{APIType: "json", FlairEnabled: flairEnabled}

	path := fmt.Sprintf("r/%s/api/setflairenabled", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetSubredditUserFlair Return list of available user flair for the current subreddit.
// Will not return flair if flair is disabled on the subreddit, the user cannot set their own flair, or they are not a moderator that can set flair.
func (s *FlairService) GetSubredditUserFlair(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/user_flair", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetSubredditUserFlairV2 Return list of available user flair for the current subreddit.
// If user is not a mod of the subreddit, this endpoint filters out mod_only templates.
func (s *FlairService) GetSubredditUserFlairV2(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/user_flair_v2", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}
