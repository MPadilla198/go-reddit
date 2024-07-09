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
	"strconv"
	"strings"
	"sync"
	"time"

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
	Message        *MessageService
	Moderation     *ModerationService
	Multi          *MultiService
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
	client.Message = &MessageService{client: client}
	client.Moderation = &ModerationService{client: client}
	client.Multi = &MultiService{client: client}
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

// Do send an API request and returns the API response. The API response is JSON decoded and stored in the value
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
	req, err := c.NewJSONRequest(http.MethodGet, path, opts)
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
	req, err := c.NewJSONRequest(http.MethodGet, path, opts)
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
	req, err := c.NewJSONRequest(http.MethodGet, path, opts)
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
	req, err := c.NewJSONRequest(http.MethodGet, path, opts)
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
	req, err := c.NewJSONRequest(http.MethodGet, path, opts)
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
	req, err := c.NewJSONRequest(http.MethodGet, path, opts)
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
	req, err := c.NewJSONRequest(http.MethodGet, path, opts)
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
	req, err := c.NewJSONRequest(http.MethodGet, path, opts)
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
	Limit    int    `json:"limit"`            // Maximum number of items to be returned, the default is 25 and max is 100.
	After    string `json:"after,omitempty"`  //  the fullname of an item. Only one of After and Before may be specified in a request
	Before   string `json:"before,omitempty"` //  the fullname of an item. Only one of After and Before may be specified in a request
	Count    int    `json:"count"`            // The number of items already seen in this listing. on the html site, the builder uses this to determine when to give values for before and after in the response. Default is 0
	Show     string `json:"show,omitempty"`   // (optional) if "all" is passed, filters such as "hide links that I have voted on" will be disabled.
	SrDetail bool   `json:"sr_detail,omitempty"`
	Name     string `json:"name,omitempty"`
}

type ListingDuplicateSortType string

const (
	ListingDuplicateSortNumComments ListingDuplicateSortType = "num_comments"
	ListingDuplicateSortNew         ListingDuplicateSortType = "new"
)

type ListingDuplicateOptions struct {
	ListingOptions

	Article        string                   `json:"article"`
	CrosspostsOnly bool                     `json:"crossposts_only"`
	Sort           ListingDuplicateSortType `json:"sort"` // one of (num_comments, new)
	SubredditName  string                   `json:"sr"`
}

type ListingRegionCodes string

