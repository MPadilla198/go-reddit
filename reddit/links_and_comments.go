package reddit

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// LinkAndCommentService handles communication with the comment
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_links_and_comments
type LinkAndCommentService struct {
	client *Client
}

type PostCommentOptions struct {
	APIType        string `json:"api_type"`
	RecaptchaToken string `json:"recaptcha_token"`
	ReturnRtjson   bool   `json:"return_rtjson"`
	RichtextJSON   string `json:"richtext_json"`
	Text           string `json:"text"`
	ThingID        string `json:"thing_id"`
}

// PostComment Submit a new comment or reply to a message.
// parent is the fullname of the thing being replied to.
// Its value changes the kind of object created by this request:
// the fullname of a Link: a top-level comment in that Link's thread. (requires submit scope)
// the fullname of a Comment: a comment reply to that comment. (requires submit scope)
// the fullname of a Message: a message reply to that message. (requires privatemessages scope)
// text should be the raw markdown body of the comment or message.
//
// To start a new message thread, use /api/compose.
func (s *LinkAndCommentService) PostComment(ctx context.Context, modHash string, opts *PostCommentOptions) (*http.Response, error) {
	path := "api/comment"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostDelete Delete a Link or Comment.
func (s *LinkAndCommentService) PostDelete(ctx context.Context, modHash, id string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` //fullname if a thing created by the user
	}{ID: id}

	path := "api/del"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type LinkEditUserTextOptions struct {
	APIType      string `json:"api_type"`
	ReturnRtjson bool   `json:"return_rtjson"`
	RichtextJSON string `json:"richtext_json"`
	Text         string `json:"text"`
	ThingID      string `json:"thing_id"`
}

// PostEditUserText Edit the body text of a comment or self-post.
func (s *LinkAndCommentService) PostEditUserText(ctx context.Context, modHash string, opts *LinkEditUserTextOptions) (*http.Response, error) {
	path := "api/editusertext"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostFollowLink Follow or unfollow a post.
// To follow, follow should be True.
// To unfollow, follow should be False.
// The user must have access to the subreddit to be able to follow a post within it.
func (s *LinkAndCommentService) PostFollowLink(ctx context.Context, modHash, fullname string, follow bool) (*http.Response, error) {
	data := struct {
		Follow   bool   `json:"follow"`
		Fullname string `json:"fullname"`
	}{Follow: follow, Fullname: fullname}

	path := "api/follow_post"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostHide Hide a link.
// This removes it from the user's default view of subreddit listings.
func (s *LinkAndCommentService) PostHide(ctx context.Context, modHash string, ids ...string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // A comma-separated list of link fullnames
	}{ID: strings.Join(ids, ",")}

	path := "api/hide"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type LinkSubredditInfoOptions struct {
	ID     []string `json:"id,omitempty"`      // A comma-separated list of thing fullnames
	SRName []string `json:"sr_name,omitempty"` // comma-delimited list of subreddit names
	URL    string   `json:"url"`               // a valid URL
}

// GetSubredditInfo Return a listing of things specified by their fullnames.
// Only Links, Comments, and Subreddits are allowed.
func (s *LinkAndCommentService) GetSubredditInfo(ctx context.Context, subreddit string, opts *LinkSubredditInfoOptions) (*http.Response, error) {
	path := "api/info"
	if subreddit != "" {
		path = fmt.Sprintf("r/%s/", subreddit) + path
	}

	req, err := s.client.NewJSONRequest(http.MethodGet, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostLock Lock a link or comment.
// Prevents a post or new child comments from receiving new comments.
// See also: /api/unlock.
func (s *LinkAndCommentService) PostLock(ctx context.Context, modHash, id string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"`
	}{ID: id}

	path := "api/lock"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-ModHash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostMarkNSFW Mark a link NSFW.
// See also: /api/unmarknsfw.
func (s *LinkAndCommentService) PostMarkNSFW(ctx context.Context, modHash, id string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"`
	}{ID: id}

	path := "api/marknsfw"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-ModHash", modHash)

	return s.client.Do(ctx, req, nil)
}

