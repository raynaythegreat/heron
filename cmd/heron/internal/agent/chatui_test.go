package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuggestionsForInput_ModelCommandIncludesMatchingModels(t *testing.T) {
	models := []modelChoice{
		{name: "gpt-5.4", modelID: "openai/gpt-5.4"},
		{name: "gpt-5.4-mini", modelID: "openai/gpt-5.4-mini"},
		{name: "claude-sonnet", modelID: "anthropic/claude-sonnet-4.6"},
	}

	got := suggestionsForInput("/model gpt-5.4", nil, models)
	require.Len(t, got, 2)
	assert.Equal(t, "/model gpt-5.4", got[0].cmd)
	assert.Equal(t, "/model gpt-5.4-mini", got[1].cmd)
}

func TestSuggestionsForInput_BaseCommandsStillWork(t *testing.T) {
	base := []cmdSuggestion{
		{cmd: "/help", desc: "show help"},
		{cmd: "/model", desc: "switch model"},
		{cmd: "/memory", desc: "search memory"},
	}

	got := suggestionsForInput("/me", base, nil)
	require.Len(t, got, 1)
	assert.Equal(t, "/memory", got[0].cmd)
}

func TestResolveModelChoice_ExactAndUniquePrefix(t *testing.T) {
	models := []modelChoice{
		{name: "gpt-5.4", modelID: "openai/gpt-5.4"},
		{name: "gpt-5.4-mini", modelID: "openai/gpt-5.4-mini"},
		{name: "claude-sonnet", modelID: "anthropic/claude-sonnet-4.6"},
	}

	exact, err := resolveModelChoice("gpt-5.4", models)
	require.NoError(t, err)
	assert.Equal(t, "gpt-5.4", exact.name)

	prefix, err := resolveModelChoice("claude", models)
	require.NoError(t, err)
	assert.Equal(t, "claude-sonnet", prefix.name)
}

func TestResolveModelChoice_AmbiguousPrefix(t *testing.T) {
	models := []modelChoice{
		{name: "gpt-5.4", modelID: "openai/gpt-5.4"},
		{name: "gpt-5.4-mini", modelID: "openai/gpt-5.4-mini"},
	}

	_, err := resolveModelChoice("gpt", models)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
	assert.Contains(t, err.Error(), "gpt-5.4")
	assert.Contains(t, err.Error(), "gpt-5.4-mini")
}
