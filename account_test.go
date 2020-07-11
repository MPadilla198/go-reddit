package reddit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var expectedInfo = &User{
	ID:               "164ab8",
	Name:             "v_95",
	Created:          &Timestamp{time.Date(2017, 3, 12, 4, 56, 47, 0, time.UTC)},
	PostKarma:        488,
	CommentKarma:     22223,
	HasVerifiedEmail: true,
	NSFW:             true,
}

var expectedKarma = []SubredditKarma{
	{Subreddit: "nba", PostKarma: 144, CommentKarma: 21999},
	{Subreddit: "redditdev", PostKarma: 19, CommentKarma: 4},
	{Subreddit: "test", PostKarma: 1, CommentKarma: 0},
	{Subreddit: "golang", PostKarma: 1, CommentKarma: 0},
}

var expectedSettings = &Settings{
	AcceptPrivateMessages:                    String("everyone"),
	ActivityRelevantAds:                      Bool(false),
	AllowClickTracking:                       Bool(false),
	Beta:                                     Bool(false),
	ShowRecentlyViewedPosts:                  Bool(true),
	CollapseReadMessages:                     Bool(false),
	Compress:                                 Bool(false),
	CredditAutorenew:                         nil,
	DefaultCommentSort:                       String("top"),
	ShowDomainDetails:                        Bool(false),
	SendEmailDigests:                         Bool(false),
	SendMessagesAsEmails:                     Bool(false),
	UnsubscribeFromAllEmails:                 Bool(true),
	DisableCustomThemes:                      Bool(false),
	Location:                                 String("GLOBAL"),
	HideAds:                                  Bool(false),
	HideFromSearchEngines:                    Bool(false),
	HideUpvotedPosts:                         Bool(false),
	HideDownvotedPosts:                       Bool(false),
	HighlightControversialComments:           Bool(true),
	HighlightNewComments:                     Bool(true),
	IgnoreSuggestedSorts:                     Bool(true),
	UseNewReddit:                             nil,
	UsesNewReddit:                            Bool(false),
	LabelNSFW:                                Bool(true),
	Language:                                 String("en-ca"),
	ShowOldSearchPage:                        Bool(false),
	EnableNotifications:                      Bool(true),
	MarkMessagesAsRead:                       Bool(true),
	ShowThumbnails:                           String("subreddit"),
	AutoExpandMedia:                          String("off"),
	MinimumCommentScore:                      nil,
	MinimumPostScore:                         nil,
	EnableMentionNotifications:               Bool(true),
	OpenLinksInNewWindow:                     Bool(true),
	DarkMode:                                 Bool(true),
	DisableProfanity:                         Bool(false),
	NumberOfComments:                         Int(200),
	NumberOfPosts:                            Int(25),
	ShowSpotlightBox:                         nil,
	SubredditTheme:                           nil,
	ShowNSFW:                                 Bool(true),
	EnablePrivateRSSFeeds:                    Bool(true),
	ProfileOptOut:                            Bool(false),
	PublicizeVotes:                           Bool(false),
	AllowResearch:                            Bool(false),
	IncludeNSFWSearchResults:                 Bool(true),
	ReceiveCrosspostMessages:                 Bool(false),
	ReceiveWelcomeMessages:                   Bool(true),
	ShowUserFlair:                            Bool(true),
	ShowPostFlair:                            Bool(true),
	ShowGoldExpiration:                       Bool(false),
	ShowLocationBasedRecommendations:         Bool(false),
	ShowPromote:                              nil,
	ShowCustomSubredditThemes:                Bool(true),
	ShowTrendingSubreddits:                   Bool(true),
	ShowTwitter:                              Bool(false),
	StoreVisits:                              Bool(false),
	ThemeSelector:                            nil,
	AllowThirdPartyDataAdPersonalization:     Bool(false),
	AllowThirdPartySiteDataAdPersonalization: Bool(false),
	AllowThirdPartySiteDataContentPersonalization: Bool(false),
	EnableThreadedMessages:                        Bool(true),
	EnableThreadedModmail:                         Bool(false),
	TopKarmaSubreddits:                            Bool(false),
	UseGlobalDefaults:                             Bool(false),
	EnableVideoAutoplay:                           Bool(true),
}

var expectedRelationships = []Relationship{
	{
		ID:      "r9_1r4879",
		User:    "test1",
		UserID:  "t2_test1",
		Created: &Timestamp{time.Date(2020, 6, 28, 16, 43, 55, 0, time.UTC)},
	},
	{
		ID:      "r9_1re930",
		User:    "test2",
		UserID:  "t2_test2",
		Created: &Timestamp{time.Date(2020, 6, 28, 16, 44, 2, 0, time.UTC)},
	},
}

