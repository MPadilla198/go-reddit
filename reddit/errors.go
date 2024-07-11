package reddit

import (
	"fmt"
	"net/http"
)

type InternalError struct {
	// Error message
	Message string `json:"message"`
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("InternalError: %s", e.Message)
}

type JSONError struct {
	// Error message
	Message string `json:"message"`
	// JSON data that caused this error
	Data []byte `json:"data"`
}

func (r *JSONError) Error() string {
	return fmt.Sprintf("JSONError: %s\n%s", r.Message, r.Data)
}

// An ResponseError reports the error caused by an API request
type ResponseError struct {
	// Error message
	Message string `json:"message"`
	// HTTP response that caused this error
	Response *http.Response `json:"response"`
}

func (r *ResponseError) Error() string {
	if r.Response != nil {
		return fmt.Sprintf(
			"ResponseError: %s %s (STATUS: %d) %s",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message,
		)
	}
	return fmt.Sprintf("ResponseError: %s", r.Message)
}

// RateLimitError occurs when the client is sending too many requests to Reddit in a given time frame.
type RateLimitError struct {
	ResponseError
	// Rate specifies the last known rate limit for the client
	Rate Rate `json:"rate"`
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("RateLimitError: %s\nRATE: %v", e.ResponseError.Error(), e.Rate)
}
