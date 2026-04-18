package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_NoConfigFile(t *testing.T) {
	cfg, err := Load(t.TempDir())
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Empty(t, cfg.API.Proxy)
}

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	content := `
api:
  proxy:
    /api/test:
      target: http://localhost:9999
      headers:
        X-Custom: "hello"
`
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte(content), 0644)

	cfg, err := Load(dir)
	require.NoError(t, err)
	require.Len(t, cfg.API.Proxy, 1)

	proxy := cfg.API.Proxy["/api/test"]
	require.Equal(t, "http://localhost:9999", proxy.Target)
	require.Equal(t, "hello", proxy.Headers["X-Custom"])
}

func TestLoad_EnvVarExpansion(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("GOSLIDE_TEST_TOKEN", "secret123")
	defer os.Unsetenv("GOSLIDE_TEST_TOKEN")

	content := `
api:
  proxy:
    /api/secure:
      target: http://localhost:8000
      headers:
        Authorization: "Bearer ${GOSLIDE_TEST_TOKEN}"
`
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte(content), 0644)

	cfg, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, "Bearer secret123", cfg.API.Proxy["/api/secure"].Headers["Authorization"])
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte("key: [unclosed bracket\n"), 0644)

	_, err := Load(dir)
	require.Error(t, err)
}

func TestLoad_InvalidTargetURL(t *testing.T) {
	dir := t.TempDir()
	content := `
api:
  proxy:
    /api/bad:
      target: "://not-a-url"
`
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte(content), 0644)

	_, err := Load(dir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid target URL")
}

func TestLoad_MultipleProxies(t *testing.T) {
	dir := t.TempDir()
	content := `
api:
  proxy:
    /api/aoi:
      target: http://localhost:8000
    /api/ollama:
      target: http://localhost:11434
`
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte(content), 0644)

	cfg, err := Load(dir)
	require.NoError(t, err)
	require.Len(t, cfg.API.Proxy, 2)
}

func TestLoad_ThemeOverrides(t *testing.T) {
	dir := t.TempDir()
	content := `
theme:
  overrides:
    slide-bg: "#1e1e2e"
    slide-accent: "#f38ba8"
`
	os.WriteFile(filepath.Join(dir, "goslide.yaml"), []byte(content), 0644)

	cfg, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, "#1e1e2e", cfg.Theme.Overrides["slide-bg"])
	require.Equal(t, "#f38ba8", cfg.Theme.Overrides["slide-accent"])
}
