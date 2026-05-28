package generator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"urlshort/internal/usecase/generator"
)

func TestGenerateShortLink(t *testing.T) {
	t.Parallel()

	t.Run("generates valid short link", func(t *testing.T) {
		t.Parallel()

		shortLink, err := generator.GenerateShortLink()
		require.NoError(t, err)

		assert.Len(t, shortLink, 10)

		for _, ch := range shortLink {
			assert.Contains(t, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_", string(ch))
		}
	})
}