const (
	ListingRegionCodeGlobal             ListingRegionCodes = "GLOBAL"
	ListingRegionCodeUnitedStates       ListingRegionCodes = "US"
	ListingRegionCodeArgentina          ListingRegionCodes = "AR"
	ListingRegionCodeAustralia          ListingRegionCodes = "AU"
	ListingRegionCodeBulgaria           ListingRegionCodes = "BG"
	ListingRegionCodeCanada             ListingRegionCodes = "CA"
	ListingRegionCodeChile              ListingRegionCodes = "CL"
	ListingRegionCodeColombia           ListingRegionCodes = "CO"
	ListingRegionCodeCroatia            ListingRegionCodes = "HR"
	ListingRegionCodeCzechia            ListingRegionCodes = "CZ"
	ListingRegionCodeFinland            ListingRegionCodes = "FI"
	ListingRegionCodeFrance             ListingRegionCodes = "FR"
	ListingRegionCodeGermany            ListingRegionCodes = "DE"
	ListingRegionCodeGreece             ListingRegionCodes = "GR"
	ListingRegionCodeHungary            ListingRegionCodes = "HU"
	ListingRegionCodeIceland            ListingRegionCodes = "IS"
	ListingRegionCodeIndia              ListingRegionCodes = "IN"
	ListingRegionCodeIreland            ListingRegionCodes = "IE"
	ListingRegionCodeItaly              ListingRegionCodes = "IT"
	ListingRegionCodeJapan              ListingRegionCodes = "JP"
	ListingRegionCodeMalaysia           ListingRegionCodes = "MY"
	ListingRegionCodeMexico             ListingRegionCodes = "MX"
	ListingRegionCodeNewZealand         ListingRegionCodes = "NZ"
	ListingRegionCodePhilippines        ListingRegionCodes = "PH"
	ListingRegionCodePoland             ListingRegionCodes = "PL"
	ListingRegionCodePortugal           ListingRegionCodes = "PT"
	ListingRegionCodePuertoRica         ListingRegionCodes = "PR"
	ListingRegionCodeRomania            ListingRegionCodes = "RO"
	ListingRegionCodeSerbia             ListingRegionCodes = "RS"
	ListingRegionCodeSingapore          ListingRegionCodes = "SG"
	ListingRegionCodeSpain              ListingRegionCodes = "ES"
	ListingRegionCodeSweden             ListingRegionCodes = "SE"
	ListingRegionCodeTaiwan             ListingRegionCodes = "TW"
	ListingRegionCodeThailand           ListingRegionCodes = "TH"
	ListingRegionCodeTurkey             ListingRegionCodes = "TR"
	ListingRegionCodeUnitedKingdom      ListingRegionCodes = "GB"
	ListingRegionCodeWashington         ListingRegionCodes = "US_WA"
	ListingRegionCodeDelaware           ListingRegionCodes = "US_DE"
	ListingRegionCodeDistrictOfColumbia ListingRegionCodes = "US_DC"
	ListingRegionCodeWisconsin          ListingRegionCodes = "US_WI"
	ListingRegionCodeWestVirginia       ListingRegionCodes = "US_WV"
	ListingRegionCodeHawaii             ListingRegionCodes = "US_HI"
	ListingRegionCodeFlorida            ListingRegionCodes = "US_FL"
	ListingRegionCodeWyoming            ListingRegionCodes = "US_WY"
	ListingRegionCodeNewHampshire       ListingRegionCodes = "US_NH"
	ListingRegionCodeNewJersey          ListingRegionCodes = "US_NJ"
	ListingRegionCodeNewMexico          ListingRegionCodes = "US_NM"
	ListingRegionCodeTexas              ListingRegionCodes = "US_TX"
	ListingRegionCodeLouisiana          ListingRegionCodes = "US_LA"
	ListingRegionCodeNorthCarolina      ListingRegionCodes = "US_NC"
	ListingRegionCodeNorthDakota        ListingRegionCodes = "US_ND"
	ListingRegionCodeNebraska           ListingRegionCodes = "US_NE"
	ListingRegionCodeTennessee          ListingRegionCodes = "US_TN"
	ListingRegionCodeNewYork            ListingRegionCodes = "US_NY"
	ListingRegionCodePennsylvania       ListingRegionCodes = "US_PA"
	ListingRegionCodeCalifornia         ListingRegionCodes = "US_CA"
	ListingRegionCodeNevada             ListingRegionCodes = "US_NV"
	ListingRegionCodeVirginia           ListingRegionCodes = "US_VA"
	ListingRegionCodeColorado           ListingRegionCodes = "US_CO"
	ListingRegionCodeAlaska             ListingRegionCodes = "US_AK"
	ListingRegionCodeAlabama            ListingRegionCodes = "US_AL"
	ListingRegionCodeArkansas           ListingRegionCodes = "US_AR"
	ListingRegionCodeVermont            ListingRegionCodes = "US_VT"
	ListingRegionCodeIllinois           ListingRegionCodes = "US_IL"
	ListingRegionCodeGeorgia            ListingRegionCodes = "US_GA"
	ListingRegionCodeIndiana            ListingRegionCodes = "US_IN"
	ListingRegionCodeIowa               ListingRegionCodes = "US_IA"
	ListingRegionCodeOklahoma           ListingRegionCodes = "US_OK"
	ListingRegionCodeArizona            ListingRegionCodes = "US_AZ"
	ListingRegionCodeIdaho              ListingRegionCodes = "US_ID"
	ListingRegionCodeConnecticut        ListingRegionCodes = "US_CT"
	ListingRegionCodeMaine              ListingRegionCodes = "US_ME"
	ListingRegionCodeMaryland           ListingRegionCodes = "US_MD"
	ListingRegionCodeMassachusetts      ListingRegionCodes = "US_MA"
	ListingRegionCodeOhio               ListingRegionCodes = "US_OH"
	ListingRegionCodeUtah               ListingRegionCodes = "US_UT"
	ListingRegionCodeMissouri           ListingRegionCodes = "US_MO"
	ListingRegionCodeMinnesota          ListingRegionCodes = "US_MN"
	ListingRegionCodeMichigan           ListingRegionCodes = "US_MI"
	ListingRegionCodeRhodeIsland        ListingRegionCodes = "US_RI"
	ListingRegionCodeKansas             ListingRegionCodes = "US_KS"
	ListingRegionCodeMontana            ListingRegionCodes = "US_MT"
	ListingRegionCodeMississippi        ListingRegionCodes = "US_MS"
	ListingRegionCodeSouthCarolina      ListingRegionCodes = "US_SC"
	ListingRegionCodeKentucky           ListingRegionCodes = "US_KY"
	ListingRegionCodeOregon             ListingRegionCodes = "US_OR"
	ListingRegionCodeSouthDakota        ListingRegionCodes = "US_SD"
)

