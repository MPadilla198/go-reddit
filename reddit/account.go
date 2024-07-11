package reddit

import (
	"context"
	"net/http"
	"net/url"
)

// AccountService handles communication with the account
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_account
type AccountService struct {
	client *Client
}

const (
	accountGetIdentityPath = "api/v1/me"
)

// SubredditKarma holds user karma data for the subreddit.
type SubredditKarma struct {
	Subreddit    string `json:"sr"`
	PostKarma    int    `json:"link_karma"`
	CommentKarma int    `json:"comment_karma"`
}

// Preferences are the user's account settings.
// Some of the fields' descriptions are taken from:
// https://praw.readthedocs.io/en/latest/code_overview/other/preferences.html#praw.models.Preferences.update
type Preferences struct {
	// Control whose private messages you see.
	// - "everyone": everyone except blocked users
	// - "whitelisted": only trusted users
	AcceptPrivateMessages          string `json:"accept_pms,omitempty"`
	ActivityRelevantAds            bool   `json:"activity_relevant_ads,omitempty"`              // Allow Reddit to use your activity on Reddit to show you more relevant advertisements.
	AllowClicktracking             bool   `json:"allow_clicktracking,omitempty"`                // Allow reddit to log my outbound clicks for personalization.
	BadCommentCollapse             string `json:"bad_comment_collapse,omitempty"`               // one of (`off`, `low`, `medium`, `high`),
	Beta                           bool   `json:"beta,omitempty"`                               // Beta test features for reddit. By enabling, you will join r/beta immediately.
	Clickgadget                    bool   `json:"clickgadget,omitempty"`                        // Show me links I've recently viewed.
	CollapseReadMessages           bool   `json:"collapse_read_messages,omitempty"`             //
	Compress                       bool   `json:"compress,omitempty"`                           // Compress the post display (make them look more compact).
	CountryCode                    string `json:"country_code,omitempty"`                       // one of (`WF`, `JP`, `JM`, `JO`, `WS`, `JE`, `GW`, `GU`, `GT`, `GS`, `GR`, `GQ`, `GP`, `GY`, `GG`, `GF`, `GE`, `GD`, `GB`, `GA`, `GN`, `GM`, `GL`, `GI`, `GH`, `PR`, `PS`, `PW`, `PT`, `PY`, `PA`, `PF`, `PG`, `PE`, `PK`, `PH`, `PN`, `PL`, `PM`, `ZM`, `ZA`, `ZZ`, `ZW`, `ME`, `MD`, `MG`, `MF`, `MA`, `MC`, `MM`, `ML`, `MO`, `MN`, `MH`, `MK`, `MU`, `MT`, `MW`, `MV`, `MQ`, `MP`, `MS`, `MR`, `MY`, `MX`, `MZ`, `FR`, `FI`, `FJ`, `FK`, `FM`, `FO`, `CK`, `CI`, `CH`, `CO`, `CN`, `CM`, `CL`, `CC`, `CA`, `CG`, `CF`, `CD`, `CZ`, `CY`, `CX`, `CR`, `CW`, `CV`, `CU`, `SZ`, `SY`, `SX`, `SS`, `SR`, `SV`, `ST`, `SK`, `SJ`, `SI`, `SH`, `SO`, `SN`, `SM`, `SL`, `SC`, `SB`, `SA`, `SG`, `SE`, `SD`, `YE`, `YT`, `LB`, `LC`, `LA`, `LK`, `LI`, `LV`, `LT`, `LU`, `LR`, `LS`, `LY`, `VA`, `VC`, `VE`, `VG`, `IQ`, `VI`, `IS`, `IR`, `IT`, `VN`, `IM`, `IL`, `IO`, `IN`, `IE`, `ID`, `BD`, `BE`, `BF`, `BG`, `BA`, `BB`, `BL`, `BM`, `BN`, `BO`, `BH`, `BI`, `BJ`, `BT`, `BV`, `BW`, `BQ`, `BR`, `BS`, `BY`, `BZ`, `RU`, `RW`, `RS`, `RE`, `RO`, `OM`, `HR`, `HT`, `HU`, `HK`, `HN`, `HM`, `EH`, `EE`, `EG`, `EC`, `ET`, `ES`, `ER`, `UY`, `UZ`, `US`, `UM`, `UG`, `UA`, `VU`, `NI`, `NL`, `NO`, `NA`, `NC`, `NE`, `NF`, `NG`, `NZ`, `NP`, `NR`, `NU`, `XK`, `XZ`, `XX`, `KG`, `KE`, `KI`, `KH`, `KN`, `KM`, `KR`, `KP`, `KW`, `KZ`, `KY`, `DO`, `DM`, `DJ`, `DK`, `DE`, `DZ`, `TZ`, `TV`, `TW`, `TT`, `TR`, `TN`, `TO`, `TL`, `TM`, `TJ`, `TK`, `TH`, `TF`, `TG`, `TD`, `TC`, `AE`, `AD`, `AG`, `AF`, `AI`, `AM`, `AL`, `AO`, `AN`, `AQ`, `AS`, `AR`, `AU`, `AT`, `AW`, `AX`, `AZ`, `QA`),
	DefaultCommentSort             string `json:"default_comment_sort,omitempty"`               // One of "confidence", "top", "new", "controversial", "old", "random", "qa", "live".
	DomainDetails                  bool   `json:"domain_details,omitempty"`                     // Show additional details in the domain text when available, such as the source subreddit or the content author’s url/name.
	EmailChatRequest               bool   `json:"email_chat_request,omitempty"`                 // Send chat requests as emails.
	EmailCommentReply              bool   `json:"email_comment_reply,omitempty"`                // Send comment replies as emails.
	EmailDigests                   bool   `json:"email_digests,omitempty"`                      // Send email digests.
	EmailMessages                  bool   `json:"email_messages,omitempty"`                     // Send messages as emails.
	EmailPostReply                 bool   `json:"email_post_reply,omitempty"`                   // Send post replies as emails.
	EmailPrivateMessages           bool   `json:"email_private_messages,omitempty"`             // Send private messages as emails.
	EmailUnsubscribeAll            bool   `json:"email_unsubscribe_all,omitempty"`              // Unsubscribe from all emails.
	EmailUpvoteComment             bool   `json:"email_upvote_comment,omitempty"`               // Send comment upvote updates as emails.
	EmailUpvotePost                bool   `json:"email_upvote_post,omitempty"`                  // Send post upvote updates as emails.
	EmailUserNewFollower           bool   `json:"email_user_new_follower,omitempty"`            // Send new follower alerts as emails.
	EmailUsernameMention           bool   `json:"email_username_mention,omitempty"`             // Send username mentions as emails.
	EnableDefaultThemes            bool   `json:"enable_default_themes,omitempty"`              // Use Reddit Theme
	EnableFollowers                bool   `json:"enable_followers,omitempty"`                   //
	EnableRedditProAnalyticsEmails bool   `json:"enable_reddit_pro_analytics_emails,omitempty"` //
	FeedRecommendationsEnabled     bool   `json:"feed_recommendations_enabled,omitempty"`       //  Enable feed recommendations.
	// One of "GLOBAL", "AR", "AU", "BG", "CA", "CL", "CO", "CZ", "FI", "GB", "GR", "HR", "HU",
	// "IE", "IN", "IS", "JP", "MX", "MY", "NZ", "PH", "PL", "PR", "PT", "RO", "RS", "SE", "SG",
	// "TH", "TR", "TW", "US", "US_AK", "US_AL", "US_AR", "US_AZ", "US_CA", "US_CO", "US_CT",
	// "US_DC", "US_DE", "US_FL", "US_GA", "US_HI", "US_IA", "US_ID", "US_IL", "US_IN", "US_KS",
	// "US_KY", "US_LA", "US_MA", "US_MD", "US_ME", "US_MI", "US_MN", "US_MO", "US_MS", "US_MT",
	// "US_NC", "US_ND", "US_NE", "US_NH", "US_NJ", "US_NM", "US_NV", "US_NY", "US_OH", "US_OK",
	// "US_OR", "US_PA", "US_RI", "US_SC", "US_SD", "US_TN", "US_TX", "US_UT", "US_VA", "US_VT",
	// "US_WA", "US_WI", "US_WV", "US_WY".
	G                      string `json:"g,omitempty"`
	HideAds                bool   `json:"hide_ads,omitempty"`
	HideDowns              bool   `json:"hide_downs,omitempty"`              // Don’t show me submissions after I’ve downvoted them, except my own.
	HideFromRobots         bool   `json:"hide_from_robots,omitempty"`        // Don't allow search engines to index my user profile.
	HideUps                bool   `json:"hide_ups,omitempty"`                // Don’t show me posts after I’ve upvoted them, except my own.
	HighlightControversial bool   `json:"highlight_controversial,omitempty"` // Show a dagger (†) on comments voted controversial (one that's been upvoted and downvoted significantly).
	HighlightNewComments   bool   `json:"highlight_new_comments,omitempty"`  // Highlight new comments.
	IgnoreSuggestedSort    bool   `json:"ignore_suggested_sort,omitempty"`   // Ignore suggested sorts for specific threads/subreddits, like Q&As.
	InRedesignBeta         bool   `json:"in_redesign_beta,omitempty"`        // Use new Reddit as my default experience.
	LabelNSFW              bool   `json:"label_nsfw,omitempty"`              // Label posts that are not safe for work (NSFW).
	Lang                   string `json:"lang,omitempty"`                    // A valid IETF language tag (underscore separated).
	LegacySearch           bool   `json:"legacy_search,omitempty"`           // Show legacy search page.
	LiveOrangereds         bool   `json:"live_orangereds,omitempty"`         // Send message notifications in my browser.
	MarkMessagesRead       bool   `json:"mark_messages_read,omitempty"`      // Mark messages as read when I open my inbox.
	// Determine whether to show thumbnails next to posts in subreddits.
	// - "on": show thumbnails next to posts
	// - "off": do not show thumbnails next to posts
	// - "subreddit": show thumbnails next to posts based on the subreddit's preferences
	Media string `json:"media,omitempty"`
	// Determine whether to auto-expand media in subreddits.
	// - "on": auto-expand media previews
	// - "off": do not auto-expand media previews
	// - "subreddit": auto-expand media previews based on the subreddit's preferences
	MediaPreview string `json:"media_preview,omitempty"`
	// Don't show me comments with a score less than this number.
	// Must be between -100 and 100 (inclusive).
	MinCommentScore int `json:"min_comment_score,omitempty"`
	// Don't show me posts with a score less than this number.
	// Must be between -100 and 100 (inclusive).
	MinLinkScore                          int    `json:"min_link_score,omitempty"`
	MonitorMentions                       bool   `json:"monitor_mentions,omitempty"`                    // Notify me when people say my username.
	NewWindow                             bool   `json:"newwindow,omitempty"`                           // Opens link in a new window/tab.
	DarkMode                              bool   `json:"nightmode,omitempty"`                           // Enable nightmode
	NoProfanity                           bool   `json:"no_profanity,omitempty"`                        // Don’t show thumbnails or media previews for anything labeled NSFW.
	NumComments                           int    `json:"num_comments,omitempty"`                        // Display this many comments by default. Must be between 1 and 500 (inclusive).
	NumSites                              int    `json:"numsites,omitempty"`                            // Number of links to display at once (between 1 and 100).
	Organic                               bool   `json:"organic,omitempty"`                             // Show the spotlight box on the home feed.
	OtherTheme                            string `json:"other_theme,omitempty"`                         // subreddit theme to use (subreddit name).
	Over18                                bool   `json:"over_18,omitempty"`                             // I am over eighteen years old and willing to view adult content.
	PrivateFeeds                          bool   `json:"private_feeds,omitempty"`                       // Enable private RSS feeds.
	ProfileOptOut                         bool   `json:"profile_opt_out,omitempty"`                     // View user profiles on desktop using legacy mode.
	PublicVotes                           bool   `json:"public_votes,omitempty"`                        // Make my upvotes and downvotes public.
	Research                              bool   `json:"research,omitempty"`                            // Allow my data to be used for research purposes.
	SearchIncludeOver18                   bool   `json:"search_include_over_18,omitempty"`              // Include not safe for work (NSFW) search results in searches.
	SendCrosspostMessages                 bool   `json:"send_crosspost_messages,omitempty"`             // Send crosspost messages.
	SendWelcomeMessages                   bool   `json:"send_welcome_messages,omitempty"`               // Send welcome messages.
	ShowFlair                             bool   `json:"show_flair,omitempty"`                          // Show a user's flair (next to their name on a post or comment).
	ShowGoldExpiration                    bool   `json:"show_gold_expiration,omitempty"`                // Show how much gold you have remaining on your profile.
	ShowLinkFlair                         bool   `json:"show_link_flair,omitempty"`                     // Show a post's flair.
	ShowLocationBasedRecommendations      bool   `json:"show_location_based_recommendations,omitempty"` // Show location based recommendations.
	ShowPresence                          bool   `json:"show_presence,omitempty"`
	ShowPromote                           bool   `json:"show_promote,omitempty"`
	ShowStylesheets                       bool   `json:"show_stylesheets,omitempty"`                           // Allow subreddits to show me custom themes.
	ShowTrending                          bool   `json:"show_trending,omitempty"`                              // Show trending subreddits on the home feed.
	ShowTwitter                           bool   `json:"show_twitter,omitempty"`                               // Show a link to your Twitter account on your profile.
	SMSNotificationEnabled                bool   `json:"sms_notification_enabled,omitempty"`                   //
	StoreVisits                           bool   `json:"store_visits,omitempty"`                               // Store visits.
	ThemeSelector                         string `json:"theme_selector,omitempty"`                             // Theme selector (subreddit name).
	ThirdPartyDataPersonalizedAds         bool   `json:"third_party_data_personalized_ads,omitempty"`          // Allow Reddit to use data provided by third-parties to show you more relevant advertisements on Reddit.
	ThirdPartyPersonalizedAds             bool   `json:"third_party_personalized_ads,omitempty"`               // Allow personalization of advertisements.
	ThirdPartySiteDataPersonalizedAds     bool   `json:"third_party_site_data_personalized_ads,omitempty"`     // Allow personalization of advertisements using data from third-party websites.
	ThirdPartySiteDataPersonalizedContent bool   `json:"third_party_site_data_personalized_content,omitempty"` // Allow personalization of content using data from third-party websites.
	ThreadedMessages                      bool   `json:"threaded_messages,omitempty"`                          // Show message conversations in the inbox.
	ThreadedModmail                       bool   `json:"threaded_modmail,omitempty"`                           // Enable threaded modmail display.
	TopKarmaSubreddits                    bool   `json:"top_karma_subreddits,omitempty"`                       // Top karma subreddits.
	UseGlobalDefaults                     bool   `json:"use_global_defaults,omitempty"`                        // Use global defaults.
	VideoAutoplay                         bool   `json:"video_autoplay,omitempty"`                             // Autoplay Reddit videos on the desktop comments page.
	WhatsappCommentReply                  bool   `json:"whatsapp_comment_reply,omitempty"`
	WhatsappEnabled                       bool   `json:"whatsapp_enabled,omitempty"`
}

