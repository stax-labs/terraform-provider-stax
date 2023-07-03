package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubtract(t *testing.T) {
	assert := require.New(t)
	left := map[string]bool{
		"a": true,
		"b": true,
		"c": true,
	}

	right := map[string]bool{
		"a": true,
		"c": true,
	}

	expected := []string{"b"}
	actual := Subtract(left, right)

	assert.Equal(expected, actual)
}
