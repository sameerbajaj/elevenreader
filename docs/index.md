# ElevenReader CLI

[![Canonical Site](https://img.shields.io/badge/Canonical%20Site-sameerbajaj.com%2Ftools%2Felevenreader-blue.svg)](https://sameerbajaj.com/tools/elevenreader/)
[![Go Reference](https://pkg.go.dev/badge/github.com/sameerbajaj/elevenreader.svg)](https://pkg.go.dev/github.com/sameerbajaj/elevenreader)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/sameerbajaj/elevenreader/blob/main/LICENSE)

A fast, lightweight, and scriptable command-line client for [ElevenReader](https://elevenreader.io/reader/library), built by [Sameer Bajaj](https://sameerbajaj.com).

Official canonical documentation and project hub:  
**[https://sameerbajaj.com/tools/elevenreader/](https://sameerbajaj.com/tools/elevenreader/)**

---

## Overview

`elevenreader` enables seamless command-line management of your ElevenReader reading library:

- **Full Library CRUD**: List library items, view metadata, read text transcripts, update titles/authors, archive, restore, and delete reads.
- **Multiple Import Sources**: Add items by web article URL, upload documents (PDF, EPUB, TXT), or create text notes directly.
- **Terminal Reading**: Read article transcripts directly in your terminal with word-timing span tags cleanly stripped.
- **Collections & Bookmarks**: Manage custom folders, organize items into collections, and export highlights as Markdown.
- **Zero-Config Browser Auth**: Automatically import active session credentials from Brave, Chrome, Edge, or Arc browser profiles.
- **JSON Automation**: Clean `--json` envelopes for shell pipelines and CLI automation.

---

## Installation

### Using Go

```bash
go install github.com/sameerbajaj/elevenreader/cmd/elevenreader@latest
```

### Precompiled Binaries

Download precompiled binaries for macOS, Linux, and Windows from [GitHub Releases](https://github.com/sameerbajaj/elevenreader/releases).

---

## Quick Reference

```bash
# Authenticate automatically from browser
elevenreader auth import

# List reading items
elevenreader reads list

# Read an article transcript in terminal
elevenreader reads view <read-id> --clean

# Import a document
elevenreader reads add document ./book.epub

# Export bookmarks
elevenreader bookmarks list --format markdown
```

For full documentation and updates, visit [https://sameerbajaj.com/tools/elevenreader/](https://sameerbajaj.com/tools/elevenreader/).
