// errmsg attaches messages and status codes to errors.
//
// forged from https://github.com/carlmjohnson/resperr
package resperr

import (
	"errors"
	"fmt"
	"net/http"
)

type StatusTexter func(code int) string

var (
	DefaultErrorMessage string = "An unexpected error occured."

	// StatusCodeToMessage converts a status code to a user-facing message.
	// Used by UserMessageStatus to generate a message in case no message is specified.
	StatusCodeToMessage StatusTexter = http.StatusText

	// Default status code for a nil error.
	StatusCodeNoErr int = http.StatusOK

	// Default status code for a non-nil error.
	StatusCodeErr int = http.StatusInternalServerError

	// Default status code for a UserMessenger.
	StatusCodeMsg int = http.StatusBadRequest
)

// StatusCoder is an error with an associated status code.
type StatusCoder interface {
	error
	StatusCode() int
}

type statusCoder struct {
	error
	code int
}

func (sc statusCoder) Unwrap() error {
	return sc.error
}

func (sc statusCoder) Error() string {
	return sc.error.Error()
}

func (sc statusCoder) StatusCode() int {
	return sc.code
}

// WithStatusCode adds a StatusCoder to err's error chain.
// Unlike pkg/errors, WithStatusCode will wrap a nil error.
func WithStatusCode(err error, code int) error {
	if err == nil {
		err = errors.New(http.StatusText(code))
	}
	return statusCoder{err, code}
}

// StatusCode returns the status code associated with an error.
// If no status code is found, it returns a StatusCodeErr which is http.StatusInternalServerError (500) by default.
// If err is nil, it returns a StatusCodeNoErr, which is http.StatusOK (200) by default.
func StatusCode(err error) (code int) {
	if err == nil {
		return StatusCodeNoErr
	}
	if sc := StatusCoder(nil); errors.As(err, &sc) {
		return sc.StatusCode()
	}
	return StatusCodeErr
}

// UserMessenger is an error with an associated user-facing message.
type UserMessenger interface {
	error
	UserMessage() string
}

type messenger struct {
	error
	msg string
}

func (msgr messenger) Unwrap() error {
	return msgr.error
}

func (msgr messenger) UserMessage() string {
	return msgr.msg
}

// Returns the status code of the error.
// If a status code has not previously been set,
// the status code will be StatusCodeMsg, which is http.StatusBadRequest (400) by default.
func (msgr messenger) StatusCode() int {
	if sc := StatusCoder(nil); errors.As(msgr.error, &sc) {
		return sc.StatusCode()
	}

	return StatusCodeMsg
}

// WithUserMessage adds a UserMessenger to err's error chain.
// If a status code has not previously been set,
// the status code will be StatusCodeMsg, which is http.StatusBadRequest (400) by default.
// Unlike pkg/errors, WithUserMessage will wrap a nil error.
func WithUserMessage(err error, msg string) error {
	if err == nil {
		err = errors.New("UserMessage<" + msg + ">")
	}
	return messenger{err, msg}
}

// WithUserMessagef calls fmt.Sprintf before calling WithUserMessage.
func WithUserMessagef(err error, format string, v ...any) error {
	return WithUserMessage(err, fmt.Sprintf(format, v...))
}

// UserMessageStatus returns the user message associated with an error.
// If no message is found, it checks StatusCode and returns that message using StatusCodeToMessage,
// which uses http.StatusText by default.
func UserMessageStatus(err error) string {
	if err == nil {
		return ""
	}
	if um := UserMessenger(nil); errors.As(err, &um) {
		return um.UserMessage()
	}
	return StatusCodeToMessage(StatusCode(err))
}

// UserMessage returns the user message associated with an error.
// If no message is found, DefaultErrorMessage is returned.
// If err is nil, it returns "".
func UserMessage(err error) string {
	if err == nil {
		return ""
	}
	if um := UserMessenger(nil); errors.As(err, &um) {
		return um.UserMessage()
	}
	return DefaultErrorMessage
}

// WithCodeAndMessage is a convenience function for calling both
// WithStatusCode and WithUserMessage.
func WithCodeAndMessage(err error, code int, msg string) error {
	return WithStatusCode(WithUserMessage(err, msg), code)
}

// New is a convenience function for calling fmt.Errorf and WithStatusCode.
func New(code int, format string, v ...any) error {
	return WithStatusCode(
		fmt.Errorf(format, v...),
		code,
	)
}
