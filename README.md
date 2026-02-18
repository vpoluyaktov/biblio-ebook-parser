# Biblio Ebook Parser

[![Go Reference](https://pkg.go.dev/badge/github.com/vpoluyaktov/biblio-ebook-parser.svg)](https://pkg.go.dev/github.com/vpoluyaktov/biblio-ebook-parser)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A unified ebook parser library for the Biblio suite that provides format-agnostic parsing, flexible rendering, and fast metadata extraction capabilities.

## Features

- **Multi-format support**: EPUB, FB2, and extensible for additional formats (MOBI, AZW3, PDF)
- **Fast extraction**: Extract covers, annotations, and metadata without parsing full content
- **Cover generation**: Generate beautiful placeholder covers with embedded fonts
- **Separation of concerns**: Parsing logic separated from output rendering
- **Pluggable renderers**: HTML (for web readers), PlainText (for TTS), and custom renderers
- **Robust error handling**: Handles malformed files, encoding issues, and edge cases
- **Thread-safe**: Safe for concurrent use
- **Type-safe**: Strongly-typed Go interfaces and data structures

## Architecture

```
biblio-ebook-parser/
├── parser/              # Core parser interfaces and registry
├── formats/             # Format-specific parsers
│   ├── epub/           # EPUB parser with fast extraction
│   └── fb2/            # FB2 parser with fast extraction
├── renderer/           # Output renderers
│   ├── html/           # HTML renderer (for web readers)
│   ├── plaintext/      # PlainText renderer (for TTS)
│   └── ssml/           # SSML renderer (future)
├── cover/              # Cover generation with embedded assets
└── testdata/           # Test fixtures
```

## Usage

### Basic Parsing

```go
import (
    "github.com/vpoluyaktov/biblio-ebook-parser/parser"
    "github.com/vpoluyaktov/biblio-ebook-parser/formats/epub"
)

// Parse an EPUB file
p := epub.NewParser()
book, err := p.Parse("/path/to/book.epub")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Title: %s\n", book.Metadata.Title)
fmt.Printf("Author: %s\n", book.Metadata.Authors[0].FullName())
fmt.Printf("Chapters: %d\n", len(book.Content.Chapters))
```

### Rendering for Web Reader (HTML)

```go
import (
    "github.com/vpoluyaktov/biblio-ebook-parser/renderer/html"
)

// Render book content as HTML for web reader
renderer := html.NewRenderer(html.Config{
    PreserveStructure: true,
})

content, err := renderer.RenderContent(book)
if err != nil {
    log.Fatal(err)
}

// content.Chapters[0].Content contains HTML
```

### Rendering for TTS (PlainText)

```go
import (
    "github.com/vpoluyaktov/biblio-ebook-parser/renderer/plaintext"
)

// Render book content as plain text for TTS
renderer := plaintext.NewRenderer(plaintext.Config{
    AddPeriods:    true,  // Add periods to paragraphs
    InsertMarkers: true,  // Insert SSML markers
    NormalizeText: true,  // Normalize text for speech
})

content, err := renderer.RenderContent(book)
if err != nil {
    log.Fatal(err)
}

// content.Chapters[0].Content contains plain text
```

### Fast Cover Extraction (Without Full Parsing)

```go
import "github.com/vpoluyaktov/biblio-ebook-parser/parser"

// Extract cover without parsing full book content (much faster!)
coverData, mimeType, err := parser.ExtractCover("/path/to/book.epub")
if err != nil {
    log.Fatal(err)
}

// coverData contains the image bytes
// mimeType is "image/jpeg" or "image/png"
```

### Fast Annotation Extraction

```go
import "github.com/vpoluyaktov/biblio-ebook-parser/parser"

// Extract book description/annotation without parsing full content
annotation, err := parser.ExtractAnnotation("/path/to/book.fb2")
if err != nil {
    log.Fatal(err)
}

fmt.Println(annotation)
```

### Fast Metadata Extraction

```go
import "github.com/vpoluyaktov/biblio-ebook-parser/parser"

// Extract only metadata without parsing content
metadata, err := parser.ExtractMetadata("/path/to/book.epub")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Title: %s\n", metadata.Title)
fmt.Printf("Authors: %v\n", metadata.Authors)
fmt.Printf("Has cover: %v\n", len(metadata.CoverData) > 0)
```

### Generate Placeholder Cover

```go
import "github.com/vpoluyaktov/biblio-ebook-parser/cover"

// Generate a beautiful cover image when book has no cover
coverData, err := cover.GeneratePlaceholder("The Great Gatsby", "F. Scott Fitzgerald")
if err != nil {
    log.Fatal(err)
}

// coverData contains JPEG image bytes
```

### Using the Registry

```go
import "github.com/vpoluyaktov/biblio-ebook-parser/parser"

// Get parser by format
p, err := parser.GetParser("epub")
if err != nil {
    log.Fatal(err)
}

book, err := p.Parse("/path/to/book.epub")
```

## Installation

```bash
go get github.com/vpoluyaktov/biblio-ebook-parser
```

## Supported Formats

- ✅ **EPUB** (2.0, 3.0) - Full support with TOC extraction
- ✅ **FB2** (FictionBook 2.0) - Full support with encoding handling
- 🚧 **MOBI** - Planned
- 🚧 **AZW3** - Planned
- 🚧 **PDF** - Planned

## API Documentation

Full API documentation is available at [pkg.go.dev](https://pkg.go.dev/github.com/vpoluyaktov/biblio-ebook-parser).

### Key Interfaces

- **`parser.Parser`** - Main parser interface for full book parsing
- **`parser.FastExtractor`** - Interface for fast metadata/cover extraction
- **`renderer.Renderer`** - Interface for rendering parsed content

### Performance Tips

1. **Use fast extraction** when you only need cover, annotation, or metadata
2. **Parse once, render multiple times** - Parse the book once, then use different renderers
3. **Use io.ReaderAt** when possible to avoid loading entire file into memory

## Testing

```bash
go test ./...
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.
