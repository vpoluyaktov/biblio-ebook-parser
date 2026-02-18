// Package parser provides a unified ebook parsing library with support for multiple formats.
//
// The library separates parsing logic from output rendering, allowing you to parse once
// and render in multiple formats (HTML for web readers, plain text for TTS, etc.).
//
// # Supported Formats
//
// Currently supported formats:
//   - EPUB (2.0, 3.0)
//   - FB2 (FictionBook 2.0)
//
// # Basic Usage
//
// Parse an ebook file:
//
//	import (
//	    "github.com/vpoluyaktov/biblio-ebook-parser/parser"
//	    "github.com/vpoluyaktov/biblio-ebook-parser/formats/epub"
//	)
//
//	p := epub.NewParser()
//	book, err := p.Parse("/path/to/book.epub")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Fast Extraction
//
// Extract cover, annotation, or metadata without parsing full content (much faster):
//
//	// Extract cover image from file
//	coverData, mimeType, err := parser.ExtractCoverFromFile("/path/to/book.epub")
//
//	// Extract cover from reader (e.g., from ZIP archive)
//	coverData, mimeType, err := parser.ExtractCoverFromReader(reader, size, "epub")
//
//	// Extract annotation/description from file
//	annotation, err := parser.ExtractAnnotationFromFile("/path/to/book.fb2")
//
//	// Extract metadata from file
//	metadata, err := parser.ExtractMetadataFromFile("/path/to/book.epub")
//
// # Rendering
//
// Render parsed content in different formats:
//
//	import (
//	    "github.com/vpoluyaktov/biblio-ebook-parser/renderer/html"
//	    "github.com/vpoluyaktov/biblio-ebook-parser/renderer/plaintext"
//	)
//
//	// HTML for web readers
//	htmlRenderer := html.NewRenderer(html.Config{})
//	htmlContent, err := htmlRenderer.RenderContent(book)
//
//	// Plain text for TTS
//	textRenderer := plaintext.NewRenderer(plaintext.Config{AddPeriods: true})
//	textContent, err := textRenderer.RenderContent(book)
//
// # Cover Generation
//
// Generate placeholder covers when books don't have covers:
//
//	import "github.com/vpoluyaktov/biblio-ebook-parser/cover"
//
//	coverData, err := cover.GeneratePlaceholder("Book Title", "Author Name")
//
// # Thread Safety
//
// All parsers and extractors are safe for concurrent use. The parser registry
// uses proper locking to ensure thread-safe registration and retrieval.
package parser
