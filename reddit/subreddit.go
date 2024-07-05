package reddit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// SubredditService handles communication with the subreddit
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_subreddits
type SubredditService struct {
	client *Client
}

type rootSubredditNames struct {
	Names []string `json:"names,omitempty"`
}

// Relationship holds information about a relationship (friend/blocked).
type Relationship struct {
	ID      string     `json:"rel_id,omitempty"`
	User    string     `json:"name,omitempty"`
	UserID  string     `json:"id,omitempty"`
	Created *Timestamp `json:"date,omitempty"`
}

// Moderator is a user who moderates a subreddit.
type Moderator struct {
	*Relationship
	Permissions []string `json:"mod_permissions"`
}

// Ban represents a banned relationship.
type Ban struct {
	*Relationship
	// nil means the ban is permanent
	DaysLeft *int   `json:"days_left"`
	Note     string `json:"note,omitempty"`
}

/**
SUBREDDIT CONSTANTS FOR ADMIN OPTIONS
*/

type SubredditMediaType string

const (
	Giphy      SubredditMediaType = "giphy"
	Unknown    SubredditMediaType = "unknown"
	Animated   SubredditMediaType = "animated"
	Static     SubredditMediaType = "static"
	Expression SubredditMediaType = "expression"
)

type SubredditLinkType string

const (
	SubredditAny  SubredditLinkType = "any"
	SubredditLink SubredditLinkType = "link"
	SubredditSelf SubredditLinkType = "self"
)

type SubredditSpamLevel string

const (
	SubredditSpamLow  SubredditSpamLevel = "low"
	SubredditSpamHigh SubredditSpamLevel = "high"
	SubredditSpamAll  SubredditSpamLevel = "all"
)

type SubredditDiscoveryType string

const (
	SubredditDiscoveryUnknown    SubredditDiscoveryType = "unknown"
	SubredditDiscoveryOnboarding SubredditDiscoveryType = "onboarding"
)

type SubredditSuggestedCommentSortType string

const (
	SubredditCommentSortConfidence    SubredditSuggestedCommentSortType = "confidence"
	SubredditCommentSortTop           SubredditSuggestedCommentSortType = "top"
	SubredditCommentSortNew           SubredditSuggestedCommentSortType = "new"
	SubredditCommentSortControversial SubredditSuggestedCommentSortType = "controversial"
	SubredditCommentSortOld           SubredditSuggestedCommentSortType = "old"
	SubredditCommentSortRandom        SubredditSuggestedCommentSortType = "random"
	SubredditCommentSortQA            SubredditSuggestedCommentSortType = "qa"
	SubredditCommentSortLive          SubredditSuggestedCommentSortType = "live"
)

type SubredditType string

const (
	SubredditTypeGoldRestricted SubredditType = "gold_restricted"
	SubredditTypeArchived       SubredditType = "archived"
	SubredditTypeRestricted     SubredditType = "restricted"
	SubredditTypePrivate        SubredditType = "private"
	SubredditTypeEmployeesOnly  SubredditType = "employees_only"
	SubredditTypeGoldOnly       SubredditType = "gold_only"
	SubredditTypePublic         SubredditType = "public"
	SubredditTypeUser           SubredditType = "user"
)

type SubredditWikiMode string

const (
	SubredditWikiModeDisabled SubredditWikiMode = "disabled"
	SubredditWikiModeModonly  SubredditWikiMode = "modonly"
	SubredditWikiModeAnyone   SubredditWikiMode = "anyone"
)

