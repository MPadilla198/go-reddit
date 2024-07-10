package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// WidgetService handles communication with the widget
// related methods of the Reddit API.
//
// Reddit API docs: https://www.reddit.com/dev/api/#section_widgets
type WidgetService struct {
	client *Client
}

type WidgetKind string

const (
	WidgetKindButton         WidgetKind = "button"
	WidgetKindCalendar       WidgetKind = "calendar"
	WidgetKindCommunityList  WidgetKind = "community-list"
	WidgetKindCustom         WidgetKind = "custom"
	WidgetKindIDCard         WidgetKind = "id-card"
	WidgetKindImage          WidgetKind = "image"
	WidgetKindMenu           WidgetKind = "menu"
	WidgetKindModerators     WidgetKind = "moderators"
	WidgetKindPostFlair      WidgetKind = "post-flair"
	WidgetKindSubredditRules WidgetKind = "subreddit-rules"
	WidgetKindText           WidgetKind = "text"
	WidgetKindTextArea       WidgetKind = "textarea"
)

type Widget interface {
	json.Marshaler
	json.Unmarshaler

	Kind() WidgetKind
}

const (
	WidgetMarshallingErrorPrefix   = "error while marshalling widget: "
	WidgetUnmarshallingErrorPrefix = "error while unmarshalling widget: "

	WidgetUnmarshallingTypeErrorMessage = WidgetUnmarshallingErrorPrefix + "unmarshalled widget is not of type "
)

type WidgetStyles struct {
	BackgroundColor string `json:"backgroundColor"` // a 6-digit rgb hex color, e.g. `#AABBCC`
	HeaderColor     string `json:"headerColor"`     // a 6-digit rgb hex color, e.g. `#AABBCC`
}

type WidgetImageData struct {
	Height  int    `json:"height"`
	LinkURL string `json:"link_url,omitempty"` // (optional) valid URL
	URL     string `json:"url"`                // valid URL of a reddit-hosted image
	Width   int    `json:"width"`
}

type WidgetImages struct {
	Data      []WidgetImageData
	ShortName [30]byte
	Styles    WidgetStyles
}

type widgetImages struct {
	Data      []WidgetImageData `json:"data"`
	Kind      WidgetKind        `json:"kind"`
	ShortName [30]byte          `json:"shortName"`
	Styles    WidgetStyles      `json:"styles"`
}

func (imgs *WidgetImages) MarshalJSON() ([]byte, error) {
	temp := widgetImages{Data: imgs.Data, Kind: WidgetKindImage, ShortName: imgs.ShortName, Styles: imgs.Styles}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (imgs *WidgetImages) UnmarshalJSON(data []byte) error {
	const KIND = WidgetKindImage
	temp := new(widgetImages)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != KIND {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(KIND),
			Data:    data,
		}
	}

	imgs.Data = temp.Data
	imgs.ShortName = temp.ShortName
	imgs.Styles = temp.Styles

	return nil
}

func (_ *WidgetImages) Kind() WidgetKind {
	return WidgetKindImage
}

type WidgetCalendarConfiguration struct {
	NumEvents       int // an integer between 1 and 50 (default: 10)
	ShowDate        bool
	ShowDescription bool
	ShowLocation    bool
	ShowTime        bool
	ShowTitle       bool
}

type widgetCalendarConfiguration struct {
	NumEvents       int  `json:"numEvents"` // an integer between 1 and 50 (default: 10)
	ShowDate        bool `json:"showDate"`
	ShowDescription bool `json:"showDescription"`
	ShowLocation    bool `json:"showLocation"`
	ShowTime        bool `json:"showTime"`
	ShowTitle       bool `json:"showTitle"`
}

type WidgetCalendar struct {
	Configuration    WidgetCalendarConfiguration
	GoogleCalendarID string // a valid email address
	RequiresSync     bool
	ShortName        [30]byte
	Styles           WidgetStyles
}

type widgetCalendar struct {
	Configuration    widgetCalendarConfiguration `json:"configuration"`
	GoogleCalendarID string                      `json:"googleCalendarId"` // a valid email address
	Kind             WidgetKind                  `json:"kind"`             // only 'calendar'
	RequiresSync     bool                        `json:"requiresSync"`
	ShortName        [30]byte                    `json:"shortName"`
	Styles           WidgetStyles                `json:"styles"`
}