type ListingSubredditSortOptions struct {
	ListingOptions

	G ListingRegionCodes `json:"g,omitempty"` // only for GET [/r/subreddit]/hot
}

type ListingTimingType string

const (
	ListingTimingHour  ListingTimingType = "hour"
	ListingTimingDay   ListingTimingType = "day"
	ListingTimingWeek  ListingTimingType = "week"
	ListingTimingMonth ListingTimingType = "month"
	ListingTimingYear  ListingTimingType = "year"
	ListingTimingAll   ListingTimingType = "all"
)

// ListingSubredditOptions defines possible options used when searching for subreddits.
type ListingSubredditOptions struct {
	ListingOptions

	T               ListingTimingType `json:"t,omitempty"`                // only for GET [/r/subreddit]/sort → [/r/subreddit]/top and [/r/subreddit]/controversial
	User            string            `json:"user,omitempty"`             // User is a valid, existing reddit username, only for GET [/r/subreddit]/about/SubredditAboutWhere → [banned, muted, wikibanned, contributors, wikicontributors, moderators]
	Q               string            `json:"q,omitempty"`                // Q is a search query, only for GET /subreddits/search and GET /users/search
	SearchQueryID   string            `json:"search_query_id,omitempty"`  // SearchQueryID is an uuid, only for GET /subreddits/search and GET /users/search
	ShowUsers       bool              `json:"show_users,omitempty"`       // ShowUsers is only for GET /subreddits/search
	TypeaheadActive *bool             `json:"typeahead_active,omitempty"` // TypeaheadActive is only for GET /subreddits/search and GET /users/search
}

// ListingLiveOptions defines possible options used when searching for subreddits, only for GET /live/thread
type ListingLiveOptions struct {
	ListingOptions

	Stylesr string `json:"stylesr,omitempty"` // Stylesr is a subreddit name, only for GET /live/thread
}

// ListingMessageOptions , only for GET /message/SubredditAboutWhere → [inbox, unread, sent]
type ListingMessageOptions struct {
	ListingOptions

	Mark       bool   `json:"mark,omitempty"`
	MaxReplies int    `json:"max_replies,omitempty"`
	Mid        string `json:"mid,omitempty"`
}

type ListingModerationActionType string

