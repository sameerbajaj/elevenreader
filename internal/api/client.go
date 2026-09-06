package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(baseURL, token, apiKey string) *Client {
	if baseURL == "" {
		baseURL = "https://api.elevenlabs.io/v1/reader"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		BaseURL: baseURL,
		Token:   token,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	relPath := strings.TrimLeft(path, "/")
	fullURL := fmt.Sprintf("%s/%s", c.BaseURL, relPath)

	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "elevenreader-cli/1.0.0")

	// Apply authentication
	if c.APIKey != "" {
		req.Header.Set("xi-api-key", c.APIKey)
	}
	if c.Token != "" {
		// If token starts with sk_ it's likely an API key
		if strings.HasPrefix(c.Token, "sk_") {
			req.Header.Set("xi-api-key", c.Token)
		} else {
			req.Header.Set("Authorization", "Bearer "+c.Token)
		}
	}

	return req, nil
}

func (c *Client) doRequest(req *http.Request, v interface{}) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			RawBody:    string(bodyBytes),
		}
		var errEnvelope struct {
			Detail APIErrorDetail `json:"detail"`
		}
		if err := json.Unmarshal(bodyBytes, &errEnvelope); err == nil && errEnvelope.Detail.Message != "" {
			apiErr.Detail = errEnvelope.Detail
		}
		return apiErr
	}

	if v != nil {
		if err := json.Unmarshal(bodyBytes, v); err != nil {
			return fmt.Errorf("decoding JSON response: %w", err)
		}
	}

	return nil
}

// ListReadsOptions contains query filters for listing reads.
type ListReadsOptions struct {
	PageSize   int
	SortBy     string
	LastSortID string
}

