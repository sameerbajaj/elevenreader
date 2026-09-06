package auth

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"
)

var jwtPattern = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]{10,}\.eyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}`)

// DetectBrowserTokens searches local browser IndexedDB storage for ElevenReader JWT tokens.
func DetectBrowserTokens() ([]string, error) {
	candidates := getPotentialIndexedDBPaths()
	var tokens []string
	seen := make(map[string]bool)

	for _, pattern := range candidates {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, dbDir := range matches {
			found, err := extractTokensFromDir(dbDir)
			if err != nil {
				continue
			}
			for _, t := range found {
				if !seen[t] {
					seen[t] = true
					tokens = append(tokens, t)
				}
			}
		}
	}

	return tokens, nil
}

func getPotentialIndexedDBPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var patterns []string
	switch runtime.GOOS {
	case "darwin":
		appSupport := filepath.Join(home, "Library", "Application Support")
		browsers := []string{
			filepath.Join("BraveSoftware", "Brave-Browser"),
			filepath.Join("Google", "Chrome"),
			"Chromium",
			filepath.Join("Arc", "User Data"),
			"Vivaldi",
			filepath.Join("Microsoft", "Edge"),
		}
		for _, b := range browsers {
			patterns = append(patterns, filepath.Join(appSupport, b, "*", "IndexedDB", "https_elevenreader.io_0.indexeddb.leveldb"))
		}
	case "linux":
		config := filepath.Join(home, ".config")
		browsers := []string{
			"google-chrome",
			"chromium",
			"BraveSoftware/Brave-Browser",
			"microsoft-edge",
		}
		for _, b := range browsers {
			patterns = append(patterns, filepath.Join(config, b, "*", "IndexedDB", "https_elevenreader.io_0.indexeddb.leveldb"))
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			patterns = append(patterns,
				filepath.Join(localAppData, "Google", "Chrome", "User Data", "*", "IndexedDB", "https_elevenreader.io_0.indexeddb.leveldb"),
				filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data", "*", "IndexedDB", "https_elevenreader.io_0.indexeddb.leveldb"),
			)
		}
	}
	return patterns
}

func extractTokensFromDir(dir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.ldb"))
	if err != nil {
		return nil, err
	}
	logs, _ := filepath.Glob(filepath.Join(dir, "*.log"))
	files = append(files, logs...)

	var results []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		matches := jwtPattern.FindAll(data, -1)
		for _, m := range matches {
			results = append(results, string(m))
		}
	}
	return results, nil
}

// VerifyToken tests whether a token is valid by making a lightweight API call.
func VerifyToken(baseURL, token string) bool {
	if baseURL == "" {
		baseURL = "https://api.elevenlabs.io/v1/reader"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/reads?page_size=1", nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "elevenreader-cli/1.0.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
