package generate

import (
	"testing"
	"time"

	"github.com/GMfatcat/goslide/internal/config"
	"github.com/stretchr/testify/require"
)

func TestResolve_UsesYAMLDefaults(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{
		BaseURL:   "https://api.openai.com/v1",
		Model:     "gpt-4o",
		APIKeyEnv: "OPENAI_API_KEY",
		Timeout:   60 * time.Second,
	}}
	t.Setenv("OPENAI_API_KEY", "sk-test")

	opts, err := Resolve(cfg, Flags{}, Input{Topic: "t"})
	require.NoError(t, err)
	require.Equal(t, "https://api.openai.com/v1", opts.BaseURL)
	require.Equal(t, "gpt-4o", opts.Model)
	require.Equal(t, "sk-test", opts.APIKey)
	require.Equal(t, 60*time.Second, opts.Timeout)
}

func TestResolve_FlagsOverrideYAML(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{
		BaseURL:   "https://api.openai.com/v1",
		Model:     "gpt-4o",
		APIKeyEnv: "OPENAI_API_KEY",
	}}
	t.Setenv("OPENAI_API_KEY", "sk-test")

	opts, err := Resolve(cfg, Flags{
		BaseURL: "http://localhost:11434/v1",
		Model:   "llama3",
	}, Input{Topic: "t"})
	require.NoError(t, err)
	require.Equal(t, "http://localhost:11434/v1", opts.BaseURL)
	require.Equal(t, "llama3", opts.Model)
}

func TestResolve_APIKeyEnvOverride(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{
		BaseURL:   "x",
		Model:     "y",
		APIKeyEnv: "OPENAI_API_KEY",
	}}
	t.Setenv("CUSTOM_KEY", "sk-custom")

	opts, err := Resolve(cfg, Flags{APIKeyEnv: "CUSTOM_KEY"}, Input{Topic: "t"})
	require.NoError(t, err)
	require.Equal(t, "sk-custom", opts.APIKey)
}

func TestResolve_MissingAPIKey(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{
		BaseURL: "x", Model: "y", APIKeyEnv: "NOPE_NOT_SET_12345",
	}}
	_, err := Resolve(cfg, Flags{}, Input{Topic: "t"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "NOPE_NOT_SET_12345")
}

func TestResolve_MissingBaseURL(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{Model: "y", APIKeyEnv: "OPENAI_API_KEY"}}
	t.Setenv("OPENAI_API_KEY", "k")
	_, err := Resolve(cfg, Flags{}, Input{Topic: "t"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "base_url")
}

func TestResolve_MissingModel(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{BaseURL: "x", APIKeyEnv: "OPENAI_API_KEY"}}
	t.Setenv("OPENAI_API_KEY", "k")
	_, err := Resolve(cfg, Flags{}, Input{Topic: "t"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "model")
}

func TestResolve_DefaultTimeout(t *testing.T) {
	cfg := &config.Config{Generate: config.GenerateConfig{BaseURL: "x", Model: "y", APIKeyEnv: "OPENAI_API_KEY"}}
	t.Setenv("OPENAI_API_KEY", "k")
	opts, err := Resolve(cfg, Flags{}, Input{Topic: "t"})
	require.NoError(t, err)
	require.Equal(t, 120*time.Second, opts.Timeout)
}
