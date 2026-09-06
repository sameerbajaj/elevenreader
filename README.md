# elevenreader

[![Go Version](https://img.shields.io/github/go-mod/go-version/sameerbajaj/elevenreader)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A fast, lightweight, and scriptable command-line client for [ElevenReader](https://elevenreader.io/reader/library).

Manage your ElevenReader library directly from your terminal: list items, read clean Markdown transcripts, import articles from URLs or documents, organize collections, export bookmarks, and automate CRUD workflows with full JSON output.

---

## Features

- **Full Library CRUD**: List library items, view metadata, read text transcripts, update titles/authors, archive, restore, and delete reads.
- **Multiple Import Sources**: Add items by web article URL, upload documents (PDF, EPUB, TXT), or create text notes directly.
- **Terminal Reading**: Read article transcripts directly in your terminal with word-timing span tags cleanly stripped, or save Markdown directly to disk.
- **Collections & Bookmarks**: Create and manage custom folders, organize library items into collections, and export highlights as Markdown.
- **Zero-Config Browser Auth**: Automatically import active session credentials from your local Brave, Chrome, Edge, or Arc browser profiles with `elevenreader auth import`.
- **Scriptable & Pipe-Friendly**: Every command supports `--json` for effortless integration with `jq`, `fzf`, and shell pipelines.
- **Shell Autocompletion**: Built-in support for `bash`, `zsh`, `fish`, and `powershell`.

---

## Installation

### Using Go

```bash
go install github.com/sameerbajaj/elevenreader/cmd/elevenreader@latest
```

### Build from Source

```bash
git clone https://github.com/sameerbajaj/elevenreader.git
cd elevenreader
go build -o elevenreader ./cmd/elevenreader
mv elevenreader /usr/local/bin/
```

---

## Authentication

`elevenreader` supports several convenient ways to authenticate:

### 1. Automatic Browser Import (Recommended)

If you are already logged into [ElevenReader](https://elevenreader.io) in Brave, Google Chrome, Microsoft Edge, or Arc, simply run:

```bash
elevenreader auth import
```

This scans your local browser storage for an active session token, validates it against the ElevenReader API, and saves it to `~/.config/elevenreader/config.json`.

### 2. Interactive Login

```bash
elevenreader auth login
```

Prompts for your ElevenReader session token or ElevenLabs API key (`sk_...`) and stores it securely.

### 3. Environment Variables

Set either environment variable in your shell profile (`~/.zshrc` or `~/.bashrc`):

```bash
export ELEVENREADER_TOKEN="your_session_token"
# OR
export ELEVENLABS_API_KEY="your_api_key"
```

### 4. CLI Flags

Pass credentials per-command:

```bash
elevenreader list --token "your_token"
# OR
elevenreader list --api-key "sk_..."
```

Check your current authentication status anytime:

```bash
elevenreader auth status
```

---

## Usage & Command Reference

### Library CRUD

#### List Library Reads

```bash
# List recent items (default: 30)
elevenreader list

# Paginate with custom limit and sort order
elevenreader list --page-size 10 --sort-by recently_added_desc

# Output as JSON
elevenreader list --json | jq '.reads[] | {id: .read_id, title: .title}'
```

Available sort options:
- `recently_added_desc` (default)
- `recently_added_asc`
- `recently_listened_desc`
- `recently_listened_asc`

#### Inspect a Read

```bash
elevenreader get <read-id>
elevenreader get u:wKUZ9PZ57r4BZMjDc2Ct --json
```

#### Read & Export Transcripts

```bash
# Print clean Markdown to terminal
elevenreader read <read-id>

# Save clean Markdown to file
elevenreader read <read-id> -o article.md

# Fetch raw source text instead of Markdown
elevenreader read <read-id> --source

# Include original word-timing span tags
elevenreader read <read-id> --raw
```

#### Add / Import Reads

```bash
# Import a web article by URL
elevenreader add --url "https://en.wikipedia.org/wiki/Go_(programming_language)"

# Upload a local document (PDF, EPUB, TXT)
elevenreader add --file paper.pdf --title "Research Paper" --author "Author Name"

# Create a read from raw text
elevenreader add --text "Quick reading note content" --title "My Meeting Notes"
```

#### Update Metadata

```bash
elevenreader update <read-id> --title "Updated Title" --author "New Author"
elevenreader update <read-id> --unread
```

#### Archive & Unarchive

```bash
# Move to archive
elevenreader archive <read-id>

# Restore from archive
elevenreader unarchive <read-id>
```

#### Delete a Read

```bash
# Interactive confirmation prompt
elevenreader delete <read-id>

# Skip confirmation (-y / --yes)
elevenreader delete <read-id> -y
```

---

### Collections

Organize reads into folders and custom collections:

```bash
# List all collections
elevenreader collection list

# Create a new collection
elevenreader collection create "AI Safety" --icon "🤖"

# Add a read to a collection
elevenreader collection add <collection-id> <read-id>

# Remove a read from a collection
elevenreader collection remove <collection-id> <read-id>

# Delete a collection
elevenreader collection delete <collection-id>
```

---

### Bookmarks

Manage highlights and saved positions:

```bash
# List bookmarks for a read
elevenreader bookmark list <read-id>

# Export all bookmarks for a read as Markdown
elevenreader bookmark export <read-id> -o bookmarks.md

# Delete a bookmark
elevenreader bookmark delete <bookmark-id>
```

---

### Shell Autocompletion

Enable tab autocompletion for your shell:

```bash
# Zsh
elevenreader completion zsh > "${fpath[1]}/_elevenreader"

# Bash
elevenreader completion bash > /etc/bash_completion.d/elevenreader

# Fish
elevenreader completion fish > ~/.config/fish/completions/elevenreader.fish
```

---

## Power User Workflows

### Interactive Fuzzy Finder with `fzf`

Quickly browse and open any article from your ElevenReader library:

```bash
elevenreader list --json | jq -r '.reads[] | "\(.read_id)\t\(.title)"' | \
  fzf --with-nth=2.. --delimiter='\t' --preview='elevenreader read {1}'
```

### Batch Export Library to Markdown Files

Export your entire ElevenReader library to local Markdown files:

```bash
elevenreader list --page-size 100 --json | jq -r '.reads[] | "\(.read_id)\t\(.title)"' | while IFS=$'\t' read -r id title; do
  safe_title=$(echo "$title" | tr '/:' '--')
  echo "Exporting: $safe_title"
  elevenreader read "$id" -o "${safe_title}.md"
done
```

---

## License

MIT License. See [LICENSE](LICENSE) for details.
