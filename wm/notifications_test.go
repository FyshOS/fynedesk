package wm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendNotification(t *testing.T) {
	var got *Notification
	SetNotificationListener(func(n *Notification) {
		got = n
	})

	message := &Notification{Title: "Test", Body: "Message"}
	SendNotification(message)

	assert.NotNil(t, got)
	assert.Equal(t, "Test", got.Title)
	assert.Equal(t, "Message", got.Body)
}

func TestNewNotification(t *testing.T) {
	n1 := NewNotification("test", "body")
	n2 := NewNotification("test", "body")

	assert.NotZero(t, n1.ID)
	assert.NotEqual(t, n1.ID, n2.ID)
}
