package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAndNormalizeCondition(t *testing.T) {
	t.Run("returns empty map for nil or empty string", func(t *testing.T) {
		result, err := parseAndNormalizeCondition(nil)
		assert.NoError(t, err)
		assert.Empty(t, result)

		result, err = parseAndNormalizeCondition("  ")
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("parses JSON string and replaces interval macros", func(t *testing.T) {
		input := `{ "$each_t": "$__interval", "nested": ["$__interval", {"inner": "$__interval"}] }`

		result, err := parseAndNormalizeCondition(input)
		assert.NoError(t, err)
		assert.Equal(t, "1s", result["$each_t"])
		nested := result["nested"].([]any)
		assert.Equal(t, "1s", nested[0])
		assert.Equal(t, "1s", nested[1].(map[string]any)["inner"])
	})

	t.Run("accepts maps and replaces macros deeply", func(t *testing.T) {
		result, err := parseAndNormalizeCondition(map[string]any{
			"$each_t": "$__interval",
			"levels":  []any{"$__interval", map[string]any{"inner": "$__interval"}},
		})

		assert.NoError(t, err)
		assert.Equal(t, "1s", result["$each_t"])
		levels := result["levels"].([]any)
		assert.Equal(t, "1s", levels[0])
		assert.Equal(t, "1s", levels[1].(map[string]any)["inner"])
	})

	t.Run("rejects invalid inputs", func(t *testing.T) {
		_, err := parseAndNormalizeCondition(123)
		assert.Error(t, err)

		_, err = parseAndNormalizeCondition("not-json")
		assert.Error(t, err)
	})
}
