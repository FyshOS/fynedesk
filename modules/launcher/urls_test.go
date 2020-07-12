package launcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrls_IsURL(t *testing.T) {
	u := newURLs().(*urls)

	assert.True(t, u.isURL("https://fyne.io"))
	assert.True(t, u.isURL("https://google.com"))
	assert.True(t, u.isURL("http://bbc.co.uk"))

	assert.False(t, u.isURL("https://bit.l"))
	assert.False(t, u.isURL("ftp://server.com/file.txt"))
	assert.False(t, u.isURL("file:///home/dir/file.txt"))
}