const (
	ModerationActionBanUser                        ListingModerationActionType = "banuser"
	ModerationActionUnbanUser                      ListingModerationActionType = "unbanuser"
	ModerationActionSpamLink                       ListingModerationActionType = "spamlink"
	ModerationActionRemoveLink                     ListingModerationActionType = "removelink"
	ModerationActionApproveLink                    ListingModerationActionType = "approvelink"
	ModerationActionSpamComment                    ListingModerationActionType = "spamcomment"
	ModerationActionRemoveComment                  ListingModerationActionType = "removecomment"
	ModerationActionApproveComment                 ListingModerationActionType = "approvecomment"
	ModerationActionAddModerator                   ListingModerationActionType = "addmoderator"
	ModerationActionShowComment                    ListingModerationActionType = "showcomment"
	ModerationActionInviteModerator                ListingModerationActionType = "invitemoderator"
	ModerationActionUninviteModerator              ListingModerationActionType = "uninvitemoderator"
	ModerationActionAcceptModeratorInvite          ListingModerationActionType = "acceptmoderatorinvite"
	ModerationActionRemoveModerator                ListingModerationActionType = "removemoderator"
	ModerationActionAddContributor                 ListingModerationActionType = "addcontributor"
	ModerationActionRemoveContributor              ListingModerationActionType = "removecontributor"
	ModerationActionEditSettings                   ListingModerationActionType = "editsettings"
	ModerationActionEditFlair                      ListingModerationActionType = "editflair"
	ModerationActionDistinguish                    ListingModerationActionType = "distinguish"
	ModerationActionMarkNSFW                       ListingModerationActionType = "marknsfw"
	ModerationActionWikiBanned                     ListingModerationActionType = "wikibanned"
	ModerationActionWikiContributor                ListingModerationActionType = "wikicontributor"
	ModerationActionWikiUnbanned                   ListingModerationActionType = "wikiunbanned"
	ModerationActionWikiPageListed                 ListingModerationActionType = "wikipagelisted"
	ModerationActionRemoveWikiContributor          ListingModerationActionType = "removewikicontributor"
	ModerationActionWikiRevise                     ListingModerationActionType = "wikirevise"
	ModerationActionWikiPermissionLevel            ListingModerationActionType = "wikipermlevel"
	ModerationActionIgnoreReports                  ListingModerationActionType = "ignorereports"
	ModerationActionUnignoreReports                ListingModerationActionType = "unignorereports"
	ModerationActionSetPermissions                 ListingModerationActionType = "setpermissions"
	ModerationActionSetSuggestedSort               ListingModerationActionType = "setsuggestedsort"
	ModerationActionSticky                         ListingModerationActionType = "sticky"
	ModerationActionUnsticky                       ListingModerationActionType = "unsticky"
	ModerationActionSetContestMode                 ListingModerationActionType = "setcontestmode"
	ModerationActionUnsetContestMode               ListingModerationActionType = "unsetcontestmode"
	ModerationActionLock                           ListingModerationActionType = "lock"
	ModerationActionUnlock                         ListingModerationActionType = "unlock"
	ModerationActionMuteUser                       ListingModerationActionType = "muteuser"
	ModerationActionUnmuteUser                     ListingModerationActionType = "unmuteuser"
	ModerationActionCreateRule                     ListingModerationActionType = "createrule"
	ModerationActionEditRule                       ListingModerationActionType = "editrule"
	ModerationActionReorderRules                   ListingModerationActionType = "reorderrules"
	ModerationActionDeleteRule                     ListingModerationActionType = "deleterule"
	ModerationActionSpoiler                        ListingModerationActionType = "spoiler"
	ModerationActionUnspoiler                      ListingModerationActionType = "unspoiler"
	ModerationActionModmailEnrollment              ListingModerationActionType = "modmail_enrollment"
	ModerationActionCommunityStatus                ListingModerationActionType = "community_status"
	ModerationActionCommunityStyling               ListingModerationActionType = "community_styling"
	ModerationActionCommunityWelcomePage           ListingModerationActionType = "community_welcome_page"
	ModerationActionCommunityWidgets               ListingModerationActionType = "community_widgets"
	ModerationActionMarkOriginalContent            ListingModerationActionType = "markoriginalcontent"
	ModerationActionCollections                    ListingModerationActionType = "collections"
	ModerationActionEvents                         ListingModerationActionType = "events"
	ModerationActionHiddenAward                    ListingModerationActionType = "hidden_award"
	ModerationActionAddCommunityTopics             ListingModerationActionType = "add_community_topics"
	ModerationActionRemoveCommunityTopics          ListingModerationActionType = "remove_community_topics"
	ModerationActionCreateScheduledPost            ListingModerationActionType = "create_scheduled_post"
	ModerationActionEditScheduledPost              ListingModerationActionType = "edit_scheduled_post"
	ModerationActionDeleteScheduledPost            ListingModerationActionType = "delete_scheduled_post"
	ModerationActionSubmitScheduledPost            ListingModerationActionType = "submit_scheduled_post"
	ModerationActionEditCommentRequirements        ListingModerationActionType = "edit_comment_requirements"
	ModerationActionEditPostRequirements           ListingModerationActionType = "edit_post_requirements"
	ModerationActionInviteSubscriber               ListingModerationActionType = "invitesubscriber"
	ModerationActionSubmitContestRatingSurvey      ListingModerationActionType = "submit_content_rating_survey"
	ModerationActionAdjustPostCrowControlLevel     ListingModerationActionType = "adjust_post_crowd_control_level"
	ModerationActionEnablePostCrowdControlFilter   ListingModerationActionType = "enable_post_crowd_control_filter"
	ModerationActionDisablePostCrowdControlFilter  ListingModerationActionType = "disable_post_crowd_control_filter"
	ModerationActionDeleteOverriddenClassification ListingModerationActionType = "deleteoverriddenclassification"
	ModerationAction                               ListingModerationActionType = "overrideclassification"
	ModerationActionReorderModerators              ListingModerationActionType = "reordermoderators"
	ModerationActionSnoozeReports                  ListingModerationActionType = "snoozereports"
	ModerationActionUnsnoozeReports                ListingModerationActionType = "unsnoozereports"
	ModerationActionAddNote                        ListingModerationActionType = "addnote"
	ModerationActionDeleteNote                     ListingModerationActionType = "deletenote"
	ModerationActionAddRemovalReason               ListingModerationActionType = "addremovalreason"
	ModerationActionCreateRemovalReason            ListingModerationActionType = "createremovalreason"
	ModerationActionUpdateRemovalReason            ListingModerationActionType = "updateremovalreason"
	ModerationActionDeleteRemovalReason            ListingModerationActionType = "deleteremovalreason"
	ModerationActionReorderRemovalReason           ListingModerationActionType = "reorderremovalreason"
	ModerationActionDevPlatformAppChanged          ListingModerationActionType = "dev_platform_app_changed"
	ModerationActionDevPlatformAppDisabled         ListingModerationActionType = "dev_platform_app_disabled"
	ModerationActionDevPlatformAppEnabled          ListingModerationActionType = "dev_platform_app_enabled"
	ModerationActionDevPlatformAppInstalled        ListingModerationActionType = "dev_platform_app_installed"
	ModerationActionDevPlatformAppUninstalled      ListingModerationActionType = "dev_platform_app_uninstalled"
	ModerationActionEditSavedResponse              ListingModerationActionType = "edit_saved_response"
	ModerationActionChatApproveMessage             ListingModerationActionType = "chat_approve_message"
	ModerationActionChatRemoveMessage              ListingModerationActionType = "chat_remove_message"
	ModerationActionChatBanUser                    ListingModerationActionType = "chat_ban_user"
	ModerationActionChatUnbanUser                  ListingModerationActionType = "chat_unban_user"
	ModerationActionChatInviteHost                 ListingModerationActionType = "chat_invite_host"
	ModerationActionChatRemoveHost                 ListingModerationActionType = "chat_remove_host"
	ModerationActionApproveAward                   ListingModerationActionType = "approve_award"
)

