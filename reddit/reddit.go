package reddit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-querystring/query"
	"golang.org/x/oauth2"
)

const (
	libraryName    = "github.com/MPadilla198/go-reddit"
	libraryVersion = "2.1.0"

	defaultBaseURL         = "https://oauth.reddit.com"
	defaultBaseURLReadonly = "https://reddit.com"
	defaultTokenURL        = "https://www.reddit.com/api/v1/access_token"

	mediaTypeJSON = "application/json"
	mediaTypeForm = "application/x-www-form-urlencoded"

	headerContentType = "Content-Type"
	headerAccept      = "Accept"
	headerUserAgent   = "User-Agent"

	headerRateLimitRemaining = "x-ratelimit-remaining"
	headerRateLimitUsed      = "x-ratelimit-used"
	headerRateLimitReset     = "x-ratelimit-reset"
)

var defaultClient, _ = NewReadonlyClient()

// DefaultClient returns a valid, read-only client with limited access to the Reddit API.
func DefaultClient() *Client {
	return defaultClient
}

// RequestCompletionCallback defines the type of the request callback function.
type RequestCompletionCallback func(*http.Request, *http.Response)

// Credentials are used to authenticate to make requests to the Reddit API.
type Credentials struct {
	ID       string
	Secret   string
	Username string
	Password string
}

// Client manages communication with the Reddit API.
type Client struct {
	// HTTP client used to communicate with the Reddit API.
	client *http.Client

	BaseURL  *url.URL
	TokenURL *url.URL

	userAgent string

	rateMu sync.Mutex
	rate   Rate

	Credentials

	// This is the client's user ID in Reddit's database.
	redditID string

	Account        *AccountService
	Captcha        *CaptchaService
	Collection     *CollectionService
	Emoji          *EmojiService
	Flair          *FlairService
	Gold           *GoldService
	LinkAndComment *LinkAndCommentService
	Listings       *ListingsService
	LiveThread     *LiveThreadService
	Message        *MessageService
	Moderation     *ModerationService
	Multi          *MultiService
	Stream         *StreamService
	Subreddit      *SubredditService
	User           *UserService
	Widget         *WidgetService
	Wiki           *WikiService

	oauth2Transport *oauth2.Transport

	onRequestCompleted RequestCompletionCallback
}

// OnRequestCompleted sets the client's request completion callback.
func (c *Client) OnRequestCompleted(rc RequestCompletionCallback) {
	c.onRequestCompleted = rc
}

func newClient() *Client {
	baseURL, _ := url.Parse(defaultBaseURL)
	tokenURL, _ := url.Parse(defaultTokenURL)

	client := &Client{client: &http.Client{}, BaseURL: baseURL, TokenURL: tokenURL}

	client.Account = &AccountService{client: client}
	client.Captcha = &CaptchaService{client: client}
	client.Collection = &CollectionService{client: client}
	client.Emoji = &EmojiService{client: client}
	client.Flair = &FlairService{client: client}
	client.Gold = &GoldService{client: client}
	client.LinkAndComment = &LinkAndCommentService{client: client}
	client.Listings = &ListingsService{client: client}
	client.LiveThread = &LiveThreadService{client: client}
	client.Message = &MessageService{client: client}
	client.Moderation = &ModerationService{client: client}
	client.Multi = &MultiService{client: client}
	client.Stream = &StreamService{client: client}
	client.Subreddit = &SubredditService{client: client}
	client.User = &UserService{client: client}
	client.Widget = &WidgetService{client: client}
	client.Wiki = &WikiService{client: client}

	return client
}

// NewClient returns a new Reddit API client.
// Use an Opt to configure the client credentials, such as WithHTTPClient or WithUserAgent.
// If the FromEnv option is used with the correct environment variables, an empty struct can
// be passed in as the credentials, since they will be overridden.
func NewClient(credentials Credentials, opts ...Opt) (*Client, error) {
	client := newClient()

	client.ID = credentials.ID
	client.Secret = credentials.Secret
	client.Username = credentials.Username
	client.Password = credentials.Password

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, &InternalError{Message: err.Error()}
		}
	}

	userAgentTransport := &userAgentTransport{
		userAgent: client.UserAgent(),
		Base:      client.client.Transport,
	}
	client.client.Transport = userAgentTransport

	if client.client.CheckRedirect == nil {
		// todo
	}

	oauthTransport := oauthTransport(client)
	client.client.Transport = oauthTransport

	return client, nil
}

