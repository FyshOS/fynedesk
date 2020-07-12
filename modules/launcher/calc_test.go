package launcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalc_IsExpression(t *testing.T) {
	c := newCalcSuggest().(*calc)

	assert.True(t, c.isExpression("5*3"))
	assert.True(t, c.isExpression("5*(3+2)"))
	assert.True(t, c.isExpression("5/5-1"))

	assert.False(t, c.isExpression("33"))
	assert.False(t, c.isExpression("5e"))
	assert.False(t, c.isExpression("2,1*4"))
	assert.False(t, c.isExpression("xterm"))
}