// SubredditAdminOptions are a subreddit's settings.
type SubredditAdminOptions struct {
	AcceptFollowers             bool   `json:"accept_followers"`
	AdminOverrideSpamComments   bool   `json:"admin_override_spam_comments"`
	AdminOverrideSpamLinks      bool   `json:"admin_override_spam_links"`
	AdminOverrideSelfposts      bool   `json:"admin_override_selfposts"`
	AllOriginalContent          bool   `json:"all_original_content"`
	AllowChatPostCreation       bool   `json:"allow_chat_post_creation"`
	AllowDiscovery              bool   `json:"allow_discovery"`
	AllowGalleries              bool   `json:"allow_galleries"`
	AllowImages                 bool   `json:"allow_images"`
	AllowPolls                  bool   `json:"allow_polls"`
	AllowPostCrossposts         bool   `json:"allow_post_crossposts"`
	AllowPredictionContributors bool   `json:"allow_prediction_contributors"`
	AllowPredictions            bool   `json:"allow_predictions"`
	AllowPredictionsTournament  bool   `json:"allow_predictions_tournament"`
	AllowTalks                  bool   `json:"allow_talks"`
	AllowTop                    bool   `json:"allow_top"`
	AllowVideos                 bool   `json:"allow_videos"`
	APIType                     string `json:"api_type"`
	CollapseDeletedComments     bool   `json:"collapse_deleted_comments"`
	CommentContributionSettings struct {
		AllowedMediaTypes []SubredditMediaType `json:"allowed_media_types"`
	} `json:"comment_contribution_settings"`
	CommentScoreHideMins            int                `json:"comment_score_hide_mins"`  // an integer between 0 and 1440 (default: 0)
	CrowdControlChatLevel           int                `json:"crowd_control_chat_level"` // an integer between 0 and 3
	CrowdControlFilter              bool               `json:"crowd_control_filter"`     //
	CrowdControlLevel               int                `json:"crowd_control_level"`      // an integer between 0 and 3
	CrowdControlMode                bool               `json:"crowd_control_mode"`       //
	CrowdControlPostLevel           int                `json:"crowd_control_post_level"` // an integer between 0 and 3
	Description                     string             `json:"description"`
	DisableContributorRequests      bool               `json:"disable_contributor_requests"`
	ExcludeBannedModqueue           bool               `json:"exclude_banned_modqueue"`
	FreeFormReports                 bool               `json:"free_form_reports"`
	HatefulContentThresholdAbuse    int                `json:"hateful_content_threshold_abuse"`    // an integer between 0 and 3
	HatefulContentThresholdIdentity int                `json:"hateful_content_threshold_identity"` // an integer between 0 and 3
	HeaderTitle                     string             `json:"header-title"`
	HideAds                         bool               `json:"hide_ads"`
	KeyColor                        string             `json:"key_color"` // a 6-digit rgb hex color, e.g. #AABBCC
	LinkType                        SubredditLinkType  `json:"link_type"`
	ModmailHarassmentFilterEnabled  bool               `json:"modmail_harassment_filter_enabled"`
	Name                            string             `json:"name"` // subreddit name
	NewPinnedPostPNSEnabled         bool               `json:"new_pinned_post_pns_enabled"`
	OriginalContentTagEnabled       bool               `json:"original_content_tag_enabled"`
	Over18                          bool               `json:"over_18"`
	PredictionLeaderboardEntryType  int                `json:"prediction_leaderboard_entry_type"` // an integer between 0 and 2
	PublicDescription               string             `json:"public_description"`
	RestrictCommenting              bool               `json:"restrict_commenting"`
	RestrictPosting                 bool               `json:"restrict_posting"`
	ShouldArchivePosts              bool               `json:"should_archive_posts"`
	ShowMedia                       bool               `json:"show_media"`
	ShowMediaPreview                bool               `json:"show_media_preview"`
	SpamComments                    SubredditSpamLevel `json:"spam_comments"`
	SpamLinks                       SubredditSpamLevel `json:"spam_links"`
	SpamSelfposts                   SubredditSpamLevel `json:"spam_selfposts"`
	SpoilersEnabled                 bool               `json:"spoilers_enabled"`
	SR                              string             `json:"sr"` // full-name of a thing If sr is specified, the request will attempt to modify the specified subreddit. If not, a subreddit with name "name" will be created.
	SubmitLinkLabel                 string             `json:"submit_link_label"`
	SubmitText                      string             `json:"submit_text"`
	SubmitTextLabel                 string             `json:"submit_text_label"` // No longer than 60 characters
	SubredditDiscoverySettings      struct {
		DisabledDiscoveryTypes []SubredditDiscoveryType `json:"disabled_discovery_types"`
	} `json:"subreddit_discovery_settings"`
	SuggestedCommentSort       SubredditSuggestedCommentSortType `json:"suggested_comment_sort"`
	Title                      string                            `json:"title"`                         // a string of no longer than 100 characters
	ToxicityThresholdChatLevel int                               `json:"toxicity_threshold_chat_level"` // an integer either 0 or 1
	Type                       SubredditType                     `json:"type"`
	UserFlairPNSEnabled        bool                              `json:"user_flair_pns_enabled"`
	WelcomeMessageEnabled      bool                              `json:"welcome_message_enabled"`
	WelcomeMessageText         string                            `json:"welcome_message_text"`
	WikiEditAge                int                               `json:"wiki_edit_age"`   // an integer between 0 and 36600 (default: 0)
	WikiEditKarma              int                               `json:"wiki_edit_karma"` // an integer between 0 and 1000000000 (default: 0)
	WikiMode                   SubredditWikiMode                 `json:"wiki_mode"`
}

