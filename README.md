# Biblio Ebook Parser

A unified ebook parser library for the Biblio suite that provides format-agnostic parsing and flexible rendering capabilities.

## Features

- **Multi-format support**: EPUB, FB2, and extensible for additional formats (MOBI, AZW3, PDF)
- **Separation of concerns**: Parsing logic separated from output rendering
- **Pluggable renderers**: HTML (for web readers), PlainText (for TTS), and custom renderers
- **Robust error handling**: Handles malformed files, encoding issues, and edge cases
- **Comprehensive testing**: Extensive test coverage for all parsers and renderers
- **Type-safe**: Strongly-typed Go interfaces and data structures

## Architecture

```
biblio-ebook-parser/
├── parser/              # Core parser interfaces and registry
├── formats/             # Format-specific parsers
│   ├── epub/           # EPUB parser
│   └── fb2/            # FB2 parser
├── renderer/           # Output renderers
│   ├── html/           # HTML renderer (for web readers)
│   ├── plaintext/      # PlainText renderer (for TTS)
│   └── ssml/           # SSML renderer (future)
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

## Testing

```bash
go test ./...
```

## License

Private - Biblio Suite

## Contributing

This is a private library for the Biblio suite. For issues or feature requests, please contact the maintainer.
