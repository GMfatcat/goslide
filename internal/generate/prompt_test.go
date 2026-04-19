package generate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildUserMessage_SimpleMode(t *testing.T) {
	in := Input{Topic: "Introduction to Kubernetes"}
	msg, err := BuildUserMessage(in)
	require.NoError(t, err)
	require.Contains(t, msg, "Introduction to Kubernetes")
	require.Contains(t, strings.ToLower(msg), "topic")
}

func TestBuildUserMessage_EmptyInputFails(t *testing.T) {
	_, err := BuildUserMessage(Input{})
	require.Error(t, err)
}