func (cal *WidgetCalendar) MarshalJSON() ([]byte, error) {
	temp := widgetCalendar{
		Configuration: widgetCalendarConfiguration{
			NumEvents:       cal.Configuration.NumEvents,
			ShowDate:        cal.Configuration.ShowDate,
			ShowDescription: cal.Configuration.ShowDescription,
			ShowLocation:    cal.Configuration.ShowLocation,
			ShowTime:        cal.Configuration.ShowTime,
			ShowTitle:       cal.Configuration.ShowTitle,
		},
		GoogleCalendarID: cal.GoogleCalendarID,
		Kind:             WidgetKindCalendar,
		RequiresSync:     cal.RequiresSync,
		ShortName:        cal.ShortName,
		Styles: WidgetStyles{
			BackgroundColor: cal.Styles.BackgroundColor,
			HeaderColor:     cal.Styles.HeaderColor,
		},
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (cal *WidgetCalendar) UnmarshalJSON(data []byte) error {
	const KIND = WidgetKindCalendar
	temp := new(widgetCalendar)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != KIND {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(KIND),
			Data:    data,
		}
	}

	cal.Configuration.NumEvents = temp.Configuration.NumEvents
	cal.Configuration.ShowDate = temp.Configuration.ShowDate
	cal.Configuration.ShowDescription = temp.Configuration.ShowDescription
	cal.Configuration.ShowLocation = temp.Configuration.ShowLocation
	cal.Configuration.ShowTime = temp.Configuration.ShowTime
	cal.Configuration.ShowTitle = temp.Configuration.ShowTitle
	cal.GoogleCalendarID = temp.GoogleCalendarID
	cal.RequiresSync = temp.RequiresSync
	cal.ShortName = temp.ShortName
	cal.Styles.BackgroundColor = temp.Styles.BackgroundColor
	cal.Styles.HeaderColor = temp.Styles.HeaderColor

	return nil
}

func (_ *WidgetCalendar) Kind() WidgetKind {
	return WidgetKindCalendar
}

type WidgetTextArea struct {
	ShortName [30]byte
	Styles    WidgetStyles
	Text      string // raw Markdown text
}

type widgetTextArea struct {
	Kind      WidgetKind   `json:"kind"` // only 'textarea'
	ShortName [30]byte     `json:"shortName"`
	Styles    WidgetStyles `json:"styles"`
	Text      string       `json:"text"` // raw Markdown text
}

func (txt *WidgetTextArea) MarshalJSON() ([]byte, error) {
	temp := widgetTextArea{
		Kind:      WidgetKindTextArea,
		ShortName: txt.ShortName,
		Styles: WidgetStyles{
			BackgroundColor: txt.Styles.BackgroundColor,
			HeaderColor:     txt.Styles.HeaderColor,
		},
		Text: txt.Text,
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (txt *WidgetTextArea) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindTextArea
	temp := new(widgetTextArea)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	txt.ShortName = temp.ShortName
	txt.Styles = WidgetStyles{
		BackgroundColor: temp.Styles.BackgroundColor,
		HeaderColor:     temp.Styles.HeaderColor,
	}
	txt.Text = temp.Text

	return nil
}

func (_ *WidgetTextArea) Kind() WidgetKind {
	return WidgetKindTextArea
}

type WidgetSubredditRulesDisplayType string

const (
	WidgetSubredditRulesDisplayFull    WidgetSubredditRulesDisplayType = "full"
	WidgetSubredditRulesDisplayCompact WidgetSubredditRulesDisplayType = "compact"
)

type WidgetSubredditRules struct {
	Display   WidgetSubredditRulesDisplayType
	ShortName [30]byte
	Styles    WidgetStyles
}

type widgetSubredditRules struct {
	Display   WidgetSubredditRulesDisplayType `json:"display"`
	Kind      WidgetKind                      `json:"kind"`
	ShortName [30]byte                        `json:"shortName"`
	Styles    WidgetStyles                    `json:"styles"`
}

func (rule *WidgetSubredditRules) MarshalJSON() ([]byte, error) {
	temp := widgetSubredditRules{
		Display:   rule.Display,
		Kind:      WidgetKindSubredditRules,
		ShortName: rule.ShortName,
		Styles: WidgetStyles{
			BackgroundColor: rule.Styles.BackgroundColor,
			HeaderColor:     rule.Styles.HeaderColor,
		},
	}
	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (rule *WidgetSubredditRules) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindSubredditRules
	temp := new(widgetSubredditRules)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	rule.Display = temp.Display
	rule.ShortName = temp.ShortName
	rule.Styles.BackgroundColor = temp.Styles.BackgroundColor
	rule.Styles.HeaderColor = temp.Styles.HeaderColor

	return nil
}

func (_ *WidgetSubredditRules) Kind() WidgetKind {
	return WidgetKindSubredditRules
}

type WidgetMenuDataChild struct {
	Text [20]byte
	URL  string // a valid url
}

type widgetMenuDataChild struct {
	Text [20]byte `json:"text"`
	URL  string   `json:"url"` // a valid url
}

type WidgetMenuData struct {
	Children []WidgetMenuDataChild
	Text     [20]byte
	URL      string // a valid url
}

type widgetMenuData struct {
	Children []widgetMenuDataChild `json:"children,omitempty"`
	Text     [20]byte              `json:"text"`
	URL      string                `json:"url,omitempty"` // a valid url
}

type WidgetMenu struct {
	Data     []WidgetMenuData // If no url, then children are needed
	ShowWiki bool
}

type widgetMenu struct {
	Data     []widgetMenuData `json:"data"`
	Kind     WidgetKind       `json:"kind"` // only 'menu'
	ShowWiki bool             `json:"showWiki"`
}

func (menu *WidgetMenu) MarshalJSON() ([]byte, error) {
	dataLen := len(menu.Data)
	temp := widgetMenu{
		Data:     make([]widgetMenuData, dataLen),
		Kind:     WidgetKindMenu,
		ShowWiki: menu.ShowWiki,
	}
	for i := 0; i < dataLen; i++ {
		temp.Data[i].Children = make([]widgetMenuDataChild, len(menu.Data[i].Children))
		for j := 0; j < len(menu.Data[i].Children); j++ {
			temp.Data[i].Children[j] = widgetMenuDataChild{
				Text: menu.Data[i].Children[j].Text,
				URL:  menu.Data[i].Children[j].URL,
			}
		}
		temp.Data[i].Text = menu.Data[i].Text
		temp.Data[i].URL = menu.Data[i].URL
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (menu *WidgetMenu) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindMenu
	temp := new(widgetMenu)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	menu.Data = make([]WidgetMenuData, len(temp.Data))
	for i := 0; i < len(menu.Data); i++ {
		datumLength := len(temp.Data[i].Children)
		menu.Data[i].Children = make([]WidgetMenuDataChild, datumLength)
		for j := 0; j < datumLength; j++ {
			menu.Data[i].Children[j] = WidgetMenuDataChild{
				Text: temp.Data[i].Children[j].Text,
				URL:  temp.Data[i].Children[j].URL,
			}
		}

		menu.Data[i].Text = temp.Data[i].Text
		menu.Data[i].URL = temp.Data[i].URL
	}
	menu.ShowWiki = temp.ShowWiki

	return nil
}

func (_ *WidgetMenu) Kind() WidgetKind {
	return WidgetKindMenu
}

type WidgetHoverState interface {
	Widget
}

type WidgetHoverStateText struct {
	Color     string // a 6-digit rgb hex color, e.g. `#AABBCC`
	FillColor string // a 6-digit rgb hex color, e.g. `#AABBCC`
	Text      string
	TextColor string // a 6-digit rgb hex color, e.g. `#AABBCC`
}

type widgetHoverStateText struct {
	Color     string     `json:"color"`     // a 6-digit rgb hex color, e.g. `#AABBCC`
	FillColor string     `json:"fillColor"` // a 6-digit rgb hex color, e.g. `#AABBCC`
	Kind      WidgetKind `json:"kind"`      // Only 'text'
	Text      string     `json:"text"`
	TextColor string     `json:"textColor"` // a 6-digit rgb hex color, e.g. `#AABBCC`
}

func (txt *WidgetHoverStateText) MarshalJSON() ([]byte, error) {
	temp := widgetHoverStateText{
		Color:     txt.Color,
		FillColor: txt.FillColor,
		Kind:      WidgetKindText,
		Text:      txt.Text,
		TextColor: txt.TextColor,
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (txt *WidgetHoverStateText) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindText
	temp := new(widgetHoverStateText)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	txt.Color = temp.Color
	txt.FillColor = temp.FillColor
	txt.Text = temp.Text
	txt.TextColor = temp.TextColor

	return nil
}

func (_ *WidgetHoverStateText) Kind() WidgetKind {
	return WidgetKindText
}

type WidgetHoverStateImage struct {
	Height   int
	ImageURL string // a valid URL of a reddit-hosted image,
	Width    int
}

type widgetHoverStateImage struct {
	Height   int        `json:"height"`
	ImageURL string     `json:"imageUrl"` // a valid URL of a reddit-hosted image,
	Kind     WidgetKind `json:"kind"`     // Only 'image'
	Width    int        `json:"width"`
}

func (img *WidgetHoverStateImage) MarshalJSON() ([]byte, error) {
	temp := widgetHoverStateImage{
		Height:   img.Height,
		ImageURL: img.ImageURL,
		Kind:     WidgetKindImage,
		Width:    img.Width,
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (img *WidgetHoverStateImage) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindImage
	temp := new(widgetHoverStateImage)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	img.Height = temp.Height
	img.ImageURL = temp.ImageURL
	img.Width = temp.Width

	return nil
}

func (_ *WidgetHoverStateImage) Kind() WidgetKind {
	return WidgetKindImage
}

type WidgetButton interface {
	Widget
}

type WidgetTextButton struct {
	Color      string // a 6-digit rgb hex color, e.g. `#AABBCC`
	FillColor  string // a 6-digit rgb hex color, e.g. `#AABBCC`
	HoverState WidgetHoverState
	Text       [30]byte
	TextColor  string // a 6-digit rgb hex color, e.g. `#AABBCC`
	URL        string // a valid url
}

type widgetTextButton struct {
	Color      string           `json:"color"`     // a 6-digit rgb hex color, e.g. `#AABBCC`
	FillColor  string           `json:"fillColor"` // a 6-digit rgb hex color, e.g. `#AABBCC`
	HoverState WidgetHoverState `json:"hoverState"`
	Kind       WidgetKind       `json:"kind"` // only 'text'
	Text       [30]byte         `json:"text"`
	TextColor  string           `json:"textColor"` // a 6-digit rgb hex color, e.g. `#AABBCC`
	URL        string           `json:"url"`       // a valid url
}

func (txt *WidgetTextButton) MarshalJSON() ([]byte, error) {
	temp := widgetTextButton{
		Color:      txt.Color,
		FillColor:  txt.FillColor,
		HoverState: txt.HoverState,
		Kind:       WidgetKindText,
		Text:       txt.Text,
		TextColor:  txt.TextColor,
		URL:        txt.URL,
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (txt *WidgetTextButton) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindText
	temp := new(widgetTextButton)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	txt.Color = temp.Color
	txt.FillColor = temp.FillColor
	txt.HoverState = temp.HoverState
	txt.Text = temp.Text
	txt.TextColor = temp.TextColor
	txt.URL = temp.URL

	return nil
}

func (_ *WidgetTextButton) Kind() WidgetKind {
	return WidgetKindText
}

type WidgetImageButton struct {
	Height     int
	HoverState WidgetHoverState
	ImageURL   string // a valid URL of a reddit-hosted image
	LinkURL    string // a valid URL of a reddit-hosted image
	Text       [30]byte
	Width      int
}

type widgetImageButton struct {
	Height     int              `json:"height"`
	HoverState WidgetHoverState `json:"hoverState"`
	ImageURL   string           `json:"imageUrl"` // a valid URL of a reddit-hosted image
	Kind       WidgetKind       `json:"kind"`     // Only 'image'
	LinkURL    string           `json:"linkUrl"`  // a valid URL of a reddit-hosted image
	Text       [30]byte         `json:"text"`
	Width      int              `json:"width"`
}

func (img *WidgetImageButton) MarshalJSON() ([]byte, error) {
	temp := widgetImageButton{
		Height:     img.Height,
		HoverState: img.HoverState,
		ImageURL:   img.ImageURL,
		Kind:       WidgetKindImage,
		LinkURL:    img.LinkURL,
		Text:       img.Text,
		Width:      img.Width,
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (img *WidgetImageButton) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindImage
	temp := new(widgetImageButton)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	img.Height = temp.Height
	img.HoverState = temp.HoverState
	img.ImageURL = temp.ImageURL
	img.LinkURL = temp.LinkURL
	img.Text = temp.Text
	img.Width = temp.Width

	return nil
}

func (_ *WidgetImageButton) Kind() WidgetKind {
	return WidgetKindImage
}

type WidgetButtons struct {
	Buttons     []WidgetButton
	Description string // raw Markdown text
	ShortName   [30]byte
	Styles      WidgetStyles
}

type widgetButtons struct {
	Buttons     []WidgetButton `json:"buttons"`
	Description string         `json:"description"` // raw Markdown text
	Kind        WidgetKind     `json:"kind"`        // Only 'button'
	ShortName   [30]byte       `json:"shortName"`
	Styles      WidgetStyles   `json:"styles"`
}

func (button *WidgetButtons) MarshalJSON() ([]byte, error) {
	temp := widgetButtons{
		Buttons:     append([]WidgetButton{}, button.Buttons...),
		Description: button.Description,
		Kind:        WidgetKindButton,
		ShortName:   button.ShortName,
		Styles: WidgetStyles{
			BackgroundColor: button.Styles.BackgroundColor,
			HeaderColor:     button.Styles.HeaderColor,
		},
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (button *WidgetButtons) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindButton
	temp := new(widgetButtons)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	button.Buttons = make([]WidgetButton, len(temp.Buttons))
	for i := 0; i < len(temp.Buttons); i++ {
		button.Buttons[i] = temp.Buttons[i]
	}
	button.Description = temp.Description
	button.ShortName = temp.ShortName
	button.Styles.BackgroundColor = temp.Styles.BackgroundColor
	button.Styles.HeaderColor = temp.Styles.HeaderColor

	return nil
}

func (_ *WidgetButtons) Kind() WidgetKind {
	return WidgetKindButton
}

type WidgetIDCard struct {
	CurrentlyViewingText [30]byte
	ShortName            [30]byte
	Styles               WidgetStyles
	SubscribersText      [30]byte
}

type widgetIDCard struct {
	CurrentlyViewingText [30]byte     `json:"currentlyViewingText"`
	Kind                 WidgetKind   `json:"kind"` // Only 'id-card'
	ShortName            [30]byte     `json:"shortName"`
	Styles               WidgetStyles `json:"styles"`
	SubscribersText      [30]byte     `json:"subscribersText"`
}

func (id *WidgetIDCard) MarshalJSON() ([]byte, error) {
	temp := widgetIDCard{
		CurrentlyViewingText: id.CurrentlyViewingText,
		Kind:                 WidgetKindIDCard,
		ShortName:            id.ShortName,
		Styles: WidgetStyles{
			BackgroundColor: id.Styles.BackgroundColor,
			HeaderColor:     id.Styles.HeaderColor,
		},
		SubscribersText: id.SubscribersText,
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (id *WidgetIDCard) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindIDCard
	temp := new(widgetIDCard)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	return nil
}

func (_ *WidgetIDCard) Kind() WidgetKind {
	return WidgetKindIDCard
}

type WidgetCommunityList struct {
	Data      []string // list of subreddit names
	ShortName [30]byte
	Styles    WidgetStyles
}

type widgetCommunityList struct {
	Data      []string     `json:"data"` // list of subreddit names
	Kind      WidgetKind   `json:"kind"` // Only 'community-list'
	ShortName [30]byte     `json:"shortName"`
	Styles    WidgetStyles `json:"styles"`
}

func (com *WidgetCommunityList) MarshalJSON() ([]byte, error) {
	temp := widgetCommunityList{
		Data:      append([]string{}, com.Data...),
		Kind:      WidgetKindCommunityList,
		ShortName: com.ShortName,
		Styles: WidgetStyles{
			HeaderColor:     com.Styles.HeaderColor,
			BackgroundColor: com.Styles.BackgroundColor,
		},
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (com *WidgetCommunityList) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindCommunityList
	temp := new(widgetCommunityList)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	com.Data = nil
	com.Data = append(com.Data, temp.Data...)
	com.ShortName = temp.ShortName
	com.Styles.BackgroundColor = temp.Styles.BackgroundColor
	com.Styles.HeaderColor = temp.Styles.HeaderColor

	return nil
}

func (_ *WidgetCommunityList) Kind() WidgetKind {
	return WidgetKindCommunityList
}

type WidgetCustom struct {
	CSS       [100000]byte
	Height    int // an integer between 50 and 500
	ImageData []struct {
		Height int      `json:"height"`
		Name   [20]byte `json:"name"`
		URL    string   `json:"url"` // a valid URL of a reddit-hosted image
		Width  int      `json:"width"`
	} `json:"imageData"`
	ShortName [30]byte
	Styles    WidgetStyles
	Text      string // raw Markdown text
}

type widgetCustom struct {
	CSS       [100000]byte `json:"css"`
	Height    int          `json:"height"` // an integer between 50 and 500
	ImageData []struct {
		Height int      `json:"height"`
		Name   [20]byte `json:"name"`
		URL    string   `json:"url"` // a valid URL of a reddit-hosted image
		Width  int      `json:"width"`
	} `json:"imageData"`
	Kind      WidgetKind   `json:"kind"` // Only 'custom'
	ShortName [30]byte     `json:"shortName"`
	Styles    WidgetStyles `json:"styles"`
	Text      string       `json:"text"` // raw Markdown text
}

func (cus *WidgetCustom) MarshalJSON() ([]byte, error) {
	temp := widgetCustom{
		CSS:       cus.CSS,
		Height:    cus.Height,
		ImageData: nil,
		Kind:      WidgetKindCustom,
		ShortName: cus.ShortName,
		Styles: WidgetStyles{
			HeaderColor:     cus.Styles.HeaderColor,
			BackgroundColor: cus.Styles.BackgroundColor,
		},
		Text: cus.Text,
	}
	temp.ImageData = append(temp.ImageData, cus.ImageData...)

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (cus *WidgetCustom) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindCustom

	temp := new(widgetCustom)

	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	cus.CSS = temp.CSS
	cus.Height = temp.Height
	cus.ImageData = nil
	cus.ImageData = append(cus.ImageData, temp.ImageData...)
	cus.ShortName = temp.ShortName
	cus.Styles.HeaderColor = temp.Styles.HeaderColor
	cus.Styles.BackgroundColor = temp.Styles.BackgroundColor
	cus.Text = temp.Text

	return nil
}

func (_ *WidgetCustom) Kind() WidgetKind {
	return WidgetKindCustom
}

type WidgetDisplayType string

const (
	WidgetDisplayCloud WidgetDisplayType = "cloud"
	WidgetDisplayList  WidgetDisplayType = "list"
)

type WidgetPostFlair struct {
	Display   WidgetDisplayType
	Order     []string // list of flair template IDs
	ShortName [30]byte
	Styles    WidgetStyles
}

type widgetPostFlair struct {
	Display   WidgetDisplayType `json:"display"`
	Kind      WidgetKind        `json:"kind"`  // Only 'post-flair'
	Order     []string          `json:"order"` // list of flair template IDs
	ShortName [30]byte          `json:"shortName"`
	Styles    WidgetStyles      `json:"styles"`
}

func (flair *WidgetPostFlair) MarshalJSON() ([]byte, error) {
	temp := widgetPostFlair{
		Display:   flair.Display,
		Kind:      WidgetKindPostFlair,
		Order:     flair.Order,
		ShortName: flair.ShortName,
		Styles: WidgetStyles{
			HeaderColor:     flair.Styles.HeaderColor,
			BackgroundColor: flair.Styles.BackgroundColor,
		},
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (flair *WidgetPostFlair) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindPostFlair
	temp := new(widgetPostFlair)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	flair.Display = temp.Display
	flair.Order = nil
	flair.Order = append(flair.Order, temp.Order...)
	flair.ShortName = temp.ShortName
	flair.Styles.HeaderColor = temp.Styles.HeaderColor
	flair.Styles.BackgroundColor = temp.Styles.BackgroundColor

	return nil
}

func (_ *WidgetPostFlair) Kind() WidgetKind {
	return WidgetKindPostFlair
}

type WidgetModerators struct {
	Styles WidgetStyles
}

type widgetModerators struct {
	Kind   WidgetKind   `json:"kind"`
	Styles WidgetStyles `json:"styles"`
}

func (mod *WidgetModerators) MarshalJSON() ([]byte, error) {
	temp := widgetModerators{
		Kind: WidgetKindModerators,
		Styles: WidgetStyles{
			HeaderColor:     mod.Styles.HeaderColor,
			BackgroundColor: mod.Styles.BackgroundColor,
		},
	}

	data, err := json.Marshal(temp)
	if err != nil {
		err = &JSONError{Message: WidgetMarshallingErrorPrefix + err.Error(), Data: data}
	}

	return data, err
}

func (mod *WidgetModerators) UnmarshalJSON(data []byte) error {
	const TYPE = WidgetKindModerators

	temp := new(widgetModerators)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return &JSONError{
			Message: WidgetUnmarshallingErrorPrefix + err.Error(),
			Data:    data,
		}
	} else if temp.Kind != TYPE {
		return &JSONError{
			Message: WidgetUnmarshallingTypeErrorMessage + string(TYPE),
			Data:    data,
		}
	}

	mod.Styles.BackgroundColor = temp.Styles.BackgroundColor
	mod.Styles.HeaderColor = temp.Styles.HeaderColor

	return nil
}

func (_ *WidgetModerators) Kind() WidgetKind {
	return WidgetKindModerators
}

// PostSubredditWidget Add and return a widget to the specified subreddit
// Accepts a JSON payload representing the widget data to be saved.
// Valid payloads differ in shape based on the "kind" attribute passed on the root object, which must be a valid widget kind.
func (s *WidgetService) PostSubredditWidget(ctx context.Context, subreddit string, widget Widget) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/widget", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, widget)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// DeleteSubredditWidgetByID Delete a widget from the specified subreddit (if it exists)
func (s *WidgetService) DeleteSubredditWidgetByID(ctx context.Context, subreddit, widgetID string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/widget/%s", subreddit, widgetID)

	req, err := s.client.NewJSONRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PutSubredditWidgetByID Update and return the data of a widget.
// Accepts a JSON payload representing the widget data to be saved.
// Valid payloads differ in shape based on the "kind" attribute passed on the root object, which must be a valid widget kind.
func (s *WidgetService) PutSubredditWidgetByID(ctx context.Context, subreddit, widgetID string, widget Widget) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/widget/%s", subreddit, widgetID)

	req, err := s.client.NewJSONRequest(http.MethodPut, path, widget)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PostWidgetImageUploadS3 Acquire and return an upload lease to s3 temp bucket.
// The return value of this function is a json object containing credentials for uploading assets to S3 bucket, S3 url for upload request and the key to use for uploading.
// Using this lease the client will upload the emoji image to S3 temp bucket (included as part of the S3 URL).
// This lease is used by S3 to verify that the upload is authorized.
func (s *WidgetService) PostWidgetImageUploadS3(ctx context.Context, subreddit, filepath, mimetype string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/widget_image_upload_s3?filepath=%s&mimetype=%s", subreddit, filepath, mimetype)

	req, err := s.client.NewJSONRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

// PatchSubredditWidgetOrder Update the order of widget_ids in the specified subreddit
func (s *WidgetService) PatchSubredditWidgetOrder(ctx context.Context, subreddit string, widgetIDs ...string) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/widget_order/sidebar", subreddit)

	req, err := s.client.NewJSONRequest(http.MethodPatch, path, widgetIDs)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}

func (s *WidgetService) GetSubredditWidgets(ctx context.Context, subreddit string, progressiveImages bool) (*http.Response, error) {
	path := fmt.Sprintf("r/%s/api/widgets?progressive_images=%t", subreddit, progressiveImages)

	req, err := s.client.NewJSONRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, &InternalError{Message: err.Error()}
	}

	return s.client.Do(ctx, req, nil)
}
