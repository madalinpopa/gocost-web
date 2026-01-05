package income

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSourceVO(t *testing.T) {
	t.Run("empty source", func(t *testing.T) {
		source, err := NewSourceVO("")
		assert.NoError(t, err)
		assert.Equal(t, "", source.Value())
	})

	t.Run("source too long", func(t *testing.T) {
		longSource := strings.Repeat("a", maxSourceLength+1)
		_, err := NewSourceVO(longSource)
		assert.ErrorIs(t, err, ErrSourceTooLong)
	})

	t.Run("valid source", func(t *testing.T) {
		validSource := "Salary"
		source, err := NewSourceVO(validSource)
		assert.NoError(t, err)
		assert.Equal(t, validSource, source.Value())
	})
}

func TestSourceVO_Equals(t *testing.T) {
	t.Run("equal sources", func(t *testing.T) {
		s1, _ := NewSourceVO("Salary")
		s2, _ := NewSourceVO("Salary")
		assert.True(t, s1.Equals(s2))
	})

	t.Run("unequal sources", func(t *testing.T) {
		s1, _ := NewSourceVO("Salary")
		s2, _ := NewSourceVO("Bonus")
		assert.False(t, s1.Equals(s2))
	})
}
