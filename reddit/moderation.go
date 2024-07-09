package reddit

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ModerationService handles communication with the moderation
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_moderation
type ModerationService struct {
	client *Client
}

// GetSubredditAboutLog Get a list of recent moderation actions.
// Moderator actions taken within a subreddit are logged.
// This listing is a view of that log with various filters to aid in analyzing the information.
// The optional mod parameter can be a comma-delimited list of moderator names to restrict the results to, or the string to restrict the results to admin actions taken within the subreddit.
// The type parameter is optional and if sent limits the log entries returned to only those of the type specified.
// This endpoint is a listing.
func (s *ModerationService) GetSubredditAboutLog(ctx context.Context, subreddit string, opts *ListingModerationOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("r/%s/about/log", subreddit)

	return s.client.getListing(ctx, path, opts)
}

type ModeratorLocationType string

const (
	ModeratorLocationReports     ModeratorLocationType = "reports"
	ModeratorLocationSpam        ModeratorLocationType = "spam"
	ModeratorLocationModqueue    ModeratorLocationType = "modqueue"
	ModeratorLocationUnmoderated ModeratorLocationType = "unmoderated"
	ModeratorLocationEdited      ModeratorLocationType = "edited"
)

// GetSubredditAboutLocation Return a listing of posts relevant to moderators.
// reports: Things that have been reported.
// spam: Things that have been marked as spam or otherwise removed.
// modqueue: Things requiring moderator review, such as reported things and items caught by the spam filter.
// unmoderated: Things that have yet to be approved/removed by a mod.
// edited: Things that have been edited recently.
// Requires the "posts" moderator permission for the subreddit.
// This endpoint is a listing.
func (s *ModerationService) GetSubredditAboutLocation(ctx context.Context, subreddit string, location ModeratorLocationType, opts *ListingModerationOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("r/%s/about/%s", subreddit, location)

	return s.client.getListing(ctx, path, opts)
}

