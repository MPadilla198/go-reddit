package reddit

import (
	"context"
	"fmt"
	"net/http"
)

// UserService handles communication with the user
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_users
type UserService struct {
	client *Client
}

// GetSearch Search user profiles by title and description.
func (s *UserService) GetSearch(ctx context.Context, opts *ListingSubredditOptions) (*Listing, *http.Response, error) {
	return s.client.getListing(ctx, "users/search", opts)
}

type UsersWhere string

const (
	UsersWherePopular UsersWhere = "popular"
	UsersWhereNew     UsersWhere = "new"
)

// GetUsersWhere gets all user subreddits.
// The where parameter chooses the order in which the subreddits are displayed.
// "popular" sorts on the activity of the subreddit and the position of the subreddits can shift around.
// "new" sorts the user subreddits based on their creation date, newest first.
func (s *UserService) GetUsersWhere(ctx context.Context, where UsersWhere, opts *ListingOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("users/%s", where)

	return s.client.getListing(ctx, path, opts)
}

type UserBlockOptions struct {
	AccountID string `json:"account_id,omitempty"` // fullname of an account
	APIType   string `json:"api_type"`
	Name      string `json:"name,omitempty"` // A valid, existing reddit username
}

// PostBlockUser For blocking a user. Only accessible to approved OAuth applications
func (s *UserService) PostBlockUser(ctx context.Context, modHash string, opts UserBlockOptions) (*http.Response, error) {
	path := "api/block_user"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type UserRelationshipType string

const (
	UserRelationshipFriend          UserRelationshipType = "friend"
	UserRelationshipModerator       UserRelationshipType = "moderator"
	UserRelationshipModeratorInvite UserRelationshipType = "moderator_invite"
	UserRelationshipContributor     UserRelationshipType = "contributor"
	UserRelationshipBanned          UserRelationshipType = "banned"
	UserRelationshipMuted           UserRelationshipType = "muted"
	UserRelationshipWikibanned      UserRelationshipType = "wikibanned"
	UserRelationshipWikicontributor UserRelationshipType = "wikicontributor"
)

type UserFriendOptions struct {
	APIType     string               `json:"api_type"`
	BanContext  string               `json:"ban_context,omitempty"` // fullname of a thing
	BanMessage  string               `json:"ban_message,omitempty"`
	Container   string               `json:"container,omitempty"` //  If type is friend or enemy, 'container' MUST be the current user's fullname; for other types, the subreddit must be set via URL
	Duration    int                  `json:"duration,omitempty"`  // an integer between 1 and 999
	Name        string               `json:"name"`                // the name of an existing user`
	Note        string               `json:"note,omitempty"`      // A string of no longer than 300 characters
	Permissions string               `json:"permissions,omitempty"`
	Type        UserRelationshipType `json:"type"`
}

// PostFriend Create a relationship between a user and another user or subreddit
//
// OAuth2 use requires appropriate scope based on the 'type' of the relationship:
// moderator: Use "moderator_invite"
// moderator_invite: modothers
// contributor: modcontributors
// banned: modcontributors
// muted: modcontributors
// wikibanned: modcontributors and modwiki
// wikicontributor: modcontributors and modwiki
// friend: Use /api/v1/me/friends/{username}
// enemy: Use /api/block
// Complement to POST_unfriend
func (s *UserService) PostFriend(ctx context.Context, subreddit, modHash string, opts UserFriendOptions) (*http.Response, error) {
	path := fmt.Sprintf("api/friend")
	if subreddit != "" {
		path = fmt.Sprintf("r/%s/%s", subreddit, path)
	}

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type UserReportOptions struct {
	Details string `json:"details"` // JSON data
	Reason  string `json:"reason"`  // a string no longer than 100 characters
	User    string `json:"user"`    // A valid, existing reddit username
}

// PostReportUser Report a user. Reporting a user brings it to the attention of a Reddit admin.
func (s *UserService) PostReportUser(ctx context.Context, opts UserReportOptions) (*http.Response, error) {
	path := "api/report_user"

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type UserPermissionsOptions struct {
	APIType     string `json:"api_type"`
	Name        string `json:"name"` // the name of an existing user
	Permissions string `json:"permissions"`
	Type        string `json:"type"`
}

func (s *UserService) PostSetPermissions(ctx context.Context, subreddit, modHash string, opts UserPermissionsOptions) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/setpermissions", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

type UserUnfriendOptions struct {
	APIType   string               `json:"api_type"`
	Container string               `json:"container,omitempty"` // If type is friend or enemy, 'container' MUST be the current user's fullname
	ID        string               `json:"id,omitempty"`        // fullname of the thing
	Name      string               `json:"name,omitempty"`      // The name of the eisting user
	Type      UserRelationshipType `json:"type"`
}

// PostUnfriend Remove a relationship between a user and another user or subreddit
//
// The user can either be passed in by name (nuser) or by fullname (iuser). If type is friend or enemy, 'container' MUST be the current user's fullname; for other types, the subreddit must be set via URL (e.g., /r/funny/api/unfriend)
//
// OAuth2 use requires appropriate scope based on the 'type' of the relationship:
//
// moderator: modothers
// moderator_invite: modothers
// contributor: modcontributors
// banned: modcontributors
// muted: modcontributors
// wikibanned: modcontributors and modwiki
// wikicontributor: modcontributors and modwiki
// friend: Use /api/v1/me/friends/{username}
// enemy: privatemessages
// Complement to POST_friend
func (s *UserService) PostUnfriend(ctx context.Context, subreddit, modHash string, opts UserUnfriendOptions) (*http.Response, error) {
	path := fmt.Sprintf("api/unfriend")
	if subreddit != "" {
		path = fmt.Sprintf("r/%s/%s", subreddit, path)
	}

	req, err := s.client.NewJSONRequest(http.MethodPost, path, opts)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	req.Header.Add("X-Modhash", modHash)

	return s.client.Do(ctx, req, nil)
}

func (s *UserService) GetUserDataByAccountIDs(ctx context.Context, userList []string) (*http.Response, error) {
	path := "api/user_data_by_account_ids"

	req, err := s.client.NewJSONRequest(http.MethodGet, path, userList)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetUsernameAvailable checks whether a username is available for registration.
func (s *UserService) GetUsernameAvailable(ctx context.Context, username string) (*http.Response, error) {
	data := struct {
		User string `json:"user"`
	}{User: username}

	path := "api/username_available"

	req, err := s.client.NewJSONRequest(http.MethodGet, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// DeleteFriendByUsername unfriends a user. User is a valid, unused, username
func (s *UserService) DeleteFriendByUsername(ctx context.Context, username string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"`
	}{ID: username}

	path := fmt.Sprintf("api/v1/me/friends/%s", username)

	req, err := s.client.NewJSONRequest(http.MethodDelete, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	return s.client.Do(ctx, req, nil)
}

// GetFriendByUsername Get information about a specific 'friend', such as notes.
func (s *UserService) GetFriendByUsername(ctx context.Context, username string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"`
	}{ID: username}

	path := fmt.Sprintf("api/v1/me/friends/%s", username)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	return s.client.Do(ctx, req, nil)
}

// PutFriendByUsername Create or update a "friend" relationship.
// This operation is idempotent. It can be used to add a new friend, or update an existing friend (e.g., add/change the note on that friend)
// "note": a string no longer than 300 characters,
func (s *UserService) PutFriendByUsername(ctx context.Context, username, note string) (*http.Response, error) {
	data := struct {
		Name string `json:"name"`
		Note string `json:"note"`
	}{Name: username, Note: note}

	path := fmt.Sprintf("api/v1/me/friends/%s", username)

	req, err := s.client.NewJSONRequest(http.MethodPut, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}
	return s.client.Do(ctx, req, nil)
}

// GetUserTrophies Return a list of trophies for the a given user.
func (s *UserService) GetUserTrophies(ctx context.Context, username string) (*http.Response, error) {
	data := struct {
		ID string `json:"id"`
	}{ID: username}

	path := fmt.Sprintf("api/v1/user/%s/trophies", username)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, data)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetUserAbout Return information about the user, including karma and gold status.
func (s *UserService) GetUserAbout(ctx context.Context, username string) (*http.Response, error) {
	path := fmt.Sprintf("user/%s/about", username)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

type UserWhere string

const (
	UserWhereOverview  UserWhere = "overview"
	UserWhereSubmitted UserWhere = "submitted"
	UserWhereComments  UserWhere = "comments"
	UserWhereUpvoted   UserWhere = "upvoted"
	UserWhereDownvoted UserWhere = "downvoted"
	UserWhereHidden    UserWhere = "hidden"
	UserWhereSaved     UserWhere = "saved"
	UserWhereGilded    UserWhere = "gilded"
)

func (s *UserService) GetUserWhere(ctx context.Context, username string, where UserWhere, opts ListingUserOptions) (*Listing, *http.Response, error) {
	path := fmt.Sprintf("user/%s/%s", username, where)

	return s.client.getListing(ctx, path, opts)
}
