package reddit

import (
	"encoding/json"
	"fmt"
)

const (
	kindComment   = "t1"
	kindAccount   = "t2"
	kindLink      = "t3"
	kindMessage   = "t4"
	kindSubreddit = "t5"
	kindAward     = "t6"

	kindListing           = "Listing"
	kindSubredditSettings = "subreddit_settings"
	kindKarmaList         = "KarmaList"
	kindTrophyList        = "TrophyList"
	kindUserList          = "UserList"
	kindMore              = "more"
	kindLiveThread        = "LiveUpdateEvent"
	kindLiveThreadUpdate  = "LiveUpdate"
	kindModAction         = "modaction"
	kindMulti             = "LabeledMulti"
	kindMultiDescription  = "LabeledMultiDescription"
	kindWikiPage          = "wikipage"
	kindWikiPageListing   = "wikipagelisting"
	kindWikiPageSettings  = "wikipagesettings"
	kindStyleSheet        = "stylesheet"
)

type Thing interface {
	json.Unmarshaler
	getID() string
	getName() string
}

// thing is an entity on Reddit.
// Its kind represents what it is and what is stored in the Data field.
// e.g. t1 = comment, t2 = user, t3 = post, etc.
type thing struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
	// Data interface{} `json:"data"`
}

func (t thing) getID() string {
	return t.ID
}

func (t thing) getName() string {
	return t.Name
}