type SubredditAboutWhere string

const (
	SubredditAboutWhereBanned           SubredditAboutWhere = "banned"
	SubredditAboutWhereMuted            SubredditAboutWhere = "muted"
	SubredditAboutWhereWikibanned       SubredditAboutWhere = "wikibaned"
	SubredditAboutWhereContributors     SubredditAboutWhere = "contributors"
	SubredditAboutWhereWikicontributors SubredditAboutWhere = "wikicontributors"
	SubredditAboutWhereModerators       SubredditAboutWhere = "moderators"
)

// GetAboutWhere gets a subreddit by name .
func (s *SubredditService) GetAboutWhere(ctx context.Context, subreddit string, w SubredditAboutWhere, opts ListingSubredditOptions) (*Listing, *http.Response, error) {
	if subreddit == "" {
		return nil, nil, &InternalError{"subreddit cannot be empty"}
	}
	path := fmt.Sprintf("r/%s/about/%s", subreddit, w)
	return s.client.getListing(ctx, path, opts)
}

// DeleteSubredditBanner Remove the subreddit's custom mobile banner.
func (s *SubredditService) DeleteSubredditBanner(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/delete_sr_banner", subreddit)

	form := url.Values{}
	form.Set("api_type", "json") // TODO MODHASH

	return s.client.PostURL(ctx, path, []byte(form.Encode()))
}

// DeleteSubredditHeader Remove the subreddit's custom header image.
// The sitewide-default header image will be shown again after this call.
func (s *SubredditService) DeleteSubredditHeader(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/delete_sr_header", subreddit)

	form := url.Values{}
	form.Set("api_type", "json") // TODO MODHASH

	return s.client.PostURL(ctx, path, []byte(form.Encode()))
}

// DeleteSubredditIcon Remove the subreddit's custom mobile icon.
func (s *SubredditService) DeleteSubredditIcon(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/delete_sr_icon", subreddit)

	form := url.Values{}
	form.Set("api_type", "json") // TODO MODHASH

	return s.client.PostURL(ctx, path, []byte(form.Encode()))
}

// DeleteSubredditImage Remove an image from the subreddit's custom image set.
// The image will no longer count against the subreddit's image limit.
// However, the actual image data may still be accessible for an unspecified amount of time.
// If the image is currently referenced by the subreddit's stylesheet, that stylesheet will no longer validate and won't be editable until the image reference is removed.
func (s *SubredditService) DeleteSubredditImage(ctx context.Context, subreddit, imageName string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/delete_sr_img", subreddit)

	form := url.Values{}
	form.Set("api_type", "json") // TODO MODHASH
	form.Set("img_name", imageName)

	return s.client.PostURL(ctx, path, []byte(form.Encode()))
}

type SubredditSearchOptions struct {
	Exact                 bool   `url:"exact"`
	IncludeOver18         bool   `url:"include_over_18"`
	IncludeUnadvertisable bool   `url:"include_unadvertisable"`
	Query                 string `url:"query"`
	SearchQueryID         string `url:"search_query_id"`
	TypeaheadActive       bool   `url:"typeahead_active"`
}

