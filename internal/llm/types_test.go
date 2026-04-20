package llm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequest_ZeroValueValid(t *testing.T) {
	var r Request
	require.Empty(t, r.Model)
	require.Empty(t, r.Prompt)
	require.Empty(t, r.Data)
	require.Zero(t, r.MaxTokens)
}

func TestResult_FromCacheFalseByDefault(t *testing.T) {
	var r Result
	require.False(t, r.FromCache)
}