// Listing is a list of things coming from the Reddit API.
// It also contains the after anchor useful to get the next results via subsequent requests.
type Listing struct {
	After    string  `json:"after"`
	Before   string  `json:"before"`
	ModHash  string  `json:"modhash"`
	Children []Thing `json:"children"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (l *Listing) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, l)
	if err != nil {
		return &JSONError{
			Message: fmt.Sprintf("error during unmarshal: %s", err.Error()),
			Data:    b}
	}
	return nil
}

type votable struct {
	Ups   int   `json:"ups"`
	Downs int   `json:"downs"`
	Likes *bool `json:"likes"`
}

type created struct {
	Created    int64 `json:"created"`
	CreatedUtc int64 `json:"created_utc"`
}

// Comment is a comment posted by a user.
type Comment struct {
	thing
	Data struct {
		votable
		created

		ApprovedBy            string     `json:"approved_by"`
		Author                string     `json:"author,omitempty"`
		AuthorFlairCSSClass   string     `json:"author_flair_css_class"`
		AuthorFlairText       string     `json:"author_flair_text"`
		BannedBy              string     `json:"banned_by"`
		Body                  string     `json:"body"`
		BodyHTML              string     `json:"body_html"`
		Distinguished         string     `json:"distinguished"`
		Edited                *Timestamp `json:"edited"`
		Gilded                int        `json:"gilded"`
		Likes                 *bool      `json:"likes"`
		LinkAuthor            string     `json:"link_author"`
		LinkID                string     `json:"link_id"`
		LinkTitle             string     `json:"link_title"`
		LinkURL               string     `json:"link_url"`
		NumReports            int        `json:"num_reports"`
		ParentID              string     `json:"parent_id"`
		Replies               []Thing    `json:"replies"`
		Saved                 bool       `json:"saved"`
		Score                 int        `json:"score"`
		SubredditName         string     `json:"subreddit"`
		SubredditNamePrefixed string     `json:"subreddit_name_prefixed"`
		SubredditID           string     `json:"subreddit_id"`
	} `json:"data"`
}

func (c *Comment) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, c)
	if err != nil {
		return &JSONError{
			Message: fmt.Sprintf("error during unmarshal: %s", err.Error()),
			Data:    b}
	}
	return nil
}

// Link is a submitted post on Reddit.
type Link struct {
	thing
	Data struct {
		votable
		created

		Author              string      `json:"author"`
		AuthorFlairCSSClass string      `json:"author_flair_css_class"`
		AuthorFlairText     string      `json:"author_flair_text"`
		Clicked             bool        `json:"clicked"`
		Distinguished       string      `json:"distinguished"`
		Domain              string      `json:"domain"`
		Hidden              bool        `json:"hidden"`
		IsSelf              bool        `json:"is_self"`
		Likes               bool        `json:"likes"`
		LinkFlairCSSClass   string      `json:"link_flair_css_class"`
		LinkFlairText       string      `json:"link_flair_text"`
		Locked              bool        `json:"locked"`
		Media               interface{} `json:"media"` // Object class
		MediaEmbed          interface{} `json:"mediaEmbed"`
		NumComments         int         `json:"num_comments"`
		Over18              bool        `json:"over18"`
		Permalink           string      `json:"permalink"`
		Saved               bool        `json:"saved"`
		Score               int         `json:"score"`
		Selftext            string      `json:"selftext"`
		SelftextHTML        string      `json:"selftext_html"`
		Stickied            bool        `json:"stickied"`
		Subreddit           string      `json:"subreddit"`
		SubredditID         string      `json:"subreddit_id"`
		Thumbnail           string      `json:"thumbnail"`
		Title               string      `json:"title"`
		URL                 string      `json:"url"`
		Edited              int64       `json:"edited"`
	} `json:"data"`
}

func (l *Link) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, l)
	if err != nil {
		return &JSONError{
			Message: fmt.Sprintf("error during unmarshal: %s", err.Error()),
			Data:    b}
	}
	return nil
}

// Subreddit holds information about a subreddit
type Subreddit struct {
	thing
	Data struct {
		AccountsActive       int    `json:"accounts_active"`
		CommentScoreHideMins int    `json:"comment_score_hide_mins"`
		Description          string `json:"description"`
		DescriptionHTML      string `json:"description_html"`
		DisplayName          string `json:"display_name"`
		HeaderImg            string `json:"header_img"`
		HeaderSize           []int  `json:"header_size"`
		HeaderTitle          string `json:"header_title"`
		Over18               bool   `json:"over18"`
		PublicDescription    string `json:"public_description"`
		PublicTraffic        bool   `json:"public_traffic"`
		Subscribers          int64  `json:"subscribers"`
		SubmissionType       string `json:"submission_type"`
		SubmitLinkLabel      string `json:"submit_link_label"`
		SubmitTextLabel      string `json:"submit_text_label"`
		SubredditType        string `json:"subreddit_type"`
		Title                string `json:"title"`
		URL                  string `json:"url"`
		UserIsBanned         bool   `json:"user_is_banned"`
		UserIsContributor    bool   `json:"user_is_contributor"`
		UserIsModerator      bool   `json:"user_is_moderator"`
		UserIsSubscriber     bool   `json:"user_is_subscriber"`
	} `json:"data"`
}

func (s *Subreddit) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, s)
	if err != nil {
		return &JSONError{
			Message: fmt.Sprintf("error during unmarshal: %s", err.Error()),
			Data:    b}
	}
	return nil
}

type Message struct {
	thing
	Data struct {
		created

		Author       string `json:"author"`
		Body         string `json:"body"`
		BodyHTML     string `json:"body_html"`
		Context      string `json:"context"`
		FirstMessage string `json:"first_message"`
		Likes        bool   `json:"likes"`
		LinkTitle    string `json:"link_title"`
		New          bool   `json:"new"`
		ParentID     string `json:"parent_id"`
		Replies      string `json:"replies"`
		Subject      string `json:"subject"`
		Subreddit    string `json:"subreddit"`
		WasComment   bool   `json:"was_comment"`
	} `json:"data"`
}

func (m *Message) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, m)
	if err != nil {
		return &JSONError{
			Message: fmt.Sprintf("error during unmarshal: %s", err.Error()),
			Data:    b}
	}
	return nil
}

type Account struct {
	thing
	Data struct {
		created

		CommentKarma     int    `json:"comment_karma"`
		HasMail          bool   `json:"has_mail"`
		HasModMail       bool   `json:"has_mod_mail"`
		HasVerifiedEmail bool   `json:"has_verified_email"`
		ID               string `json:"id"`
		InboxCount       int    `json:"inbox_count"`
		IsFriend         bool   `json:"is_friend"`
		IsGold           bool   `json:"is_gold"`
		IsMod            bool   `json:"is_mod"`
		LinkKarma        int    `json:"link_karma"`
		Modhash          string `json:"modhash"`
		Name             string `json:"name"`
		Over18           bool   `json:"over18"`
	} `json:"data"`
}

func (a *Account) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, a)
	if err != nil {
		return &JSONError{
			Message: fmt.Sprintf("error during unmarshal: %s", err.Error()),
			Data:    b}
	}
	return nil
}

type Award struct {
	thing
}

func (a *Award) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, a)
	if err != nil {
		return &JSONError{
			Message: fmt.Sprintf("error during unmarshal: %s", err.Error()),
			Data:    b}
	}
	return nil
}

// More holds information used to retrieve additional comments omitted from a base comment tree.
type More struct {
	thing
	Data struct {
		Children []string `json:"children"`
	} `json:"data"`
}

func (m *More) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, m)
	if err != nil {
		return &JSONError{
			Message: fmt.Sprintf("error during unmarshal: %s", err.Error()),
			Data:    b}
	}
	return nil
}
