package generate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystemPrompt_NotEmpty(t *testing.T) {
	p := SystemPrompt()
	require.NotEmpty(t, p)
}

func TestSystemPrompt_ContainsCoreKeywords(t *testing.T) {
	p := SystemPrompt()
	for _, kw := range []string{"layout:", "theme:", "two-column", "dashboard", "---"} {
		require.Truef(t, strings.Contains(p, kw), "system prompt missing keyword %q", kw)
	}
}