type LinkMoreChildrenOptions struct {
	Children      []string                          `json:"children"`
	Depth         int                               `json:"depth,omitempty"`
	ID            string                            `json:"id,omitempty"` // (optional) `:"optional"` id of the associated MoreChildren object
	LimitChildren bool                              `json:"limit_children"`
	LinkID        string                            `json:"link_id"` // fullname of a link
	Sort          SubredditSuggestedCommentSortType `json:"sort"`
}

// GetMoreChildren Retrieve additional comments omitted from a base comment tree.
// When a comment tree is rendered, the most relevant comments are selected for display first.
// Remaining comments are stubbed out with "MoreComments" links.
// This API call is used to retrieve the additional comments represented by those stubs, up to 100 at a time.
// The two core parameters required are link and children.
// link is the fullname of the link of the fetched comments.
// children is a comma-delimited list of comment ID36s that need to be fetched.
// If id is passed, it should be the ID of the MoreComments object this call is replacing.
// This is needed only for the HTML UI's purposes and is optional otherwise.
// NOTE: you may only make one request at a time to this API endpoint.
// Higher concurrency will result in an error being returned.
// If limit_children is True, only return the children requested.
// depth is the maximum depth of subtrees in the thread.
func (s *LinkAndCommentService) GetMoreChildren(ctx context.Context, opts *LinkMoreChildrenOptions) (*http.Response, error) {
	path := "api/morechildren"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type LinkReportOptions struct {
	AdditionalInfo string   `json:"additional_info"` // a string no longer than 2000 characters
	APIType        string   `json:"api_type"`
	CustomText     string   `json:"custom_text"` // a string no longer than 2000 characters
	FromHelpDesk   bool     `json:"from_help_desk"`
	FromModmail    bool     `json:"from_modmail"`
	ModmailConvID  string   `json:"modmail_conv_id"` // base36 modmail conversation id
	OtherReason    string   `json:"other_reason"`    // a string no longer than 100 characters
	Reason         string   `json:"reason"`          // a string no longer than 100 characters
	RuleReason     string   `json:"rule_reason"`     // a string no longer than 100 characters
	SiteReason     string   `json:"site_reason"`     // a string no longer than 100 characters
	SRName         string   `json:"sr_name"`         // a string no longer than 100 characters
	ThingID        string   `json:"thing_id"`        // fullname of a thing
	Usernames      []string `json:"usernames"`
}

// PostLinkReport Report a link, comment or message.
// Reporting a thing brings it to the attention of the subreddit's moderators.
// Reporting a message sends it to a system for admin review.
// For links and comments, the thing is implicitly hidden as well (see /api/hide for details).
// See /r/{subreddit}/about/rules for more about subreddit rules, and /r/{subreddit}/about for more about free_form_reports.
func (s *LinkAndCommentService) PostLinkReport(ctx context.Context, opts *LinkReportOptions) (*http.Response, error) {
	path := "api/report"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

func (s *LinkAndCommentService) PostLinkReportAward(ctx context.Context, awardID, reason string) (*http.Response, error) {
	data := struct {
		AwardID string `json:"award_id"`
		Reason  string `json:"reason"` // a string no longer than 100 characters
	}{AwardID: awardID, Reason: reason}

	path := "api/report_award"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostLinkSave Save a link or comment.
// Saved things are kept in the user's saved listing for later perusal.
// See also: /api/unsave.
func (s *LinkAndCommentService) PostLinkSave(ctx context.Context, modHash, id, category string) (*http.Response, error) {
	data := struct {
		Category string `json:"category"`
		ID       string `json:"id"` // fullname of a thing
	}{Category: category, ID: id}

	path := "api/save"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// GetSavedCategories Get a list of categories in which things are currently saved.
// See also: /api/save.
func (s *LinkAndCommentService) GetSavedCategories(ctx context.Context) (*http.Response, error) {
	path := "api/saved_categories"

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostSendReplies Enable or disable inbox replies for a link or comment.
// state is a boolean that indicates whether you are enabling or disabling inbox replies - true to enable, false to disable.
func (s *LinkAndCommentService) PostSendReplies(ctx context.Context, modHash, id string, state bool) (*http.Response, error) {
	data := struct {
		ID    string `json:"id"` // fullname of a thing created by the user
		State bool   `json:"state"`
	}{ID: id, State: state}

	path := "api/sendreplies"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostSetContestMode Set or unset "contest mode" for a link's comments.
// state is a boolean that indicates whether you are enabling or disabling contest mode - true to enable, false to disable.
func (s *LinkAndCommentService) PostSetContestMode(ctx context.Context, modHash, id string, state bool) (*http.Response, error) {
	data := struct {
		APIType string `json:"api_type"`
		ID      string `json:"id"` // fullname of a thing created by the user
		State   bool   `json:"state"`
	}{APIType: "json", ID: id, State: state}

	path := "api/set_contest_mode"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type LinkSubredditStickyOptions struct {
	APIType   string `json:"api_type"`
	ID        string `json:"id"`
	Num       int    `json:"num"` // integer between 1 and 4
	State     bool   `json:"state"`
	ToProfile bool   `json:"to_profile"`
}

// PostSetSubredditSticky Set or unset a Link as the sticky in its subreddit.
// state is a boolean that indicates whether to sticky or unsticky this post - true to sticky, false to unsticky.
// The num argument is optional, and only used when stickying a post.
// It allows specifying a particular "slot" to sticky the post into, and if there is already a post stickied in that slot it will be replaced.
// If there is no post in the specified slot to replace, or num is None, the bottom-most slot will be used.
func (s *LinkAndCommentService) PostSetSubredditSticky(ctx context.Context, modHash string, opts *LinkSubredditStickyOptions) (*http.Response, error) {
	path := "api/set_subreddit_sticky"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostSetSuggestedSort Set a suggested sort for a link.
// Suggested sorts are useful to display comments in a certain preferred way for posts.
// For example, casual conversation may be better sorted by new by default, or AMAs may be sorted by Q&A.
// A "sort" consisting of an empty string clears the default sort.
func (s *LinkAndCommentService) PostSetSuggestedSort(ctx context.Context, modHash, id string, sort SubredditSuggestedCommentSortType) (*http.Response, error) {
	data := struct {
		APIType string                            `json:"api_type"`
		ID      string                            `json:"id"` // fullname of a thing created by the user
		Sort    SubredditSuggestedCommentSortType `json:"sort"`
	}{APIType: "json", ID: id, Sort: sort}

	path := "api/set_suggested_sort"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostLinkSpoiler Set link spoiler.
func (s *LinkAndCommentService) PostLinkSpoiler(ctx context.Context, modHash, id string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a link
	}{ID: id}

	path := "api/spoiler"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostStoreVisits Requires a subscription to reddit premium
func (s *LinkAndCommentService) PostStoreVisits(ctx context.Context, modHash, id string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a link
	}{ID: id}

	path := "api/store_visits"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type LinkKindType string

const (
	LinkKindLink     LinkKindType = "link"
	LinkKindSelf     LinkKindType = "self"
	LinkKindImage    LinkKindType = "image"
	LinkKindVideo    LinkKindType = "video"
	LinkKindVideoGIF LinkKindType = "videogif"
)

type LinkSubmitOptions struct {
	Ad                   bool         `json:"ad"`
	APIType              string       `json:"api_type"`
	App                  string       `json:"app"`
	CollectionID         string       `json:"collection_id"` // (beta) the UUID of a collection
	Extension            string       `json:"extension"`     // extension used for redirects
	FlairID              string       `json:"flairID"`       // a string no longer than 36 characters
	FlairText            string       `json:"flair_text"`    // a string no longer than 36 characters
	GRecaptchaResponse   string       `json:"g-recaptcha-response"`
	Kind                 LinkKindType `json:"kind"`
	NSFW                 bool         `json:"nsfw"`
	PostSetDefaultPostID string       `json:"post_set_default_post_id"`
	PostSetID            string       `json:"post_set_id"`
	RecaptchaToken       string       `json:"recaptcha_token"`
	Resubmit             bool         `json:"resubmit"`
	RichtextJSON         string       `json:"richtext_json"`
	SendReplies          bool         `json:"send_replies"`
	Spoiler              bool         `json:"spoiler"`
	SR                   string       `json:"sr"`               // subreddit name
	Text                 string       `json:"text"`             // raw Markdown text
	Title                string       `json:"title"`            // title of the submission, up to 300 characters long
	URL                  string       `json:"url"`              // a valid url
	VideoPosterURL       string       `json:"video_poster_url"` // a valid url
}

// PostLinkSubmit Submit a link to a subreddit.
// Submit will create a link or self-post in the subreddit sr with the title.
// If kind is "link", then url is expected to be a valid URL to link to.
// Otherwise, text, if present, will be the body of the self-post unless richtext_json is present, in which case it will be converted into the body of the self-post.
// An error is thrown if both text and richtext_json are present.
// extension is used for determining which view-type (e.g. json, compact etc.) to use for the redirect that is generated after submit.
func (s *LinkAndCommentService) PostLinkSubmit(ctx context.Context, modHash string, opts *LinkSubmitOptions) (*http.Response, error) {
	path := "api/submit"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostLinkUnhide Unhide a link.
// See also: /api/hide.
func (s *LinkAndCommentService) PostLinkUnhide(ctx context.Context, modHash string, id ...string) (*http.Response, error) {
	data := struct {
		ID []string `json:"id"` // A comma-separated list of link fullnames
	}{ID: id}

	path := "api/unhide"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostLinkUnlock Unlock a link or comment.
// Allow a post or comment to receive new comments.
// See also: /api/lock.
func (s *LinkAndCommentService) PostLinkUnlock(ctx context.Context, modHash, id string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a thing
	}{ID: id}

	path := "api/unlock"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostLinkUnmarkNSFW Remove the NSFW marking from a link.
// See also: /api/marknsfw.
func (s *LinkAndCommentService) PostLinkUnmarkNSFW(ctx context.Context, modHash, id string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a thing
	}{ID: id}

	path := "api/unmarknsfw"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostLinkUnsave Unsave a link or comment.
// This removes the thing from the user's saved listings as well.
// See also: /api/save.
func (s *LinkAndCommentService) PostLinkUnsave(ctx context.Context, modHash, id string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a thing
	}{ID: id}

	path := "api/unsave"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostLinkUnspoiler Remove spoiler from thing.
func (s *LinkAndCommentService) PostLinkUnspoiler(ctx context.Context, modHash, id string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a thing
	}{ID: id}

	path := "api/unspoiler"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type LinkVoteDirection int

const (
	LinkDownVote = iota - 1
	LinkUnVote
	LinkUpVote
)

type LinkVoteOptions struct {
	Dir  LinkVoteDirection `json:"dir"`
	ID   string            `json:"id"`   // fullname of a thing
	Rank int               `json:"rank"` // an integer greater than 1
}

// PostLinkVote Cast a vote on a thing.
// id should be the fullname of the Link or Comment to vote on.
// dir indicates the direction of the vote.
// Voting 1 is an upvote, -1 is a downvote, and 0 is equivalent to "un-voting" by clicking again on a highlighted arrow.
// Note: votes must be cast by humans.
// That is, API clients proxying a human's action one-for-one are OK, but bots deciding how to vote on content or amplifying a human's vote are not.
// See the reddit rules for more details on what constitutes vote cheating.
func (s *LinkAndCommentService) PostLinkVote(ctx context.Context, modHash string, opts *LinkVoteOptions) (*http.Response, error) {
	path := "api/unspoiler"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}