type ListingModerationOnlyType string

const (
	ListingModerationOnlyLinks        ListingModerationOnlyType = "links"
	ListingModerationOnlyComments     ListingModerationOnlyType = "comments"
	ListingModerationOnlyChatComments ListingModerationOnlyType = "chat_comments"
)

// ListingModerationOptions defines possible options used when getting moderation actions in a subreddit.
type ListingModerationOptions struct {
	ListingOptions

	Moderator string                      `json:"mod,omitempty"`      // Moderator is a specified mod filter, only for GET [/r/subreddit]/about/log
	Type      ListingModerationActionType `json:"type"`               // only for GET [/r/subreddit]/about/log
	Location  string                      `json:"location,omitempty"` // Location is only for GET [/r/subreddit]/about/locationread → [reports, spam, modqueue, unmoderated, edited]
	Only      ListingModerationOnlyType   `json:"only,omitempty"`     // Only is for GET [/r/subreddit]/about/locationread → [reports, spam, modqueue, unmoderated, edited]
}

type ListingSearchSortType string

const (
	ListingSearchSortRelevance ListingSearchSortType = "relevance"
	ListingSearchSortHot       ListingSearchSortType = "hot"
	ListingSearchSortTop       ListingSearchSortType = "top"
	ListingSearchSortNew       ListingSearchSortType = "new"
	ListingSearchSortComments  ListingSearchSortType = "comments"
)

