package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommaDelimitedOptionalValue(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		result := CommaDelimitedOptionalValue([]string{})
		assert.Nil(t, result)
	})

	t.Run("single input", func(t *testing.T) {
		result := CommaDelimitedOptionalValue([]string{"foo"})
		assert.Equal(t, "foo", *result)
	})

	t.Run("multiple inputs", func(t *testing.T) {
		result := CommaDelimitedOptionalValue([]string{"foo", "bar", "baz"})
		assert.Equal(t, "foo, bar, baz", *result)
	})
}
