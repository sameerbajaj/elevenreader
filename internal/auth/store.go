package auth

import (
	"errors"
	"os"
	"strings"

	"github.com/sameerbajaj/elevenreader/internal/config"
)

type Credentials struct {
	Token   string
	APIKey  string
	Source  string
	BaseURL string
}

// ResolveCredentials determines the active token or API key from flags, env, config, or browser.
func ResolveCredentials(flagToken, flagAPIKey, flagBaseURL string) (*Credentials, error) {
	cfg, _ := config.Load()

	baseURL := flagBaseURL
	if baseURL == "" {
		if envBase := os.Getenv("ELEVENREADER_BASE_URL"); envBase != "" {
			baseURL = envBase
		} else if cfg != nil && cfg.BaseURL != "" {
			baseURL = cfg.BaseURL
		} else {
			baseURL = config.DefaultBaseURL
		}
	}

	// 1. Explicit CLI flag
	if flagToken != "" {
		return &Credentials{Token: flagToken, BaseURL: baseURL, Source: "flag (--token)"}, nil
	}
	if flagAPIKey != "" {
		return &Credentials{APIKey: flagAPIKey, BaseURL: baseURL, Source: "flag (--api-key)"}, nil
	}

	// 2. Environment variables
	if envToken := os.Getenv("ELEVENREADER_TOKEN"); envToken != "" {
		return &Credentials{Token: envToken, BaseURL: baseURL, Source: "env (ELEVENREADER_TOKEN)"}, nil
	}
	if envKey := os.Getenv("ELEVENLABS_API_KEY"); envKey != "" {
		return &Credentials{APIKey: envKey, BaseURL: baseURL, Source: "env (ELEVENLABS_API_KEY)"}, nil
	}
	if envKey := os.Getenv("ELEVEN_API_KEY"); envKey != "" {
		return &Credentials{APIKey: envKey, BaseURL: baseURL, Source: "env (ELEVEN_API_KEY)"}, nil
	}

	// 3. Saved config file
	if cfg != nil {
		if cfg.Token != "" {
			return &Credentials{Token: cfg.Token, BaseURL: baseURL, Source: "config file (~/.config/elevenreader/config.json)"}, nil
		}
		if cfg.APIKey != "" {
			return &Credentials{APIKey: cfg.APIKey, BaseURL: baseURL, Source: "config file (~/.config/elevenreader/config.json)"}, nil
		}
	}

	// 4. Fallback: try detecting browser session
	tokens, _ := DetectBrowserTokens()
	for _, t := range tokens {
		if VerifyToken(baseURL, t) {
			return &Credentials{Token: t, BaseURL: baseURL, Source: "local browser session (auto-detected)"}, nil
		}
	}

	return &Credentials{BaseURL: baseURL, Source: "unauthenticated"}, errors.New("no credentials found. Run 'elevenreader auth login', 'elevenreader auth import', or set ELEVENREADER_TOKEN")
}

// MaskToken obscures all but the first 4 and last 4 characters.
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "********"
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}
