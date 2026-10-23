package modhelper

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyshos.com/tyde"
	"github.com/stretchr/testify/assert"
)

type testSuggestion struct {
	title    string
	launched bool
}

func (s *testSuggestion) Icon() fyne.Resource { return nil }
func (s *testSuggestion) Title() string       { return s.title }
func (s *testSuggestion) Launch()             { s.launched = true }

func TestCommand_AnsweredBy(t *testing.T) {
	item := &testSuggestion{title: "Brightness 50%"}
	var asked string

	cmd := tyde.Command{Name: "brightness"}
	AnswerWith(&cmd, func(args string) tyde.LaunchSuggestion {
		asked = args
		return item
	})

	reply, err := cmd.Run("50")
	assert.NoError(t, err)
	assert.Equal(t, "50", asked)             // the words after the command name
	assert.Equal(t, "Brightness 50%", reply) // what the launcher would have shown
	assert.True(t, item.launched)
}

// A module that cannot act on the arguments says so, rather than reporting a
// success it did not have.
func TestCommand_AnsweredBy_Refuses(t *testing.T) {
	cmd := tyde.Command{Name: "brightness"}
	AnswerWith(&cmd, func(string) tyde.LaunchSuggestion { return nil })

	_, err := cmd.Run("banana")
	assert.EqualError(t, err, `cannot act on "banana"`)

	_, err = cmd.Run("")
	assert.EqualError(t, err, "needs something to act on")
}
