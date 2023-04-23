package resperr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestGetCode(t *testing.T) {
	base := WithStatusCode(errors.New(""), 5)
	wrapped := fmt.Errorf("wrapping: %w", base)

	testCases := map[string]struct {
		error
		int
	}{
		"nil":         {nil, 200},
		"default":     {errors.New(""), 500},
		"set":         {WithStatusCode(errors.New(""), 3), 3},
		"set-nil":     {WithStatusCode(nil, 4), 4},
		"wrapped":     {wrapped, 5},
		"set-message": {WithUserMessage(nil, "xxx"), 400},
		"set-both":    {WithCodeAndMessage(nil, 6, "xx"), 6},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			if StatusCode(tc.error) != tc.int {
				t.Errorf("got: %d, want: %d", StatusCode(tc.error), tc.int)
			}
		})
	}
}

func TestSetCode(t *testing.T) {
	t.Run("same-message", func(t *testing.T) {
		err := errors.New("hello")
		coder := WithStatusCode(err, 400)
		got := coder.Error()
		want := err.Error()

		if want != got {
			t.Errorf("got: %s, want: %s", got, want)
		}
	})
	t.Run("keep-chain", func(t *testing.T) {
		err := errors.New("hello")
		coder := WithStatusCode(err, 3)

		if !errors.Is(coder, err) {
			t.Errorf("'%v' does not match '%v'", coder, err)
		}
	})
	t.Run("set-nil", func(t *testing.T) {
		coder := WithStatusCode(nil, 400)

		if !strings.Contains(coder.Error(), http.StatusText(400)) {
			t.Errorf("'%s' does not contain '%s'", coder.Error(), http.StatusText(400))
		}
	})
	t.Run("override-default", func(t *testing.T) {
		err := context.DeadlineExceeded
		coder := WithStatusCode(err, 3)
		code := StatusCode(coder)

		if code != 3 {
			t.Errorf("got: %d, want: %d", code, 3)
		}
	})
}

func TestGetMsg(t *testing.T) {
	base := WithUserMessage(errors.New(""), "5")
	wrapped := fmt.Errorf("wrapping: %w", base)

	testCases := map[string]struct {
		error
		string
	}{
		"nil":     {nil, ""},
		"default": {errors.New(""), "Internal Server Error"},
		"set":     {WithUserMessage(errors.New(""), "3"), "3"},
		"set-nil": {WithUserMessage(nil, "4"), "4"},
		"wrapped": {wrapped, "5"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {

			if UserMessage(tc.error) != tc.string {
				t.Errorf("got: %s, want: %s", UserMessage(tc.error), tc.string)
			}
		})
	}
}

func TestSetMsg(t *testing.T) {
	t.Run("same-message", func(t *testing.T) {
		err := errors.New("hello")
		msgr := WithUserMessage(err, "a")

		if msgr.Error() != err.Error() {
			t.Errorf("got: %s, want: %s", msgr.Error(), err.Error())
		}
	})
	t.Run("keep-chain", func(t *testing.T) {
		err := errors.New("hello")
		msgr := WithUserMessage(err, "a")

		if !errors.Is(msgr, err) {
			t.Errorf("'%v' does not match '%v'", msgr, err)
		}
	})
	t.Run("set-nil", func(t *testing.T) {
		msgr := WithUserMessage(nil, "a")

		if msgr.Error() != "UserMessage<a>" {
			t.Errorf("got: %s, want: %s", msgr.Error(), "UserMessage<a>")
		}
	})
}

func TestMsgf(t *testing.T) {
	msg := "hello 1, 2, 3"
	err := WithUserMessagef(nil, "hello %d, %d, %d", 1, 2, 3)

	if UserMessage(err) != msg {
		t.Errorf("got: %s, want: %s", UserMessage(err), msg)
	}
}

func TestNew(t *testing.T) {
	t.Run("flat", func(t *testing.T) {
		err := New(404, "hello %s", "world")

		if UserMessage(err) != "Not Found" {
			t.Errorf("got: %s, want: %s", UserMessage(err), "Not Found")
		}

		if StatusCode(err) != 404 {
			t.Errorf("got: %d, want: %d", StatusCode(err), 404)
		}

		if err.Error() != "hello world" {
			t.Errorf("got: %s, want: %s", err.Error(), "hello world")
		}
	})

	t.Run("chain", func(t *testing.T) {
		const setMsg = "msg1"
		inner := WithUserMessage(nil, setMsg)
		err500 := WithUserMessage(WithStatusCode(errors.New("error"), 500), setMsg)

		w1 := New(5, "w1: %w", inner)
		w2 := New(6, "w2: %w", w1)

		if UserMessage(w2) != setMsg {
			t.Errorf("got: %s, want: %s", UserMessage(w2), setMsg)
		}

		if StatusCode(err500) != 500 {
			t.Errorf("got: %d, want: %d", StatusCode(err500), 500)
		}
		if StatusCode(w1) != 5 {
			t.Errorf("got: %d, want: %d", StatusCode(w1), 5)
		}

		if StatusCode(w2) != 6 {
			t.Errorf("got: %d, want: %d", StatusCode(w2), 6)
		}

		if w2.Error() != "w2: w1: UserMessage<msg1>" {
			t.Errorf("got: %s, want: %s", w2.Error(), "w2: w1: UserMessage<msg1>")
		}
	})
}
