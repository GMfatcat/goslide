package pdfexport

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// FindChrome locates a Chrome/Edge/Chromium binary. Search order:
//  1. GOSLIDE_CHROME_PATH env var (user override)
//  2. PATH (via exec.LookPath) for common names
//  3. Platform-specific known install locations
func FindChrome() (string, error) {
	return defaultFinder.find()
}

var defaultFinder = &chromeFinder{
	fileExists: func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	},
	lookPath: exec.LookPath,
}

type chromeFinder struct {
	fileExists func(path string) bool
	lookPath   func(name string) (string, error)
}

func (f *chromeFinder) find() (string, error) {
	// 1. Env override
	if env := os.Getenv("GOSLIDE_CHROME_PATH"); env != "" {
		if f.fileExists(env) {
			return env, nil
		}
		return "", fmt.Errorf("GOSLIDE_CHROME_PATH=%s but the file does not exist", env)
	}

	// 2. PATH search
	for _, name := range []string{"chrome", "chromium", "chromium-browser", "google-chrome", "microsoft-edge"} {
		if path, err := f.lookPath(name); err == nil {
			return path, nil
		}
	}

	// 3. Platform paths
	candidates := platformPaths()
	for _, p := range candidates {
		if f.fileExists(p) {
			return p, nil
		}
	}

	return "", notFoundError(candidates)
}

func platformPaths() []string {
	switch runtime.GOOS {
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		paths := []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		}
		if local != "" {
			paths = append(paths, local+`\Google\Chrome\Application\chrome.exe`)
		}
		return paths
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	default:
		return nil
	}
}

func notFoundError(checked []string) error {
	var sb strings.Builder
	sb.WriteString("Chrome/Edge/Chromium not found.\n\n")
	sb.WriteString("Checked PATH for: chrome, chromium, chromium-browser, google-chrome, microsoft-edge\n")
	if len(checked) > 0 {
		sb.WriteString("Checked known install paths:\n")
		for _, p := range checked {
			sb.WriteString("  - ")
			sb.WriteString(p)
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\nInstall Chrome/Edge/Chromium, or set GOSLIDE_CHROME_PATH to an explicit binary path.")
	return errors.New(sb.String())
}