// GetIdentity returns some general information about your account.
func (s *AccountService) GetIdentity(ctx context.Context) (*Account, *http.Response, error) {
	path := accountGetIdentityPath

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(Account)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, nil, &ResponseError{err.Error(), resp}
	}

	return root, resp, nil
}

// GetKarma returns a breakdown of your karma per subreddit.
func (s *AccountService) GetKarma(ctx context.Context) (resp *http.Response, err error) {
	path := "api/v1/me/karma"
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetPreferences returns your account settings.
func (s *AccountService) GetPreferences(ctx context.Context) (*Preferences, *http.Response, error) {
	path := "api/v1/me/prefs"

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(Preferences)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root, resp, nil
}

// UpdatePreferences updates your account settings and returns the modified version.
func (s *AccountService) UpdatePreferences(ctx context.Context, settings *Preferences) (*Preferences, *http.Response, error) {
	path := "api/v1/me/prefs"

	req, err := s.client.NewJSONRequest(http.MethodPatch, path, settings)
	if err != nil {
		return nil, nil, err
	}

	root := new(Preferences)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root, resp, nil
}

// GetTrophies returns a list of your trophies.
func (s *AccountService) GetTrophies(ctx context.Context) (resp *http.Response, err error) {
	path := "api/v1/me/trophies"
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// GetFriends returns a list of your friends.
func (s *AccountService) GetFriends(ctx context.Context, opts ListingOptions) (*Listing, *http.Response, error) {
	path := "prefs/friends"

	return s.client.getListing(ctx, path, opts)
}

// GetBlocked returns a list of your blocked users.
func (s *AccountService) GetBlocked(ctx context.Context, opts ListingOptions) (*Listing, *http.Response, error) {
	path := "prefs/blocked"

	return s.client.getListing(ctx, path, opts)
}

// GetMessaging returns blocked users and trusted users, respectively.
func (s *AccountService) GetMessaging(ctx context.Context, opts ListingOptions) (*Listing, *http.Response, error) {
	path := "prefs/messaging"

	return s.client.getListing(ctx, path, opts)
}

// GetTrusted returns a list of your trusted users.
func (s *AccountService) GetTrusted(ctx context.Context, opts ListingOptions) (*Listing, *http.Response, error) {
	path := "prefs/trusted"

	return s.client.getListing(ctx, path, opts)
}

// AddTrusted adds a user to your trusted users.
// This is not visible in the Reddit API docs.
func (s *AccountService) AddTrusted(ctx context.Context, username string) (*http.Response, error) {
	path := "api/add_whitelisted"

	form := url.Values{}
	form.Set("api_type", "json")
	form.Set("name", username)

	req, err := s.client.NewRequest(http.MethodPost, path, []byte(form.Encode()))
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// RemoveTrusted removes a user from your trusted users.
// This is not visible in the Reddit API docs.
func (s *AccountService) RemoveTrusted(ctx context.Context, username string) (*http.Response, error) {
	path := "api/remove_whitelisted"

	form := url.Values{}
	form.Set("name", username)

	req, err := s.client.NewRequest(http.MethodPost, path, []byte(form.Encode()))
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}