type ListingSearchResultType string

const (
	ListingSearchSubreddit ListingSearchResultType = "sr"
	ListingSearchLink      ListingSearchResultType = "link"
	ListingSearchUser      ListingSearchResultType = "user"
)

// ListingSearchOptions defines possible options used when searching for posts within a subreddit.
// only for GET [/r/subreddit]/search
type ListingSearchOptions struct {
	ListingOptions

	Category          string                    `json:"category"`       // Category is a string no longer than 5 characters
	IncludeFacets     bool                      `json:"include_facets"` //
	Q                 string                    `json:"q"`              // Q is a string no longer than 512 characters
	RestrictSubreddit bool                      `json:"restrict_sr"`
	Sort              ListingSearchSortType     `json:"sort"`
	T                 ListingTimingType         `json:"t"`
	Type              []ListingSearchResultType `json:"type,omitempty"` // Type is (optional) comma-delimited list of result types (sr, link, user)
}

type ListingUserSortType string

const (
	ListingUserSortHot           ListingUserSortType = "hot"
	ListingUserSortNew           ListingUserSortType = "new"
	ListingUserSortTop           ListingUserSortType = "top"
	ListingUserSortControversial ListingUserSortType = "controversial"
)

type ListingUserType string

const (
	ListingUserLinks    ListingUserType = "links"
	ListingUserComments ListingUserType = "comments"
)

// ListingUserOptions is
// only for GET /user/username/SubredditAboutWhere → [overview, submitted, comments, upvoted, downvoted, hidden, saved, gilded]
type ListingUserOptions struct {
	ListingOptions

	Context  int                 `json:"context,omitempty"` // Context is an integer between 2 and 10
	Sort     ListingUserSortType `json:"sort,omitempty"`
	T        ListingTimingType   `json:"t,omitempty"`
	Type     ListingUserType     `json:"type,omitempty"`
	Username string              `json:"username,omitempty"` // Username is the name of an existing user
}

// ListingWikiOptions is
// only for GET [/r/subreddit]/wiki/discussions/page and GET [/r/subreddit]/wiki/revisions/page
type ListingWikiOptions struct {
	ListingOptions

	Page string `json:"page"` // Page is the name of an existing wiki page
}