// NewReadonlyClient returns a new read-only Reddit API client.
// The client will have limited access to the Reddit API.
// Options that modify credentials (such as FromEnv) won't have any effect on this client.
func NewReadonlyClient(opts ...Opt) (*Client, error) {
	client := newClient()
	client.BaseURL, _ = url.Parse(defaultBaseURLReadonly)

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, &InternalError{Message: err.Error()}
		}
	}

	if client.client == nil {
		client.client = &http.Client{}
	}

	userAgentTransport := &userAgentTransport{
		userAgent: client.UserAgent(),
		Base:      client.client.Transport,
	}
	client.client.Transport = userAgentTransport

	return client, nil
}

// The readonly Reddit url needs .json at the end of its path to return responses in JSON instead of HTML.
func (c *Client) appendJSONExtensionToRequestURLPath(req *http.Request) {
	readonlyURL, err := url.Parse(defaultBaseURLReadonly)
	if err != nil {
		return
	}

	if req.URL.Host != readonlyURL.Host {
		return
	}

	req.URL.Path += ".json"
}

// UserAgent returns the client's user agent.
func (c *Client) UserAgent() string {
	if c.userAgent == "" {
		userAgent := fmt.Sprintf("golang:%s:v%s", libraryName, libraryVersion)
		if c.Username != "" {
			userAgent += fmt.Sprintf(" (by /u/%s)", c.Username)
		}
		c.userAgent = userAgent
	}
	return c.userAgent
}

// NewRequest creates an API request with form data as the body.
// The path is the relative URL which will be resolved to the BaseURL of the Client.
// It should always be specified without a preceding slash.
func (c *Client) NewRequest(method string, path string, form []byte) (*http.Request, error) {
	u, err := c.BaseURL.Parse(path)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	var body io.Reader
	if form != nil {
		body = bytes.NewReader(form)
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	c.appendJSONExtensionToRequestURLPath(req)
	req.Header.Add(headerContentType, mediaTypeForm)
	req.Header.Add(headerAccept, mediaTypeJSON)

	return req, nil
}

// NewJSONRequest creates an API request with a JSON body.
// The path is the relative URL which will be resolved to the BaseURL of the Client.
// It should always be specified without a preceding slash.
func (c *Client) NewJSONRequest(method string, path string, body interface{}) (*http.Request, error) {
	u, err := c.BaseURL.Parse(path)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	buf := new(bytes.Buffer)
	if body != nil {
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, &JSONError{
				Message: err.Error(),
				Data:    buf.Bytes(),
			}
		}
	}

	reqBody := bytes.NewReader(buf.Bytes())
	req, err := http.NewRequest(method, u.String(), reqBody)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	c.appendJSONExtensionToRequestURLPath(req)
	req.Header.Add(headerContentType, mediaTypeJSON)
	req.Header.Add(headerAccept, mediaTypeJSON)

	return req, nil
}

// parseRate parses the rate related headers.
func parseRate(r *http.Response) Rate {
	var rate Rate
	if remaining := r.Header.Get(headerRateLimitRemaining); remaining != "" {
		v, _ := strconv.ParseFloat(remaining, 64)
		rate.Remaining = int(v)
	}
	if used := r.Header.Get(headerRateLimitUsed); used != "" {
		rate.Used, _ = strconv.Atoi(used)
	}
	if reset := r.Header.Get(headerRateLimitReset); reset != "" {
		if v, _ := strconv.ParseInt(reset, 10, 64); v != 0 {
			rate.Reset = time.Now().Truncate(time.Second).Add(time.Second * time.Duration(v))
		}
	}
	return rate
}

// Do sends an API request and returns the API response. The API response is JSON decoded and stored in the value
// pointed to by v, or returned as an error if an API error has occurred. If v implements the io.Writer interface,
// the raw response will be written to v, without attempting to decode it.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	if err := c.checkRateLimitBeforeDo(req); err != nil {
		return nil, err
	}

	resp, err := DoRequestWithClient(ctx, c.client, req)
	if err != nil {
		return nil, &ResponseError{Message: err.Error(), Response: resp}
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
		}
	}(resp.Body)

	if c.onRequestCompleted != nil {
		c.onRequestCompleted(req, resp)
	}

	rate := parseRate(resp)

	c.rateMu.Lock()
	c.rate = rate
	c.rateMu.Unlock()

	if err = CheckResponse(resp); err != nil {
		return nil, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			if _, err = io.Copy(w, resp.Body); err != nil {
				return nil, &InternalError{
					Message: err.Error(),
				}
			}
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				data := make([]byte, resp.ContentLength)
				if _, err = resp.Body.Read(data); err != nil {
					return nil, &JSONError{
						Message: err.Error(),
						Data:    data,
					}
				}
				return nil, &JSONError{
					Message: err.Error(),
					Data:    data,
				}
			}
		}
	}

	return resp, nil
}