var expectedRelationships2 = []Relationship{
	{
		ID:      "r9_1re60i",
		User:    "test3",
		UserID:  "t2_test3",
		Created: &Timestamp{time.Date(2020, 3, 6, 2, 27, 0, 0, time.UTC)},
	},
}

func TestAccountService_Info(t *testing.T) {
	setup()
	defer teardown()

	blob := readFileContents(t, "testdata/account/info.json")

	mux.HandleFunc("/api/v1/me", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, blob)
	})

	info, _, err := client.Account.Info(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedInfo, info)
}

func TestAccountService_Karma(t *testing.T) {
	setup()
	defer teardown()

	blob := readFileContents(t, "testdata/account/karma.json")

	mux.HandleFunc("/api/v1/me/karma", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, blob)
	})

	karma, _, err := client.Account.Karma(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedKarma, karma)
}

func TestAccountService_Settings(t *testing.T) {
	setup()
	defer teardown()

	blob := readFileContents(t, "testdata/account/settings.json")

	mux.HandleFunc("/api/v1/me/prefs", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, blob)
	})

	settings, _, err := client.Account.Settings(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedSettings, settings)
}

func TestAccountService_UpdateSettings(t *testing.T) {
	setup()
	defer teardown()

	blob := readFileContents(t, "testdata/account/settings.json")
	expectedSettingsBody := &Settings{NumberOfPosts: Int(10), MinimumCommentScore: Int(5), Compress: Bool(true)}

	mux.HandleFunc("/api/v1/me/prefs", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)

		settingsBody := new(Settings)
		err := json.NewDecoder(r.Body).Decode(settingsBody)
		assert.NoError(t, err)
		assert.Equal(t, expectedSettingsBody, settingsBody)

		fmt.Fprint(w, blob)
	})

	settings, _, err := client.Account.UpdateSettings(ctx, expectedSettingsBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedSettings, settings)
}

func TestAccountService_Trophies(t *testing.T) {
	setup()
	defer teardown()

	blob := readFileContents(t, "testdata/account/trophies.json")

	mux.HandleFunc("/api/v1/me/trophies", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, blob)
	})

	trophies, _, err := client.Account.Trophies(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedTrophies, trophies)
}

func TestAccountService_Friends(t *testing.T) {
	setup()
	defer teardown()

	blob := readFileContents(t, "testdata/account/friends.json")

	mux.HandleFunc("/prefs/friends", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, blob)
	})

	relationships, _, err := client.Account.Friends(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedRelationships, relationships)
}

func TestAccountService_Blocked(t *testing.T) {
	setup()
	defer teardown()

	blob := readFileContents(t, "testdata/account/blocked.json")

	mux.HandleFunc("/prefs/blocked", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, blob)
	})

	relationships, _, err := client.Account.Blocked(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedRelationships, relationships)
}

func TestAccountService_Messaging(t *testing.T) {
	setup()
	defer teardown()

	blob := readFileContents(t, "testdata/account/messaging.json")

	mux.HandleFunc("/prefs/messaging", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, blob)
	})

	blocked, trusted, _, err := client.Account.Messaging(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedRelationships, blocked)
	assert.Equal(t, expectedRelationships2, trusted)
}

func TestAccountService_Trusted(t *testing.T) {
	setup()
	defer teardown()

	blob := readFileContents(t, "testdata/account/trusted.json")

	mux.HandleFunc("/prefs/trusted", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		fmt.Fprint(w, blob)
	})

	relationships, _, err := client.Account.Trusted(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedRelationships, relationships)
}

func TestAccountService_AddTrusted(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/add_whitelisted", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		form := url.Values{}
		form.Set("api_type", "json")
		form.Set("name", "test123")

		err := r.ParseForm()
		assert.NoError(t, err)
		assert.Equal(t, form, r.Form)
	})

	_, err := client.Account.AddTrusted(ctx, "test123")
	assert.NoError(t, err)
}

func TestAccountService_RemoveTrusted(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/remove_whitelisted", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		form := url.Values{}
		form.Set("name", "test123")

		err := r.ParseForm()
		assert.NoError(t, err)
		assert.Equal(t, form, r.Form)
	})

	_, err := client.Account.RemoveTrusted(ctx, "test123")
	assert.NoError(t, err)
}