// PostSubredditAcceptModeratorInvite Accept an invitation to moderate the specified subreddit.
// The authenticated user must have been invited to moderate the subreddit by one of its current moderators.
// See also: /api/friend and /subreddits/mine.
func (s *ModerationService) PostSubredditAcceptModeratorInvite(ctx context.Context, modHash, subreddit string) (*http.Response, error) {
	data := struct {
		APIType string `json:"api_type"`
	}{APIType: "json"}

	path := fmt.Sprintf("r/%s/api/accept_moderator_invite", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostApprove Approve a link or comment.
// If the thing was removed, it will be re-inserted into appropriate listings.
// Any reports on the approved thing will be discarded.
// See also: /api/remove.
func (s *ModerationService) PostApprove(ctx context.Context, modHash, fullname string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"`
	}{ID: fullname}

	path := "api/approve"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type ModeratorDistinguishHowType string

const (
	ModeratorDistinguishHowYes     ModeratorDistinguishHowType = "yes"
	ModeratorDistinguishHowNo      ModeratorDistinguishHowType = "no"
	ModeratorDistinguishHowAdmin   ModeratorDistinguishHowType = "admin"
	ModeratorDistinguishHowSpecial ModeratorDistinguishHowType = "special"
)

type ModeratorDistinguishOptions struct {
	APIType string                      `json:"api_type"`
	How     ModeratorDistinguishHowType `json:"how"`
	ID      string                      `json:"id"` // fullname of a thing
	Sticky  bool                        `json:"sticky"`
}

// PostDistinguish Distinguish a thing's author with a sigil.
// This can be useful to draw attention to and confirm the identity of the user in the context of a link or comment of theirs.
// The options for distinguish are as follows:
// yes - add a moderator distinguish ([M]). only if the user is a moderator of the subreddit the thing is in.
// no - remove any distinguishes.
// admin - add an admin distinguish ([A]). admin accounts only.
// special - add a user-specific distinguish. depends on user.
// The first time a top-level comment is moderator distinguished, the author of the link the comment is in reply to will get a notification in their inbox.
// sticky is a boolean flag for comments, which will stick the distinguished comment to the top of all comments threads.
// If a comment is marked sticky, it will override any other stickied comment for that link (as only one comment may be stickied at a time.) Only top-level comments may be stickied.
func (s *ModerationService) PostDistinguish(ctx context.Context, modHash string, opts *ModeratorDistinguishOptions) (*http.Response, error) {
	path := "api/distinguish"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostIgnoreReports Prevent future reports on a thing from causing notifications.
// Any reports made about a thing after this flag is set on it will not cause notifications or make the thing show up in the various moderation listings.
// See also: /api/unignore_reports.
func (s *ModerationService) PostIgnoreReports(ctx context.Context, modHash, fullname string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a thing
	}{ID: fullname}

	path := "api/ignore_reports"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostLeaveContributor Abdicate approved user status in a subreddit.
// See also: /api/friend.
func (s *ModerationService) PostLeaveContributor(ctx context.Context, modHash, fullname string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a thing
	}{ID: fullname}

	path := "api/leavecontributor"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostLeaveModerator Abdicate moderator status in a subreddit.
// See also: /api/friend.
func (s *ModerationService) PostLeaveModerator(ctx context.Context, modHash, fullname string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a thing
	}{ID: fullname}

	path := "api/leavemoderator"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostMuteMessageAuthor For muting user via modmail.
func (s *ModerationService) PostMuteMessageAuthor(ctx context.Context, modHash, fullname string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a thing
	}{ID: fullname}

	path := "api/mute_message_author"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostRemove Remove a link, comment, or modmail message.
// If the thing is a link, it will be removed from all subreddit listings.
// If the thing is a comment, it will be redacted and removed from all subreddit comment listings.
// See also: /api/approve.
func (s *ModerationService) PostRemove(ctx context.Context, modHash, fullname string, spam bool) (*http.Response, error) {
	data := struct {
		ID   string `json:"id"` // fullname of a thing
		Spam bool   `json:"spam"`
	}{ID: fullname, Spam: spam}

	path := "api/remove"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostShowComment Mark a comment that it should not be collapsed because of crowd control.
// The comment could still be collapsed for other reasons.
func (s *ModerationService) PostShowComment(ctx context.Context, modHash, fullname string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a thing
	}{ID: fullname}

	path := "api/show_comment"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostSnoozeReports Prevent future reports on a thing from causing notifications.
// For users who reported this thing (post, comment, etc.) with the given report reason, reports from those users in the next 7 days will not be escalated to moderators.
// See also: /api/unsnooze_reports.
func (s *ModerationService) PostSnoozeReports(ctx context.Context, modHash, fullname, reason string) (*http.Response, error) {
	data := struct {
		ID     string `json:"id"` // fullname of a thing
		Reason string `json:"reason"`
	}{ID: fullname, Reason: reason}

	path := "api/snooze_reports"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostUnignoreReports Allow future reports on a thing to cause notifications.
// See also: /api/ignore_reports.
func (s *ModerationService) PostUnignoreReports(ctx context.Context, modHash, fullname string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a thing
	}{ID: fullname}

	path := "api/unignore_reports"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostUnmuteMessageAuthor For unmuting user via modmail.
func (s *ModerationService) PostUnmuteMessageAuthor(ctx context.Context, modHash, fullname string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"` // fullname of a thing
	}{ID: fullname}

	path := "api/unmute_message_author"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostUnsnoozeReports For users whose reports were snoozed (see /api/snooze_reports), to go back to escalating future reports from those users.
func (s *ModerationService) PostUnsnoozeReports(ctx context.Context, modHash, fullname, reason string) (*http.Response, error) {
	data := struct {
		ID     string `json:"id"` // fullname of a thing
		Reason string `json:"reason"`
	}{ID: fullname, Reason: reason}

	path := "api/unsnooze_reports"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type ModeratorCrowdControlLevel int

const (
	ModeratorCrowdControlLevelNone ModeratorCrowdControlLevel = iota
	ModeratorCrowdControlLevelLow
	ModeratorCrowdControlLevelMedium
	ModeratorCrowdControlLevelHigh
)

// PostUpdateCrowdControlLevel Change the post's crowd control level.
func (s *ModerationService) PostUpdateCrowdControlLevel(ctx context.Context, modHash, fullname string, level ModeratorCrowdControlLevel) (*http.Response, error) {
	data := struct {
		ID    string                     `json:"id"` // fullname of a thing
		Level ModeratorCrowdControlLevel `json:"level"`
	}{ID: fullname, Level: level}

	path := "api/update_crowd_control_level"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

// PostSubredditStylesheet Redirect to the subreddit's stylesheet if one exists.
// See also: /api/subreddit_stylesheet.
func (s *ModerationService) PostSubredditStylesheet(ctx context.Context, subreddit string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/stylesheet", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

/**********************************************************
 *********************** MODMAIL **************************
 **********************************************************/

type ModmailStateType string

const (
	ModmailStateAll           ModmailStateType = "all"
	ModmailStateAppeals       ModmailStateType = "appeals"
	ModmailStateNotifications ModmailStateType = "notifications"
	ModmailStateInbox         ModmailStateType = "inbox"
	ModmailStateFiltered      ModmailStateType = "filtered"
	ModmailStateInProgress    ModmailStateType = "inprogress"
	ModmailStateMod           ModmailStateType = "mod"
	ModmailStateArchived      ModmailStateType = "archived"
	ModmailStateDefault       ModmailStateType = "default"
	ModmailStateHighlighted   ModmailStateType = "highlighted"
	ModmailStateJoinRequests  ModmailStateType = "join_requests"
	ModmailStateNew           ModmailStateType = "new"
)

// PostModmailBulkRead Marks all conversations read for a particular conversation state within the passed list of subreddits.
func (s *ModerationService) PostModmailBulkRead(ctx context.Context, state ModmailStateType, entity ...string) (*http.Response, error) {
	data := struct {
		Entity []string         `json:"entity"`
		State  ModmailStateType `json:"state"`
	}{Entity: entity, State: state}

	path := "api/mod/bulk_read"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type ModmailSortType string

const (
	ModmailSortRecent ModmailSortType = "recent"
	ModmailSortMod    ModmailSortType = "mod"
	ModmailSortUser   ModmailSortType = "user"
	ModmailSortUnread ModmailSortType = "unread"
)

type ModmailGetConversationOptions struct {
	After  string           `json:"after"`  // A Modmail Conversation ID, in the form ModmailConversation_<id>
	Entity []string         `json:"entity"` // comma-delimited list of subreddit names
	Limit  int              `json:"limit"`  // an integer between 1 and 100 (default: 25)
	Sort   ModmailSortType  `json:"sort"`
	State  ModmailStateType `json:"state"`
}

// GetModmailConversations Get conversations for a logged-in user or subreddits
func (s *ModerationService) GetModmailConversations(ctx context.Context, opts *ModmailGetConversationOptions) (*http.Response, error) {
	path := "api/mod/conversations"

	req, err := s.client.NewJSONRequest(http.MethodGet, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type ModmailPostConversationOptions struct {
	Body           string `json:"body"` // raw Markdown text
	IsAuthorHidden bool   `json:"isAuthorHidden"`
	SubredditName  string `json:"srName"`
	Subject        string `json:"subject"` // a string no longer than 100 characters
	To             string `json:"to"`      // modmail conversation recipient fullname
}

// PostModmailConversations Creates a new conversation for a particular SR.
// This endpoint will create a ModmailConversation object as well as the first ModmailMessage within the ModmailConversation object.
// A note on to:
// The to field for this endpoint is somewhat confusing. It can be:
// A User, passed like "username" or "u/username"
// A Subreddit, passed like "r/subreddit"
// null, meaning an internal moderator discussion
// In this way to is a bit of a misnomer in modmail conversations.
// What it really means is the participant of the conversation who is not a mod of the subreddit.
func (s *ModerationService) PostModmailConversations(ctx context.Context, opts *ModmailPostConversationOptions) (*http.Response, error) {
	path := "api/mod/conversations"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetModmailConversationsByID Returns all messages, mod actions and conversation metadata for a given conversation id
func (s *ModerationService) GetModmailConversationsByID(ctx context.Context, conversationID string, markRead bool) (*http.Response, error) {
	data := struct {
		ConversationID string `json:"conversation_id"` // A Modmail Conversation ID, in the form ModmailConversation_<id>
		MarkRead       bool   `json:"markRead"`
	}{ConversationID: conversationID, MarkRead: markRead}

	path := fmt.Sprintf("api/mod/conversations/%s", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type ModmailPostConversationByIDOptions struct {
	Body           string `json:"body"`            // raw Markdown text
	ConversationID string `json:"conversation_id"` // A Modmail Conversation ID, in the form ModmailConversation_<id>
	IsAuthorHidden bool   `json:"isAuthorHidden"`
	IsInternal     bool   `json:"isInternal"`
}

// PostModmailConversation Creates a new message for a particular conversation.
func (s *ModerationService) PostModmailConversation(ctx context.Context, opts *ModmailPostConversationByIDOptions) (*http.Response, error) {
	path := fmt.Sprintf("api/mod/conversations/%s", opts.ConversationID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostModmailConversationApproveByID Approve the non-mod user associated with a particular conversation.
func (s *ModerationService) PostModmailConversationApproveByID(ctx context.Context, conversationID string) (*http.Response, error) {
	data := struct {
		ConversationID string `json:"conversation_id"` // base36 modmail conversation id
	}{ConversationID: conversationID}

	path := fmt.Sprintf("api/mod/conversations/%s/approve", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostModmailConversationArchiveByID Marks a conversation as archived.
func (s *ModerationService) PostModmailConversationArchiveByID(ctx context.Context, conversationID string) (*http.Response, error) {
	data := struct {
		ConversationID string `json:"conversation_id"` // A Modmail Conversation ID, in the form ModmailConversation_<id>
	}{ConversationID: conversationID}

	path := fmt.Sprintf("api/mod/conversations/%s/archive", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostModmailConversationDisapproveByID Disapprove the non-mod user associated with a particular conversation.
func (s *ModerationService) PostModmailConversationDisapproveByID(ctx context.Context, conversationID string) (*http.Response, error) {
	data := struct {
		ConversationID string `json:"conversation_id"` // base36 modmail conversation id
	}{ConversationID: conversationID}

	path := fmt.Sprintf("api/mod/conversations/%s/disapprove", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// DeleteModmailConversationHighlightByID Removes a highlight from a conversation.
func (s *ModerationService) DeleteModmailConversationHighlightByID(ctx context.Context, conversationID string) (*http.Response, error) {
	data := struct {
		ConversationID string `json:"conversation_id"` // A Modmail Conversation ID, in the form ModmailConversation_<id>
	}{ConversationID: conversationID}

	path := fmt.Sprintf("api/mod/conversations/%s/highlight", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodDelete, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostModmailConversationHighlightByID Marks a conversation as highlighted.
func (s *ModerationService) PostModmailConversationHighlightByID(ctx context.Context, conversationID string) (*http.Response, error) {
	data := struct {
		ConversationID string `json:"conversation_id"` // A Modmail Conversation ID, in the form ModmailConversation_<id>
	}{ConversationID: conversationID}

	path := fmt.Sprintf("api/mod/conversations/%s/highlight", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type ModmailMuteHourType int

const (
	ModmailMute72Hours  ModmailMuteHourType = 72
	ModmailMute168Hours ModmailMuteHourType = 168
	ModmailMute672Hours ModmailMuteHourType = 672
)

// PostModmailConversationMuteByID Mutes the non-mod user associated with a particular conversation.
func (s *ModerationService) PostModmailConversationMuteByID(ctx context.Context, conversationID string, numHours ModmailMuteHourType) (*http.Response, error) {
	data := struct {
		ConversationID string              `json:"conversation_id"` // base36 modmail conversation id
		NumHours       ModmailMuteHourType `json:"num_hours"`
	}{ConversationID: conversationID, NumHours: numHours}

	path := fmt.Sprintf("api/mod/conversations/%s/mute", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostModmailConversationTempBanByID Temporary ban (switch from permanent to temporary ban) the non-mod user associated with a particular conversation.
func (s *ModerationService) PostModmailConversationTempBanByID(ctx context.Context, conversationID string, duration int) (*http.Response, error) {
	data := struct {
		ConversationID string `json:"conversation_id"` // base36 modmail conversation id
		Duration       int    `json:"duration"`        // an integer between 1 and 999
	}{ConversationID: conversationID, Duration: duration}

	path := fmt.Sprintf("api/mod/conversations/%s/temp_ban", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostModmailConversationUnarchiveByID Marks conversation as unarchived.
func (s *ModerationService) PostModmailConversationUnarchiveByID(ctx context.Context, conversationID string) (*http.Response, error) {
	data := struct {
		ConversationID string `json:"conversation_id"` // A Modmail Conversation ID, in the form ModmailConversation_<id>
	}{ConversationID: conversationID}

	path := fmt.Sprintf("api/mod/conversations/%s/unarchive", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostModmailConversationUnbanByID Unban the non-mod user associated with a particular conversation.
func (s *ModerationService) PostModmailConversationUnbanByID(ctx context.Context, conversationID string) (*http.Response, error) {
	data := struct {
		ConversationID string `json:"conversation_id"` // base36 modmail conversation id
	}{ConversationID: conversationID}

	path := fmt.Sprintf("api/mod/conversations/%s/unban", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostModmailConversationUnmuteByID Unmutes the non-mod user associated with a particular conversation.
func (s *ModerationService) PostModmailConversationUnmuteByID(ctx context.Context, conversationID string) (*http.Response, error) {
	data := struct {
		ConversationID string `json:"conversation_id"` // base36 modmail conversation id
	}{ConversationID: conversationID}

	path := fmt.Sprintf("api/mod/conversations/%s/unmute", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetModmailConversationUserByID Returns recent posts, comments and modmail conversations for a given user.
func (s *ModerationService) GetModmailConversationUserByID(ctx context.Context, conversationID string) (*http.Response, error) {
	data := struct {
		ConversationID string `json:"conversation_id"` // base36 modmail conversation id
	}{ConversationID: conversationID}

	path := fmt.Sprintf("api/mod/conversations/%s/user", conversationID)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostModmailConversationReadByIDs Marks a list of conversations as read for the user.
func (s *ModerationService) PostModmailConversationReadByIDs(ctx context.Context, conversationIDs ...string) (*http.Response, error) {
	data := struct {
		ConversationIDs []string `json:"conversationIds"` // A comma-separated list of items
	}{ConversationIDs: conversationIDs}

	path := "api/mod/conversations/read"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetModmailConversationSubreddits Returns a list of srs that the user moderates with mail permission
func (s *ModerationService) GetModmailConversationSubreddits(ctx context.Context) (*http.Response, error) {
	path := "api/mod/conversations/subreddits"

	req, err := s.client.NewJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostModmailConversationUnreadByIDs Marks conversations as unread for the user.
func (s *ModerationService) PostModmailConversationUnreadByIDs(ctx context.Context, conversationIDs ...string) (*http.Response, error) {
	data := struct {
		ConversationIDs []string `json:"conversationIds"` // A comma-separated list of items
	}{ConversationIDs: conversationIDs}

	path := "api/mod/conversations/unread"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetModmailConversationUnreadCount Endpoint to retrieve the unread conversation count by conversation state.
func (s *ModerationService) GetModmailConversationUnreadCount(ctx context.Context) (*http.Response, error) {
	path := "api/mod/conversations/unread/count"

	req, err := s.client.NewJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

/**********************************************************
 *********************** MODNOTE **************************
 **********************************************************/

type ModNoteDeleteOptions struct {
	NoteID    string // a unique ID for the note to be deleted (should have a ModNote_ prefix)
	Subreddit string // subreddit name
	User      string // account username
}

func (opts *ModNoteDeleteOptions) Params() url.Values {
	result := url.Values{}
	result.Add("note_id", opts.NoteID)
	result.Add("subreddit", opts.Subreddit)
	result.Add("user", opts.User)

	result.Add("type", "NOTE")

	return result
}

// DeleteModNotes Delete a mod user note where type=NOTE.
// Parameters should be passed as query parameters.
func (s *ModerationService) DeleteModNotes(ctx context.Context, opts *ModNoteDeleteOptions) (*http.Response, error) {
	params := opts.Params()

	path := "api/mod/notes?" + params.Encode()

	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type ModNoteFilterType string

const (
	ModNoteFilterNone          ModNoteFilterType = ""
	ModNoteFilterNote          ModNoteFilterType = "NOTE"
	ModNoteFilterApproval      ModNoteFilterType = "APPROVAL"
	ModNoteFilterRemoval       ModNoteFilterType = "REMOVAL"
	ModNoteFilterBan           ModNoteFilterType = "BAN"
	ModNoteFilterMute          ModNoteFilterType = "MUTE"
	ModNoteFilterInvite        ModNoteFilterType = "INVITE"
	ModNoteFilterSpam          ModNoteFilterType = "SPAM"
	ModNoteFilterContentChange ModNoteFilterType = "CONTENT_CHANGE"
	ModNoteFilterModAction     ModNoteFilterType = "MOD_ACTION"
	ModNoteFilterAll           ModNoteFilterType = "ALL"
)

type ModNoteGetOptions struct {
	Before    string            // (optional) an encoded string used for pagination with mod notes
	Filter    ModNoteFilterType // (optional) to be used for querying specific types of mod notes (default: all)
	Limit     int               // (optional) the number of mod notes to return in the response payload (default: 25, max: 100)'}
	Subreddit string            // subreddit name
	User      string            // account username
}

func (opts *ModNoteGetOptions) Params() url.Values {
	result := url.Values{}
	if opts.Before != "" {
		result.Add("before", opts.Before)
	}
	if opts.Filter != ModNoteFilterNone {
		result.Add("filter", string(opts.Filter))
	}
	if opts.Limit <= 100 && opts.Limit > 0 {
		result.Add("limit", strconv.Itoa(opts.Limit))
	}
	result.Add("subreddit", opts.Subreddit)
	result.Add("user", opts.User)

	return result
}

// GetModNotes Get mod notes for a specific user in a given subreddit.
func (s *ModerationService) GetModNotes(ctx context.Context, opts *ModNoteGetOptions) (*http.Response, error) {
	params := opts.Params()

	path := "api/mod/notes?" + params.Encode()

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type ModNotePostType string

const (
	ModNotePostNone             ModNotePostType = ""
	ModNotePostBotBan           ModNotePostType = "BOT_BAN"
	ModNotePostPermanentBan     ModNotePostType = "PERMA_BAN"
	ModNotePostBan              ModNotePostType = "BAN"
	ModNotePostAbuseWarning     ModNotePostType = "ABUSE_WARNING"
	ModNotePostSpamWarning      ModNotePostType = "SPAM_WARNING"
	ModNotePostSpamWatch        ModNotePostType = "SPAM_WATCH"
	ModNotePostSolidContributor ModNotePostType = "SOLID_CONTRIBUTOR"
	ModNotePostHelpfulUser      ModNotePostType = "HELPFUL_USER"
)

type ModNotePostOptions struct {
	Label     ModNotePostType // (optional)
	Note      string          // Content of the note, should be a string with a maximum character limit of 250
	RedditID  string          // (optional) a fullname of a comment or post (should have either a t1 or t3 prefix)
	Subreddit string          // subreddit name
	User      string          // account username
}

func (opts *ModNotePostOptions) Params() url.Values {
	result := url.Values{}
	if opts.Label != "" {
		result.Add("label", string(opts.Label))
	}
	result.Add("note", opts.Note)
	if opts.RedditID != "" {
		result.Add("reddit_id", opts.RedditID)
	}
	result.Add("subreddit", opts.Subreddit)
	result.Add("user", opts.User)

	result.Add("type", "NOTE")

	return result
}

// PostModNotes Create a mod user note where type=NOTE.
func (s *ModerationService) PostModNotes(ctx context.Context, opts *ModNotePostOptions) (*http.Response, error) {
	params := opts.Params()

	path := "api/mod/notes?" + params.Encode()

	req, err := s.client.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type ModNoteGetRecentOptions struct {
	Subreddits []string // a list of subreddits by name
	Users      []string // a list of usernames
}

func (opts *ModNoteGetRecentOptions) Params() url.Values {
	result := url.Values{}

	result.Set("subreddits", strings.Join(opts.Subreddits, ","))
	result.Set("users", strings.Join(opts.Users, ","))

	return result
}

// GetModNotesRecent Fetch the most recent notes written by a moderator
// Both parameters should be comma separated lists of equal lengths.
// The first subreddit will be paired with the first account to represent a query for a mod written note for that account in that subreddit and so forth for all subsequent pairs of subreddits and accounts.
// This request accepts up to 500 pairs of subreddit names and usernames.
// Parameters should be passed as query parameters.
// The response will be a list of mod notes in the order that subreddits and accounts were given.
// If no note exist for a given subreddit/account pair, then null will take its place in the list.
func (s *ModerationService) GetModNotesRecent(ctx context.Context, opts *ModNoteGetRecentOptions) (*http.Response, error) {
	params := opts.Params()

	path := "api/mod/notes/recent?" + params.Encode()

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}
