package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

// PrintJSON outputs the value as pretty indented JSON.
func PrintJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// FormatNumber formats an integer with thousands commas (e.g. 1,234,567).
func FormatNumber(n int) string {
	in := strconv.Itoa(n)
	out := make([]byte, len(in)+(len(in)-1)/3)
	for i, j, k := len(in)-1, len(out)-1, 0; i >= 0; i, j, k = i-1, j-1, k+1 {
		if k > 0 && k%3 == 0 {
			out[j] = ','
			j--
		}
		out[j] = in[i]
	}
	return string(out)
}

// FormatTimeRelative converts a unix timestamp to human readable relative time.
func FormatTimeRelative(unixSec int64) string {
	if unixSec <= 0 {
		return "-"
	}
	t := time.Unix(unixSec, 0)
	diff := time.Since(t)

	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", mins)
	}
	if diff < 24*time.Hour {
		hrs := int(diff.Hours())
		if hrs == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hrs)
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "1d ago"
	}
	if days < 30 {
		return fmt.Sprintf("%dd ago", days)
	}
	return t.Format("Jan 02, 2006")
}

// Truncate ensures text does not exceed maxLen, appending ellipsis if needed.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// ErrorPrint prints an error message to stderr.
func ErrorPrint(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", a...)
}

// SuccessPrint prints a success message.
func SuccessPrint(format string, a ...interface{}) {
	fmt.Printf("✓ "+format+"\n", a...)
}

var spanTagRegex = regexp.MustCompile(`<span[^>]*>(.*?)</span>`)

// CleanMarkdown strips timing span tags returned by the ElevenReader markdown endpoint.
func CleanMarkdown(s string) string {
	return spanTagRegex.ReplaceAllString(s, "$1")
}
