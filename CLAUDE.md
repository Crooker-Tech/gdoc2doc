# gdoc2doc

A Go command-line tool that uses natural language queries (via Together AI) to filter Google Docs and export them to various formats.

## Project Structure

```
gdoc2doc/
├── main.go              # CLI entry point
├── config.go            # Token and API key loading
├── google_drive.go      # Google Drive API integration
├── together_filter.go   # Together AI document filtering
├── downloads/           # Downloaded documents (gitignored)
└── tools/
    ├── save-key.ps1           # Store API key in Windows Credential Manager
    ├── load-key.ps1           # Load API key to environment
    └── setup_google_oauth.py  # OAuth setup for Google Drive
```

## Downloads

All exported documents are saved to the `downloads/` directory by default. This directory is gitignored to avoid committing potentially large or sensitive documents.

## Usage

```powershell
# Load keys
. .\tools\load-key.ps1 -Service together -Target API
. .\tools\load-key.ps1 -Service google-docs -Target JWT

# Export to PDF (default)
.\gdoc2doc.exe "meeting notes"

# Export to other formats
.\gdoc2doc.exe -t md "project proposal"
.\gdoc2doc.exe -t docx "report"
.\gdoc2doc.exe -type html -output ./exports "documentation"

# List all documents
.\gdoc2doc.exe -list
```

## Supported Export Formats

| Format | Extension | Flag |
|--------|-----------|------|
| PDF | .pdf | `-t pdf` (default) |
| Word | .docx | `-t docx` |
| OpenDocument | .odt | `-t odt` |
| Rich Text | .rtf | `-t rtf` |
| Plain Text | .txt | `-t txt` |
| HTML | .html | `-t html` |
| EPUB | .epub | `-t epub` |
| Markdown | .md | `-t md` or `-t markdown` |

## Interactive Selection

When multiple documents match your query, you'll be prompted to:
- Enter a number to export a specific document
- Enter `all` to export all matching documents
- Enter `q` to quit

## Setup

### 1. Google Cloud Console

1. Go to https://console.cloud.google.com/apis/credentials
2. Create OAuth 2.0 Client ID (Desktop application)
3. Enable Google Drive API
4. Run: `python tools/setup_google_oauth.py`

### 2. Together AI API Key

1. Get API key from https://api.together.xyz/settings/api-keys
2. Store: `.\tools\save-key.ps1 -Service together -Target API -Key <your_key>`

### 3. Build & Run

```bash
go build -o gdoc2doc.exe .
```

## SKYGOD Principles Applied

- **S (SOLID)**: Each Go file has single responsibility
- **K (KISS)**: Simple HTTP client for Together AI, no framework overhead
- **Y (YAGNI)**: Only implements what's needed for the use case
- **G (GRASP)**: High cohesion within modules, low coupling between them
- **O (O&O)**: Observable naming (`documentFilter` not `df`, `outputDirectory` not `dir`)
- **D (DRY)**: Shared credential patterns via PowerShell scripts

## CLAUDE.md Rules

1. **Python: UTF-8 Encoding**: Python scripts start with:
   ```python
   import io, sys
   if sys.stdout.encoding != 'utf-8':
       sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
   ```

2. **Python: Use `uv pip`**: Always prefer `uv pip install` with fallback.

3. **Python: Use `pathlib.Path`**: All file operations use Path, not string paths.

4. **Observable Naming**: No aggressive abbreviations.

5. **No Generic Names**: Avoid "manager", "utilities", "helper", "service".

## Free Model

Uses `ServiceNow-AI/Apriel-1.6-15b-Thinker` (free usage tier on Together AI).

## Key Assumptions

1. **Google Docs Only**: Only exports Google Docs (not Sheets, Slides, or other formats)
2. **Trashed Excluded**: Documents in trash are automatically excluded from listing
3. **OAuth Token Format**: The `GOOGLE-DOCS_JWT_KEY` environment variable expects JSON with:
   - `token`, `refresh_token`, `token_uri`, `client_id`, `client_secret`, `scopes`
4. **Filename Sanitization**: Invalid filename characters (`\ / : * ? " < > |`) are replaced with underscores
5. **Flexible Matching**: AI-filtered document names match via case-insensitive exact match or contains
6. **Go 1.21+**: Requires Go 1.21 or higher (uses `golang.org/x/oauth2` and `google.golang.org/api`)
