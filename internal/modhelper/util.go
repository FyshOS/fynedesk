package modhelper

import (
	"errors"
	"fmt"
	"strings"

	"fyshos.com/tyde"
)

// AnswerWith states that this command does whatever the module's own launcher
// entry does, and returns the command so it can be written in place:
//
//	tyde.Command{Name: "brightness", ...}.AnsweredBy(b.item)
func AnswerWith(c *tyde.Command, suggest func(args string) tyde.LaunchSuggestion) {
	c.Run = func(args string) (string, error) {
		item := suggest(args)
		if item == nil {
			if strings.TrimSpace(args) == "" {
				return "", errors.New("needs something to act on")
			}
			return "", fmt.Errorf("cannot act on %q", args)
		}

		item.Launch()
		return item.Title(), nil
	}
}
