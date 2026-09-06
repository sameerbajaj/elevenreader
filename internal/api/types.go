package api

import "fmt"

// Read represents a document, book, or article in the ElevenReader library.
type Read struct {
	ReadID           string    `json:"read_id"`
	CanonicalReadID  string    `json:"canonical_read_id,omitempty"`
	Title            string    `json:"title"`
	Subtitle         *string   `json:"subtitle,omitempty"`
	Author           *string   `json:"author,omitempty"`
	Description      *string   `json:"description,omitempty"`
	WordCount        int       `json:"word_count"`
	CharCount        int       `json:"char_count"`
	CreatedAtUnix    int64     `json:"created_at_unix"`
	UpdatedAtUnix    int64     `json:"updated_at_unix"`
	AddedAtUnix      int64     `json:"added_at_unix"`
	Source           string    `json:"source,omitempty"`
	SourceField      string    `json:"source_field,omitempty"`
	URL              *string   `json:"url,omitempty"`
	ArticleImageURL  *string   `json:"article_image_url,omitempty"`
	IsArchived       bool      `json:"is_archived"`
	InUserLibrary    bool      `json:"in_user_library"`
	CanEdit          bool      `json:"can_edit"`
	CanDelete        bool      `json:"can_delete"`
	CreationStatus   string    `json:"creation_status,omitempty"`
	CreationProgress float64   `json:"creation_progress,omitempty"`
	Language         *string   `json:"language,omitempty"`
	Category         *string   `json:"category,omitempty"`
	Genre            *string   `json:"genre,omitempty"`
	ContentType      *string   `json:"content_type,omitempty"`
	AudioDurationSec *float64  `json:"audio_duration_seconds,omitempty"`
	Chapters         []Chapter `json:"chapters,omitempty"`
}

// Chapter represents a chapter or section in a read.
type Chapter struct {
	ChapterID   string `json:"chapter_id,omitempty"`
	Title       string `json:"title,omitempty"`
	StartOffset int    `json:"start_offset,omitempty"`
	EndOffset   int    `json:"end_offset,omitempty"`
}

// ReadsListResponse represents the response when listing reads.
type ReadsListResponse struct {
	Reads      []Read  `json:"reads"`
	HasMore    bool    `json:"has_more"`
	LastSortID *string `json:"last_sort_id,omitempty"`
	NextCursor *string `json:"next_cursor,omitempty"`
}

// ReadUpdatePayload represents fields that can be updated on a read item.
type ReadUpdatePayload struct {
	Title        *string `json:"title,omitempty"`
	Author       *string `json:"author,omitempty"`
	Description  *string `json:"description,omitempty"`
	MarkedUnread *bool   `json:"marked_as_unread,omitempty"`
}

// UserCollection represents a folder or collection in ElevenReader.
type UserCollection struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Icon        *string `json:"icon,omitempty"`
	ItemCount   int     `json:"item_count"`
	IsOwner     bool    `json:"is_owner"`
	IsEditable  bool    `json:"is_editable"`
	Description *string `json:"description,omitempty"`
}

// CollectionsResponse wraps the list of collections returned by ElevenReader API.
type CollectionsResponse struct {
	Collections []UserCollection `json:"collections"`
}

// Bookmark represents a saved highlight or location within a read.
type Bookmark struct {
	BookmarkID string  `json:"bookmark_id"`
	ReadID     string  `json:"read_id"`
	CharOffset int     `json:"char_offset"`
	Note       *string `json:"note,omitempty"`
	Quote      *string `json:"quote,omitempty"`
	CreatedAt  int64   `json:"created_at_unix,omitempty"`
}

// APIErrorDetail wraps the detail object from ElevenLabs API errors.
type APIErrorDetail struct {
	Type      string `json:"type,omitempty"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	Status    string `json:"status,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// APIError is returned when ElevenReader API responds with HTTP >= 400.
type APIError struct {
	StatusCode int
	Detail     APIErrorDetail
	RawBody    string
}

func (e *APIError) Error() string {
	if e.Detail.Message != "" {
		return fmt.Sprintf("API error (%d %s): %s", e.StatusCode, e.Detail.Code, e.Detail.Message)
	}
	if e.RawBody != "" {
		return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.RawBody)
	}
	return fmt.Sprintf("API error (%d)", e.StatusCode)
}