func (c *Client) PostURL(ctx context.Context, path string, form []byte) (*http.Response, error) {
	req, err := c.NewRequest(http.MethodPost, path, form)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return c.Do(ctx, req, nil)
}

func (c *Client) checkRateLimitBeforeDo(req *http.Request) *RateLimitError {
	c.rateMu.Lock()
	rate := c.rate
	c.rateMu.Unlock()

	if !rate.Reset.IsZero() && rate.Remaining == 0 && time.Now().Before(rate.Reset) {
		// Create a fake 429 response.
		resp := &http.Response{
			Status:     http.StatusText(http.StatusTooManyRequests),
			StatusCode: http.StatusTooManyRequests,
			Request:    req,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(strings.NewReader("")),
		}
		return &RateLimitError{
			Rate: rate,
			ResponseError: ResponseError{
				Response: resp,
				Message:  fmt.Sprintf("API rate limit still exceeded until %s, not making remote request.", rate.Reset)},
		}
	}

	return nil
}

// DoRequestWithClient submits an HTTP request using the specified client.
func DoRequestWithClient(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return client.Do(req)
}

// CheckResponse checks the API response for errors, and returns them if present.
// A response is considered an error if it has a status code outside the 200 range.
// Reddit also sometimes sends errors with 200 codes; we check for those too.
func CheckResponse(r *http.Response) error {
	if r.Header.Get(headerRateLimitRemaining) == "0" {
		rate := parseRate(r)
		return &RateLimitError{
			Rate: rate,
			ResponseError: ResponseError{
				Response: r,
				Message:  fmt.Sprintf("API rate limit has been exceeded until %s.", rate.Reset)},
		}
	}

	data, err := ioutil.ReadAll(r.Body)
	if err == nil {
		return &JSONError{Message: err.Error(), Data: data}
	}

	if c := r.StatusCode; c == 200 {
		return nil
	}

	return &ResponseError{Response: r, Message: err.Error()}
}

// Rate represents the rate limit for the client.
type Rate struct {
	// The number of remaining requests the client can make in the current 10-minute window.
	Remaining int `json:"remaining"`
	// The number of requests the client has made in the current 10-minute window.
	Used int `json:"used"`
	// The time at which the current rate limit will reset.
	Reset time.Time `json:"reset"`
}

func (c *Client) getListing(ctx context.Context, path string, opts interface{}) (*Listing, *http.Response, error) {
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	list := new(Listing)
	resp, err := c.Do(ctx, req, list)
	if err != nil {
		return nil, nil, err
	}

	return list, resp, nil
}

func (c *Client) getComment(ctx context.Context, path string, opts interface{}) (*Comment, *http.Response, error) {
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	comment := new(Comment)
	resp, err := c.Do(ctx, req, comment)
	if err != nil {
		return nil, nil, err
	}

	return comment, resp, nil
}

func (c *Client) getLink(ctx context.Context, path string, opts interface{}) (*Link, *http.Response, error) {
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	link := new(Link)
	resp, err := c.Do(ctx, req, link)
	if err != nil {
		return nil, nil, err
	}

	return link, resp, nil
}

func (c *Client) getSubreddit(ctx context.Context, path string, opts interface{}) (*Subreddit, *http.Response, error) {
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	subreddit := new(Subreddit)
	resp, err := c.Do(ctx, req, subreddit)
	if err != nil {
		return nil, nil, err
	}

	return subreddit, resp, nil
}

func (c *Client) getMessage(ctx context.Context, path string, opts interface{}) (*Message, *http.Response, error) {
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	message := new(Message)
	resp, err := c.Do(ctx, req, message)
	if err != nil {
		return nil, nil, err
	}

	return message, resp, nil
}

func (c *Client) getAccount(ctx context.Context, path string, opts interface{}) (*Account, *http.Response, error) {
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	account := new(Account)
	resp, err := c.Do(ctx, req, account)
	if err != nil {
		return nil, nil, err
	}

	return account, resp, nil
}

func (c *Client) getAward(ctx context.Context, path string, opts interface{}) (*Award, *http.Response, error) {
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	award := new(Award)
	resp, err := c.Do(ctx, req, award)
	if err != nil {
		return nil, nil, err
	}

	return award, resp, nil
}

func (c *Client) getMore(ctx context.Context, path string, opts interface{}) (*More, *http.Response, error) {
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := c.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	more := new(More)
	resp, err := c.Do(ctx, req, more)
	if err != nil {
		return nil, nil, err
	}

	return more, resp, nil
}

