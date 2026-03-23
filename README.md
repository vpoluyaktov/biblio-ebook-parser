# Biblio Ebook Parser

> Part of the [BiblioHub](https://github.com/vpoluyaktov/biblio-hub) application suite

A shared Go library for e-book parsing used by [biblio-ebooks-catalog](https://github.com/vpoluyaktov/biblio-ebooks-catalog) and [biblio-audiobook-builder-tts](https://github.com/vpoluyaktov/biblio-audiobook-builder-tts). Provides format-agnostic parsing, pluggable rendering, and fast metadata extraction.

**Live demo (via Biblio Catalog): [https://demo.bibliohub.org/catalog/](https://demo.bibliohub.org/catalog/)**

## Features

- **Multi-format support** — EPUB (2.0, 3.0) and FB2 (FictionBook 2.0)
- **Fast extraction** — Extract covers, annotations, and metadata without parsing full content
- **Cover generation** — Generate placeholder covers with embedded fonts
- **Pluggable renderers** — HTML (for web readers), PlainText (for TTS)
- **Robust error handling** — Handles malformed files, encoding issues, and edge cases
- **Thread-safe** — Safe for concurrent use

## Technology Stack

- **Language**: Go 1.24+
- **Formats**: EPUB, FB2
- **Type**: Library (imported as Go module)

## Installation

```bash
go get github.com/vpoluyaktov/biblio-ebook-parser
```

## Architecture

```
biblio-ebook-parser/
├── parser/              # Core parser interfaces and registry
├── formats/
│   ├── epub/            # EPUB parser with fast extraction
│   └── fb2/             # FB2 parser with fast extraction
├── renderer/
│   ├── html/            # HTML renderer (for web readers)
│   └── plaintext/       # PlainText renderer (for TTS)
├── cover/               # Placeholder cover generation
└── testdata/            # Test fixtures
```

## Usage

### Basic Parsing

```go
import "github.com/vpoluyaktov/biblio-ebook-parser/formats/epub"

p := epub.NewParser()
book, err := p.Parse("/path/to/book.epub")

fmt.Printf("Title: %s\n", book.Metadata.Title)
fmt.Printf("Chapters: %d\n", len(book.Content.Chapters))
```

### Fast Cover Extraction

```go
import "github.com/vpoluyaktov/biblio-ebook-parser/parser"

coverData, mimeType, err := parser.ExtractCoverFromFile("/path/to/book.epub")
```

### Fast Metadata Extraction

```go
import "github.com/vpoluyaktov/biblio-ebook-parser/parser"

metadata, err := parser.ExtractMetadataFromFile("/path/to/book.epub")
```

### Rendering for TTS

```go
import "github.com/vpoluyaktov/biblio-ebook-parser/renderer/plaintext"

renderer := plaintext.NewRenderer(plaintext.Config{
    AddPeriods:    true,
    InsertMarkers: true,
    NormalizeText: true,
})
content, err := renderer.RenderContent(book)
```

### Rendering for Web Reader

```go
import "github.com/vpoluyaktov/biblio-ebook-parser/renderer/html"

renderer := html.NewRenderer(html.Config{PreserveStructure: true})
content, err := renderer.RenderContent(book)
```

### Placeholder Cover Generation

```go
import "github.com/vpoluyaktov/biblio-ebook-parser/cover"

coverData, err := cover.GeneratePlaceholder("The Great Gatsby", "F. Scott Fitzgerald")
```

## Key Interfaces

- **`parser.Parser`** — Full book parsing
- **`parser.FastExtractor`** — Fast metadata/cover extraction without full parse
- **`renderer.Renderer`** — Render parsed content to different output formats

## Testing

```bash
go test ./...
```

## License

MIT