// ListReads retrieves reads in user's library.
func (c *Client) ListReads(opts ListReadsOptions) (*ReadsListResponse, error) {
	u := "/reads"
	q := url.Values{}
	if opts.PageSize > 0 {
		q.Set("page_size", strconv.Itoa(opts.PageSize))
	}
	if opts.SortBy != "" {
		q.Set("sort_by", opts.SortBy)
	}
	if opts.LastSortID != "" {
		q.Set("last_sort_id", opts.LastSortID)
	}
	if len(q) > 0 {
		u = u + "?" + q.Encode()
	}

	req, err := c.newRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var res ReadsListResponse
	if err := c.doRequest(req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// GetRead returns a single read by ID.
func (c *Client) GetRead(readID string) (*Read, error) {
	path := fmt.Sprintf("/reads/%s", url.PathEscape(readID))
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var r Read
	if err := c.doRequest(req, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// GetReadMarkdown returns the markdown text for a read.
func (c *Client) GetReadMarkdown(readID string) (string, error) {
	path := fmt.Sprintf("/reads/%s/markdown", url.PathEscape(readID))
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return "", &APIError{StatusCode: resp.StatusCode, RawBody: string(b)}
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetReadSource returns the source text for a read.
func (c *Client) GetReadSource(readID string) (string, error) {
	path := fmt.Sprintf("/reads/%s/source", url.PathEscape(readID))
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return "", &APIError{StatusCode: resp.StatusCode, RawBody: string(b)}
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// AddReadOptions holds arguments when adding a read to library.
type AddReadOptions struct {
	SourceURL   string
	FilePath    string
	Text        string
	Title       string
	Author      string
	Description string
}

// AddRead imports a website URL, document file, or raw text into the ElevenReader library.
func (c *Client) AddRead(opts AddReadOptions) (*Read, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if opts.SourceURL != "" {
		_ = writer.WriteField("source", "website")
		_ = writer.WriteField("source_url", opts.SourceURL)
	} else if opts.FilePath != "" {
		file, err := os.Open(opts.FilePath)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		defer file.Close()

		_ = writer.WriteField("source", "file")
		part, err := writer.CreateFormFile("from_document", filepath.Base(opts.FilePath))
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(part, file); err != nil {
			return nil, err
		}
	} else if opts.Text != "" {
		_ = writer.WriteField("source", "file")
		filename := "note.txt"
		if opts.Title != "" {
			filename = strings.ReplaceAll(opts.Title, " ", "_") + ".txt"
		}
		part, err := writer.CreateFormFile("from_document", filename)
		if err != nil {
			return nil, err
		}
		if _, err := io.WriteString(part, opts.Text); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("must provide one of: URL, file path, or text")
	}

	if opts.Author != "" {
		_ = writer.WriteField("author", opts.Author)
	}
	if opts.Description != "" {
		_ = writer.WriteField("description", opts.Description)
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := c.newRequest(http.MethodPost, "/reads/add/v2", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var created Read
	if err := c.doRequest(req, &created); err != nil {
		return nil, err
	}

	// If a custom title was provided, patch it immediately
	if opts.Title != "" && created.ReadID != "" {
		_ = c.UpdateRead(created.ReadID, ReadUpdatePayload{
			Title: &opts.Title,
		})
		created.Title = opts.Title
	}

	return &created, nil
}

// UpdateRead modifies read metadata (title, author, description, unread status).
func (c *Client) UpdateRead(readID string, payload ReadUpdatePayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/reads/%s", url.PathEscape(readID))
	req, err := c.newRequest(http.MethodPatch, path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, nil)
}

// ArchiveRead moves a read into archive.
func (c *Client) ArchiveRead(readID string) error {
	path := fmt.Sprintf("/reads/%s/archive", url.PathEscape(readID))
	req, err := c.newRequest(http.MethodPost, path, bytes.NewReader([]byte("{}")))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, nil)
}

// UnarchiveRead restores an archived read to active library.
func (c *Client) UnarchiveRead(readID string) error {
	path := fmt.Sprintf("/reads/%s/unarchive", url.PathEscape(readID))
	req, err := c.newRequest(http.MethodPost, path, bytes.NewReader([]byte("{}")))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, nil)
}

// DeleteRead permanently removes a read from library.
func (c *Client) DeleteRead(readID string) error {
	path := fmt.Sprintf("/reads/%s", url.PathEscape(readID))
	req, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}

// ListCollections returns all user collections.
func (c *Client) ListCollections() ([]UserCollection, error) {
	req, err := c.newRequest(http.MethodGet, "/collections", nil)
	if err != nil {
		return nil, err
	}

	var res CollectionsResponse
	if err := c.doRequest(req, &res); err != nil {
		return nil, err
	}
	return res.Collections, nil
}

// CreateCollection creates a new user collection folder.
func (c *Client) CreateCollection(title, icon string) (*UserCollection, error) {
	payload := map[string]string{
		"title": title,
		"icon":  icon,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(http.MethodPost, "/user-collections", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var created UserCollection
	if err := c.doRequest(req, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// DeleteCollection deletes a collection.
func (c *Client) DeleteCollection(collectionID string) error {
	path := fmt.Sprintf("/user-collections/%s", url.PathEscape(collectionID))
	req, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}

// AddReadToCollection adds a read to a user collection.
func (c *Client) AddReadToCollection(collectionID, readID string) error {
	path := fmt.Sprintf("/user-collections/%s/reads/%s", url.PathEscape(collectionID), url.PathEscape(readID))
	req, err := c.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}

// RemoveReadFromCollection removes a read from a user collection.
func (c *Client) RemoveReadFromCollection(collectionID, readID string) error {
	path := fmt.Sprintf("/user-collections/%s/reads/%s", url.PathEscape(collectionID), url.PathEscape(readID))
	req, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}

// ListBookmarks returns all bookmarks for a specific read.
func (c *Client) ListBookmarks(readID string) ([]Bookmark, error) {
	path := fmt.Sprintf("/bookmarks/%s", url.PathEscape(readID))
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var bookmarks []Bookmark
	if err := c.doRequest(req, &bookmarks); err != nil {
		return nil, err
	}
	return bookmarks, nil
}

// ExportBookmarksMarkdown downloads all bookmarks for a read formatted as Markdown.
func (c *Client) ExportBookmarksMarkdown(readID string) (string, error) {
	path := fmt.Sprintf("/bookmarks/%s/markdown", url.PathEscape(readID))
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return "", &APIError{StatusCode: resp.StatusCode, RawBody: string(b)}
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DeleteBookmark deletes a single bookmark.
func (c *Client) DeleteBookmark(bookmarkID string) error {
	path := fmt.Sprintf("/bookmarks/%s", url.PathEscape(bookmarkID))
	req, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, nil)
}