// SearchNames List subreddit names that begin with a query string.
// Subreddits whose names begin with query will be returned.
// If include_over_18 is false, subreddits with over-18 content restrictions will be filtered from the results.
// If include_unadvertisable is False, subreddits that have hide_ads set to True or are on the anti_ads_subreddits list will be filtered.
// If exact is true, only an exact match will be returned.
// Exact matches are inclusive of over_18 subreddits, but not hide_ad subreddits when include_unadvertisable is False.
func (s *SubredditService) SearchNames(ctx context.Context, opts SubredditSearchOptions) ([]string, *http.Response, error) {
	path, err := addOptions("api/search_reddit_names", opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(rootSubredditNames)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Names, resp, nil
}

// SearchSubreddits searches for subreddits beginning with the query provided.
// List subreddits that begin with a query string.
// Subreddits whose names begin with query will be returned.
// If include_over_18 is false, subreddits with over-18 content restrictions will be filtered from the results.
// If include_unadvertisable is False, subreddits that have hide_ads set to True or are on the anti_ads_subreddits list will be filtered.
// If exact is true, only an exact match will be returned.
// Exact matches are inclusive of over_18 subreddits, but not hide_ad subreddits when include_unadvertisable is False.
func (s *SubredditService) SearchSubreddits(ctx context.Context, opts SubredditSearchOptions) ([]string, *http.Response, error) {
	path, err := addOptions("api/search_subreddits", opts)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(rootSubredditNames)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Names, resp, nil
}

// PostSiteAdmin Create or configure a subreddit.
// If sr is specified, the request will attempt to modify the specified subreddit.
// If not, a subreddit with name "name" will be created.
// This endpoint expects all values to be supplied on every request.
// If modifying a subset of options, it may be useful to get the current settings from /about/edit.json first.
// For backwards compatibility, description is the sidebar text and public_description is the publicly visible subreddit description.
// Most of the parameters for this endpoint are identical to options visible in the user interface and their meanings are best explained there.
func (s *SubredditService) PostSiteAdmin(ctx context.Context, modHash string, opts SubredditAdminOptions) (*http.Response, error) {
	path := "api/site_admin"
	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Set("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetSubmitText Get the submission text for the subreddit.
// This text is set by the subreddit moderators and intended to be displayed on the submission form.
func (s *SubredditService) GetSubmitText(ctx context.Context, subreddit string) (string, *http.Response, error) {
	path := fmt.Sprintf("r/%s/api/submit_text", subreddit)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", nil, &InternalError{Message: err.Error()}
	}

	data := new(struct {
		Text string `json:"submit_text"`
	})
	var resp *http.Response

	resp, err = s.client.Do(ctx, req, data)
	if err != nil {
		return "", nil, &ResponseError{Message: err.Error(), Response: resp}
	}

	return data.Text, resp, nil
}

type SubredditAutocompleteOptions struct {
	IncludeOver18   bool   `json:"include_over_18"`
	IncludeProfiles bool   `json:"include_profiles"`
	Query           string `json:"query"` // a string up to 25 characters long, consisting of printable characters.
}

// GetSubredditAutocomplete returns a list of subreddits and data for subreddits whose names start with 'query'.
// Uses typeahead endpoint to receive the list of subreddits names. Typeahead provides exact matches, typo correction, fuzzy matching and boosts subreddits to the top that the user is subscribed to.
func (s *SubredditService) GetSubredditAutocomplete(ctx context.Context, opts SubredditAutocompleteOptions) (*http.Response, error) {
	path := "api/subreddit_autocomplete"

	req, err := s.client.NewJSONRequest(http.MethodGet, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type SubredditAutocompleteV2Options struct {
	IncludeOver18   bool   `json:"include_over_18"`
	IncludeProfiles bool   `json:"include_profiles"`
	Limit           int    `json:"limit,omitempty"` // an integer between 1 and 10 (default: 5)
	Query           string `json:"query"`           // a string up to 25 characters long, consisting of printable characters.
	SearchQueryID   string `json:"search_query_id"` // a uuid
	TypeaheadActive bool   `json:"typeahead_active,omitempty"`
}

func (s *SubredditService) GetSubredditAutocompleteV2(ctx context.Context, opts SubredditAutocompleteV2Options) (*http.Response, error) {
	path := "api/subreddit_autocomplete_v2"

	req, err := s.client.NewJSONRequest(http.MethodGet, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type SubredditOPType string

const (
	SubredditOPSave    SubredditOPType = "save"
	SubredditOPPreview SubredditOPType = "preview"
)

type SubredditStylesheetOptions struct {
	APIType            string          `json:"api_type"` // Usually always "JSON"
	OP                 SubredditOPType `json:"op"`
	Reason             string          `json:"reason"`              // a string up to 256 characters long, consisting of printable characters.
	StylesheetContents string          `json:"stylesheet_contents"` // the new stylesheet content
}

// PostSubredditStylesheet updates a subreddit's stylesheet. op should be save to update the contents of the stylesheet.
func (s *SubredditService) PostSubredditStylesheet(ctx context.Context, subreddit, modHash string, opts SubredditStylesheetOptions) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/subreddit_stylesheet", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type SubredditSubscribeActionType string

const (
	SubredditSubscribeActionSub   SubredditSubscribeActionType = "sub"
	SubredditSubscribeActionUnsub SubredditSubscribeActionType = "unsub"
)

type SubredditSubscribeActionSourceType string

const (
	SubredditSubscribeActionSourceOnboarding    SubredditSubscribeActionSourceType = "onboarding"
	SubredditSubscribeActionSourceAutosubscribe SubredditSubscribeActionSourceType = "autosubscribe"
)

type SubredditSubscribeOptions struct {
	Action       SubredditSubscribeActionType       `json:"action"`
	ActionSource SubredditSubscribeActionSourceType `json:"action_source"`
	SRs          []string                           `json:"sr,omitempty"`      // A comma-separated list of subreddit fullnames (when using the "sr" parameter),
	SRNames      []string                           `json:"sr_name,omitempty"` // or of subreddit names (when using the "sr_name" parameter).
}

// PostSubscribe subscribes to or unsubscribe from a subreddit.
// To subscribe, action should be sub. To unsubscribe, action should be unsub. The user must have access to the subreddit to be able to subscribe to it.
// The skip_initial_defaults param can be set to True to prevent automatically subscribing the user to the current set of defaults when they take their first subscription action. Attempting to set it for an unsubscribe action will result in an error.
func (s *SubredditService) PostSubscribe(ctx context.Context, modHash string, opts SubredditSubscribeOptions) (*http.Response, error) {
	path := "api/subscribe"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type SubredditImageType string

const (
	SubredditImagePNG SubredditImageType = "png"
	SubredditImageJPG SubredditImageType = "jpg"
)

type SubredditImageUploadType string

const (
	SubredditImageUploadIMG    SubredditImageUploadType = "img"
	SubredditImageUploadHeader SubredditImageUploadType = "header"
	SubredditImageUploadIcon   SubredditImageUploadType = "icon"
	SubredditImageUploadBanner SubredditImageUploadType = "banner"
)

type SubredditUploadImageOptions struct {
	File       []byte                   `json:"file"` // file upload with maximum size of 500 KiB
	FormID     string                   `json:"formid,omitempty"`
	Header     int                      `json:"header"`             // either 1 or 0
	IMGType    SubredditImageType       `json:"img_type,omitempty"` // one of png or jpg (default: png)
	Name       string                   `json:"name,omitempty"`     // a valid subreddit image name
	UploadType SubredditImageUploadType `json:"upload_type"`
}

// PostUploadSubredditImage Add or replace a subreddit image, custom header logo, custom mobile icon, or custom mobile banner.
//
// If the upload_type value is img, an image for use in the subreddit stylesheet is uploaded with the name specified in name.
// If the upload_type value is header then the image uploaded will be the subreddit's new logo and name will be ignored.
// If the upload_type value is icon then the image uploaded will be the subreddit's new mobile icon and name will be ignored.
// If the upload_type value is banner then the image uploaded will be the subreddit's new mobile banner and name will be ignored.
// For backwards compatibility, if upload_type is not specified, the header field will be used instead:
//
// If the header field has value 0, then upload_type is img.
// If the header field has value 1, then upload_type is header.
// The img_type field specifies whether to store the uploaded image as a PNG or JPEG.
//
// Subreddits have a limited number of images that can be in use at any given time. If no image with the specified name already exists, one of the slots will be consumed.
//
// If an image with the specified name already exists, it will be replaced. This does not affect the stylesheet immediately, but will take effect the next time the stylesheet is saved.
func (s *SubredditService) PostUploadSubredditImage(ctx context.Context, subreddit, modHash string, opts SubredditUploadImageOptions) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/upload_sr_image", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetPostRequirements Fetch moderator-designated requirements to post to the subreddit.
//
// Moderators may enable certain restrictions, such as minimum title length, when making a submission to their subreddit.
//
// Clients may use the values returned by this endpoint to pre-validate fields before making a request to POST /api/submit. This may allow the client to provide a better user experience to the user, for example by creating a text field in their app that does not allow the user to enter more characters than the max title length.
//
// A non-exhaustive list of possible requirements a moderator may enable:
//
// body_blacklisted_strings: List of strings. Users may not submit posts that contain these words.
// body_restriction_policy: String. One of "required", "notAllowed", or "none", meaning that a self-post body is required, not allowed, or optional, respectively.
// domain_blacklist: List of strings. Users may not submit links to these domains
// domain_whitelist: List of strings. Users submissions MUST be from one of these domains
// is_flair_required: Boolean. If True, flair must be set at submission time.
// title_blacklisted_strings: List of strings. Submission titles may NOT contain any of the listed strings.
// title_required_strings: List of strings. Submission title MUST contain at least ONE of the listed strings.
// title_text_max_length: Integer. Maximum length of the title field.
// title_text_min_length: Integer. Minimum length of the title field.
func (s *SubredditService) GetPostRequirements(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("api/v1/%s/post_requirements", subreddit)
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	var resp *http.Response
	resp, err = s.client.Do(ctx, req, nil)
	if err != nil {
		return nil, &ResponseError{Message: err.Error(), Response: resp}
	}

	return resp, nil
}

// GetAbout Return information about the subreddit.
// Data includes the subscriber count, description, and header image.
func (s *SubredditService) GetAbout(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/about", subreddit)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	return s.client.Do(ctx, req, nil)
}

// GetAboutEdit Get the current settings of a subreddit.
// In the API, this returns the current settings of the subreddit as used by /api/site_admin.
// On the HTML site, it will display a form for editing the subreddit.
func (s *SubredditService) GetAboutEdit(ctx context.Context, subreddit string) (*SubredditAdminOptions, *http.Response, error) {
	path := fmt.Sprintf("r/%s/about/edit", subreddit)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, &InternalError{Message: err.Error()}
	}
	options := new(SubredditAdminOptions)
	var resp *http.Response
	resp, err = s.client.Do(ctx, req, options)
	if err != nil {
		return nil, nil, &ResponseError{
			Message:  err.Error(),
			Response: resp,
		}
	}
	return options, resp, nil
}

// GetAboutRules Get the rules for the current subreddit
func (s *SubredditService) GetAboutRules(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/about/rules", subreddit)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	return s.client.Do(ctx, req, nil)
}

// GetAboutTraffic Gets traffic
func (s *SubredditService) GetAboutTraffic(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/about/traffic", subreddit)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	return s.client.Do(ctx, req, nil)
}

// GetSidebar Get the sidebar for the current subreddit
func (s *SubredditService) GetSidebar(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/sidebar", subreddit)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	return s.client.Do(ctx, req, nil)
}

// GetSticky Redirect to one of the posts stickied in the current subreddit
// The "num" argument can be used to select a specific sticky, and will default to 1 (the top sticky) if not specified.
// Will 404 if there is not currently a sticky post in this subreddit.
func (s *SubredditService) GetSticky(ctx context.Context, subreddit string, num int) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/sticky", subreddit)

	if num != 1 && num != 2 {
		num = 1
	}

	req, err := s.client.NewRequest(http.MethodGet, path, []byte(fmt.Sprintf("{\n\tnum: %d\n}", num)))
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	return s.client.Do(ctx, req, nil)
}

type SubredditsMineWhere string

const (
	SubredditMineWhereSubscriber  SubredditsMineWhere = "subscriber"
	SubredditMineWhereContributor SubredditsMineWhere = "contributor"
	SubredditMineWhereModerator   SubredditsMineWhere = "moderator"
	SubredditMineWhereStreams     SubredditsMineWhere = "streams"
)

// GetMineWhere Get subreddits the user has a relationship with.
//
// The where parameter chooses which subreddits are returned as follows:
//
// subscriber - subreddits the user is subscribed to
// contributor - subreddits the user is an approved user in
// moderator - subreddits the user is a moderator of
// streams - subscribed to subreddits that contain hosted video links
func (s *SubredditService) GetMineWhere(ctx context.Context, where SubredditsMineWhere, opts *ListingOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("subreddits/mine/%s", where)

	return s.client.getListing(ctx, path, opts)
}

// GetSubredditsSearch search subreddits by title and description.
func (s *SubredditService) GetSubredditsSearch(ctx context.Context, opts *ListingSubredditOptions) (*Listing, *http.Response, error) {
	return s.client.getListing(ctx, "subreddits/search", opts)
}

type SubredditsWhere string

const (
	SubredditsWherePopular SubredditsWhere = "popular"
	SubredditsWhereNew     SubredditsWhere = "new"
	SubredditsWhereGold    SubredditsWhere = "gold"
	SubredditsWhereDefault SubredditsWhere = "default"
)

func (s *SubredditService) GetSubredditsWhere(ctx context.Context, where SubredditsMineWhere, opts *ListingOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("subreddits/%s", where)

	return s.client.getListing(ctx, path, opts)
}