// ListingOptions specifies the optional parameters to various API calls that return a listing.
type ListingOptions struct {
	// Maximum number of items to be returned.
	// Generally, the default is 25 and max is 100.
	Limit int `url:"limit,omitempty"`
	// The full ID of an item in the listing to use
	// as the anchor point of the list. Only items
	// appearing after it will be returned.
	After string `url:"after,omitempty"`
	// The full ID of an item in the listing to use
	// as the anchor point of the list. Only items
	// appearing before it will be returned.
	Before   string `url:"before,omitempty"`
	Count    int    `url:"count,omitempty"`
	Show     string `url:"show,omitempty"`
	SrDetail bool   `url:"sr_detail,omitempty"`
	Name     string `url:"name,omitempty"`
}

type ListingDuplicateOptions struct {
	ListingOptions

	Article        string `url:"article,omitempty"`
	CrosspostsOnly bool   `url:"crossposts_only,omitempty"`
	Sort           string `url:"sort,omitempty"`
	SubredditName  string `url:"sr,omitempty"`
}

// ListingSubredditOptions defines possible options used when searching for subreddits.
type ListingSubredditOptions struct {
	ListingOptions
	// G is one of (GLOBAL, US, AR, AU, BG, CA, CL, CO, HR, CZ, FI, FR, DE, GR, HU, IS, IN, IE, IT, JP, MY, MX, NZ, PH, PL, PT, PR, RO, RS, SG, ES, SE, TW, TH, TR, GB, US_WA, US_DE, US_DC, US_WI, US_WV, US_HI, US_FL, US_WY, US_NH, US_NJ, US_NM, US_TX, US_LA, US_NC, US_ND, US_NE, US_TN, US_NY, US_PA, US_CA, US_NV, US_VA, US_CO, US_AK, US_AL, US_AR, US_VT, US_IL, US_GA, US_IN, US_IA, US_OK, US_AZ, US_ID, US_CT, US_ME, US_MD, US_MA, US_OH, US_UT, US_MO, US_MN, US_MI, US_RI, US_KS, US_MT, US_MS, US_SC, US_KY, US_OR, US_SD)
	// only for GET [/r/subreddit]/hot
	G string `url:"g,omitempty"`
	// T is one of (hour, day, week, month, year, all)
	// only for GET [/r/subreddit]/sort → [/r/subreddit]/top and [/r/subreddit]/controversial
	T string `url:"t,omitempty"`
	// User is a valid, existing reddit username
	// only for GET [/r/subreddit]/about/SubredditAboutWhere
	//→ [/r/subreddit]/about/banned
	//→ [/r/subreddit]/about/muted
	//→ [/r/subreddit]/about/wikibanned
	//→ [/r/subreddit]/about/contributors
	//→ [/r/subreddit]/about/wikicontributors
	//→ [/r/subreddit]/about/moderators
	User string `url:"user,omitempty"`

	// Q is a search query
	// only for GET /subreddits/search and GET /users/search
	Q string `url:"q,omitempty"`
	// SearchQueryID is a uuid
	// only for GET /subreddits/search and GET /users/search
	SearchQueryID string `url:"search_query_id,omitempty"`
	// ShowUsers is
	// only for GET /subreddits/search
	ShowUsers bool `url:"show_users,omitempty"`
	// TypeaheadActive is
	// only for GET /subreddits/search and GET /users/search
	TypeaheadActive *bool `url:"typeahead_active,omitempty"`
}

// ListingLiveOptions defines possible options used when searching for subreddits, only for GET /live/thread
type ListingLiveOptions struct {
	ListingOptions
	// Stylesr is a subreddit name
	// only for GET /live/thread
	Stylesr string `url:"stylesr,omitempty"`
}

// ListingMessageOptions , only for GET /message/SubredditAboutWhere → /message/inbox , /message/unread , /message/sent
type ListingMessageOptions struct {
	ListingOptions

	Mark       bool   `url:"mark,omitempty"`
	MaxReplies int    `url:"max_replies,omitempty"`
	Mid        string `url:"mid,omitempty"`
}

