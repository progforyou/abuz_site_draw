package telesend

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindAllTags(t *testing.T) {
	var test = "<p><b>test <i>find</i></b> <div name=\"test\">hello</div></p>"
	test = removeUnsupportedTags(test)
	assert.Equal(t, test, "<b>test <i>find</i></b> hello\n")
}
