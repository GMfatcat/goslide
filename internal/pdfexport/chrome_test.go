package pdfexport

import (
	"errors"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindChrome_EnvOverrideWins(t *testing.T) {
	t.Setenv("GOSLIDE_CHROME_PATH", "/fake/chrome")
	finder := &chromeFinder{
		fileExists: func(path string) bool { return path == "/fake/chrome" },
		lookPath:   func(string) (string, error) { return "", errors.New("not found") },
	}
	got, err := finder.find()
	require.NoError(t, err)
	require.Equal(t, "/fake/chrome", got)
}

func TestFindChrome_EnvOverrideMissingErrors(t *testing.T) {
	t.Setenv("GOSLIDE_CHROME_PATH", "/nonexistent/chrome")
	finder := &chromeFinder{
		fileExists: func(string) bool { return false },
		lookPath:   func(string) (string, error) { return "", errors.New("not found") },
	}
	_, err := finder.find()
	require.Error(t, err)
	require.Contains(t, err.Error(), "GOSLIDE_CHROME_PATH")
}

func TestFindChrome_PATHSearch(t *testing.T) {
	t.Setenv("GOSLIDE_CHROME_PATH", "")
	finder := &chromeFinder{
		fileExists: func(string) bool { return false },
		lookPath: func(name string) (string, error) {
			if name == "chromium" {
				return "/usr/bin/chromium", nil
			}
			return "", errors.New("not found")
		},
	}
	got, err := finder.find()
	require.NoError(t, err)
	require.Equal(t, "/usr/bin/chromium", got)
}

func TestFindChrome_PlatformPaths(t *testing.T) {
	t.Setenv("GOSLIDE_CHROME_PATH", "")
	var probed []string
	finder := &chromeFinder{
		fileExists: func(path string) bool {
			probed = append(probed, path)
			return false
		},
		lookPath: func(string) (string, error) { return "", errors.New("not found") },
	}
	_, err := finder.find()
	require.Error(t, err)
	require.Contains(t, err.Error(), "Chrome/Edge/Chromium")
	require.NotEmpty(t, probed)
	for _, p := range probed {
		require.NotEmpty(t, p)
	}
}

func TestFindChrome_PlatformPathMatch(t *testing.T) {
	t.Setenv("GOSLIDE_CHROME_PATH", "")
	var target string
	switch runtime.GOOS {
	case "windows":
		target = `C:\Program Files\Google\Chrome\Application\chrome.exe`
	case "darwin":
		target = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	default:
		t.Skip("platform paths test only meaningful on windows/darwin; linux uses PATH")
	}
	finder := &chromeFinder{
		fileExists: func(path string) bool { return path == target },
		lookPath:   func(string) (string, error) { return "", errors.New("not found") },
	}
	got, err := finder.find()
	require.NoError(t, err)
	require.Equal(t, target, got)
}