// ListingModerationOptions defines possible options used when getting moderation actions in a subreddit.
type ListingModerationOptions struct {
	ListingOptions

	// Moderator is a specified mod filter
	// only for GET [/r/subreddit]/about/log
	Moderator string `url:"mod,omitempty"`
	// Type is one of (banuser, unbanuser, spamlink, removelink, approvelink, spamcomment, removecomment, approvecomment, addmoderator, showcomment, invitemoderator, uninvitemoderator, acceptmoderatorinvite, removemoderator, addcontributor, removecontributor, editsettings, editflair, distinguish, marknsfw, wikibanned, wikicontributor, wikiunbanned, wikipagelisted, removewikicontributor, wikirevise, wikipermlevel, ignorereports, unignorereports, setpermissions, setsuggestedsort, sticky, unsticky, setcontestmode, unsetcontestmode, lock, unlock, muteuser, unmuteuser, createrule, editrule, reorderrules, deleterule, spoiler, unspoiler, modmail_enrollment, community_status, community_styling, community_welcome_page, community_widgets, markoriginalcontent, collections, events, hidden_award, add_community_topics, remove_community_topics, create_scheduled_post, edit_scheduled_post, delete_scheduled_post, submit_scheduled_post, edit_comment_requirements, edit_post_requirements, invitesubscriber, submit_content_rating_survey, adjust_post_crowd_control_level, enable_post_crowd_control_filter, disable_post_crowd_control_filter, deleteoverriddenclassification, overrideclassification, reordermoderators, snoozereports, unsnoozereports, addnote, deletenote, addremovalreason, createremovalreason, updateremovalreason, deleteremovalreason, reorderremovalreason, dev_platform_app_changed, dev_platform_app_disabled, dev_platform_app_enabled, dev_platform_app_installed, dev_platform_app_uninstalled, edit_saved_response, chat_approve_message, chat_remove_message, chat_ban_user, chat_unban_user, chat_invite_host, chat_remove_host, approve_award)
	// only for GET [/r/subreddit]/about/log
	Type string `url:"type,omitempty"`

	// Location is
	// only for GET [/r/subreddit]/about/locationread
	//→ [/r/subreddit]/about/reports
	//→ [/r/subreddit]/about/spam
	//→ [/r/subreddit]/about/modqueue
	//→ [/r/subreddit]/about/unmoderated
	//→ [/r/subreddit]/about/edited
	Location string `url:"location,omitempty"`
	// Only is one of (links, comments, chat_comments)
	// only for GET [/r/subreddit]/about/locationread
	//→ [/r/subreddit]/about/reports
	//→ [/r/subreddit]/about/spam
	//→ [/r/subreddit]/about/modqueue
	//→ [/r/subreddit]/about/unmoderated
	//→ [/r/subreddit]/about/edited
	Only string `url:"only,omitempty"`
}

// ListingSearchOptions defines possible options used when searching for posts within a subreddit.
// only for GET [/r/subreddit]/search
type ListingSearchOptions struct {
	ListingOptions
	// Category is a string no longer than 5 characters
	Category      string `url:"category,omitempty"`
	IncludeFacets bool   `url:"include_facets,omitempty"`
	// Q is a string no longer than 512 characters
	Q                 string `url:"q,omitempty"`
	RestrictSubreddit bool   `url:"restrict_sr,omitempty"`
	// Sort is one of (relevance, hot, top, new, comments)
	Sort string `url:"sort,omitempty"`
	// T is one of (hour, day, week, month, year, all)
	T string `url:"t,omitempty"`
	// Type is (optional) comma-delimited list of result types (sr, link, user)
	Type string `url:"type,omitempty"`
}

// ListingUserOptions is
// only for GET /user/username/SubredditAboutWhere
// → /user/username/overview
// → /user/username/submitted
// → /user/username/comments
// → /user/username/upvoted
// → /user/username/downvoted
// → /user/username/hidden
// → /user/username/saved
// → /user/username/gilded
type ListingUserOptions struct {
	ListingOptions
	// Context is an integer between 2 and 10
	Context int `url:"context,omitempty"`
	// Sort is one of (hot, new, top, controversial)
	Sort string `url:"sort,omitempty"`
	// T is one of (hour, day, week, month, year, all)
	T string `url:"t,omitempty"`
	// Type is one of (links, comments)
	Type string `url:"type,omitempty"`
	// Username is the name of an existing user
	Username string `url:"username,omitempty"`
}

// ListingWikiOptions is
// only for GET [/r/subreddit]/wiki/discussions/page and GET [/r/subreddit]/wiki/revisions/page
type ListingWikiOptions struct {
	ListingOptions
	// Page is the name of an existing wiki page
	Page string `url:"page"`
}

func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	origURL, err := url.Parse(s)
	if err != nil {
		return s, &InternalError{Message: err.Error()}
	}

	origValues := origURL.Query()

	newValues, err := query.Values(opt)
	if err != nil {
		return s, &InternalError{Message: err.Error()}
	}

	for k, v := range newValues {
		origValues[k] = v
	}

	origURL.RawQuery = origValues.Encode()
	return origURL.String(), nil
}
