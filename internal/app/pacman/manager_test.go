package pacman

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testManager() *Manager {
	return &Manager{provider: &dummyProvider{}}
}

func TestManager_HasUpdates(t *testing.T) {
	m := testManager()
	assert.True(t, m.HasUpdates())

	m.updates = []*Update{}
	assert.False(t, m.HasUpdates())
}

func TestManager_Search(t *testing.T) {
	m := testManager()
	assert.Equal(t, 1, len(m.Search("age 1")))
	assert.Equal(t, 0, len(m.Search("Missing")))
}
