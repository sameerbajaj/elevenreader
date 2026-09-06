# ElevenReader CLI & Agent Context

## Architecture Overview
- **Go CLI**: High-performance, zero-dependency standalone CLI for [ElevenReader](https://elevenreader.io/reader/library).
- **Backend API**: `https://api.elevenlabs.io/v1/reader`
- **Authentication**:
  - Automatically searches local browser session profiles (Brave, Chrome, Edge, Arc) in IndexedDB/LevelDB storage.
  - Supports `--token` / `--api-key` flags, environment variables (`ELEVENREADER_TOKEN`, `ELEVENLABS_API_KEY`), and stored configuration (`~/.config/elevenreader/config.json`).

## Critical API Gotchas & Learnings
1. **Multipart Upload Headers**:
   - `POST /reads/add/v2` requires explicit `Content-Type: text/plain` (or `application/pdf`, etc.) on the `from_document` part. The Go standard library's `CreateFormFile` hardcodes `application/octet-stream`, which causes the backend document pipeline to fail or reject the file.
   - Always pass `parse_content="true"` in the multipart form to trigger background extraction.
2. **Sort Enum Values**:
   - `sort_by` strictly accepts: `recently_added_desc`, `recently_added_asc`, `recently_listened_desc`, `recently_listened_asc`.
3. **Markdown Sanitization**:
   - ElevenReader returns timing spans `<span c="...">word</span>` for audio synchronization. The CLI strips these by default for readable markdown, with `--raw` preserving them.
4. **Readwise Sync Integration**:
   - `readwise-to-elevenreader` skill (`~/.gemini/config/skills/readwise-to-elevenreader/SKILL.md`) bridges longform essays from Readwise Reader into ElevenReader, strictly excluding videos, podcasts, tweets, and short notes.
