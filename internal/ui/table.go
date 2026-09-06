package ui

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/sameerbajaj/elevenreader/internal/api"
)

// PrintReadsTable prints reads in a clean aligned table.
func PrintReadsTable(reads []api.Read) {
	if len(reads) == 0 {
		fmt.Println("No items found in your ElevenReader library.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tAUTHOR\tWORDS\tSTATUS\tADDED")

	for _, r := range reads {
		title := r.Title
		if title == "" {
			title = "(untitled)"
		}
		title = Truncate(title, 42)

		author := "-"
		if r.Author != nil && *r.Author != "" {
			author = Truncate(*r.Author, 20)
		}

		status := "active"
		if r.IsArchived {
			status = "archived"
		}

		added := FormatTimeRelative(r.AddedAtUnix)
		words := FormatNumber(r.WordCount)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			r.ReadID,
			title,
			author,
			words,
			status,
			added,
		)
	}
	w.Flush()
}

// PrintReadDetails prints detailed metadata for a single read item.
func PrintReadDetails(r *api.Read) {
	title := r.Title
	if title == "" {
		title = "(untitled)"
	}

	author := "-"
	if r.Author != nil && *r.Author != "" {
		author = *r.Author
	}

	fmt.Printf("ID:          %s\n", r.ReadID)
	fmt.Printf("Title:       %s\n", title)
	if r.Subtitle != nil && *r.Subtitle != "" {
		fmt.Printf("Subtitle:    %s\n", *r.Subtitle)
	}
	fmt.Printf("Author:      %s\n", author)
	fmt.Printf("Words:       %s (%s chars)\n", FormatNumber(r.WordCount), FormatNumber(r.CharCount))
	fmt.Printf("Status:      %s\n", map[bool]string{true: "archived", false: "active"}[r.IsArchived])
	fmt.Printf("Added:       %s\n", FormatTimeRelative(r.AddedAtUnix))
	fmt.Printf("Updated:     %s\n", FormatTimeRelative(r.UpdatedAtUnix))

	if r.Source != "" {
		fmt.Printf("Source:      %s\n", r.Source)
	}
	if r.URL != nil && *r.URL != "" {
		fmt.Printf("URL:         %s\n", *r.URL)
	}
	if r.Description != nil && *r.Description != "" {
		desc := strings.TrimSpace(*r.Description)
		fmt.Printf("Description: %s\n", Truncate(desc, 120))
	}
	if len(r.Chapters) > 0 {
		fmt.Printf("Chapters:    %d\n", len(r.Chapters))
		for i, ch := range r.Chapters {
			if i >= 5 {
				fmt.Printf("  ... and %d more chapters\n", len(r.Chapters)-5)
				break
			}
			chTitle := ch.Title
			if chTitle == "" {
				chTitle = fmt.Sprintf("Chapter %d", i+1)
			}
			fmt.Printf("  [%d] %s\n", i+1, chTitle)
		}
	}
}

// PrintCollectionsTable prints user collections.
func PrintCollectionsTable(cols []api.UserCollection) {
	if len(cols) == 0 {
		fmt.Println("No collections found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tITEMS\tOWNER")
	for _, c := range cols {
		title := c.Title
		if c.Icon != nil && *c.Icon != "" {
			title = fmt.Sprintf("%s %s", *c.Icon, title)
		}
		owner := "system"
		if c.IsOwner {
			owner = "you"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
			c.ID,
			title,
			c.ItemCount,
			owner,
		)
	}
	w.Flush()
}

// PrintBookmarksTable prints bookmarks.
func PrintBookmarksTable(bms []api.Bookmark) {
	if len(bms) == 0 {
		fmt.Println("No bookmarks found for this read.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tOFFSET\tQUOTE / NOTE\tCREATED")
	for _, b := range bms {
		text := "-"
		if b.Quote != nil && *b.Quote != "" {
			text = Truncate(*b.Quote, 50)
		} else if b.Note != nil && *b.Note != "" {
			text = Truncate(*b.Note, 50)
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
			b.BookmarkID,
			b.CharOffset,
			text,
			FormatTimeRelative(b.CreatedAt),
		)
	}
	w.Flush()
}
