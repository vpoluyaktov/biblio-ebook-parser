# Biblio Ebook Parser - Specification

## Overview

The Biblio Ebook Parser is a unified library that provides ebook parsing capabilities for the entire Biblio suite. It addresses the code duplication problem between `biblio-ebooks-catalog` and `biblio-audiobook-builder-tts` by providing a single, well-tested implementation of ebook parsers with pluggable output renderers.

## Goals

1. **Eliminate code duplication** - Single source of truth for ebook parsing logic
2. **Separation of concerns** - Decouple parsing from rendering
3. **Extensibility** - Easy to add new formats and renderers
4. **Robustness** - Handle malformed files, encoding issues, edge cases
5. **Performance** - Efficient parsing with minimal memory overhead
6. **Maintainability** - Well-tested, documented, and type-safe

## Architecture

### Core Concepts

#### 1. Book Structure
The library uses a format-agnostic `Book` structure that represents parsed ebook content:

```go
type Book struct {
    Metadata Metadata
    Content  Content
}
```

#### 2. Element-Based Content Model
Content is represented as a tree of typed elements rather than raw HTML or text:

```go
type Element interface {
    Type() ElementType
}

// Element types: Paragraph, Heading, Image, Table, EmptyLine, Epigraph
```

This allows renderers to convert the same parsed structure into different output formats.

#### 3. Parser Interface
Format-specific parsers implement a common interface:

```go
type Parser interface {
    Parse(filePath string) (*Book, error)
    ParseReader(r io.ReaderAt, size int64) (*Book, error)
    Format() string
}
```

#### 4. Renderer Interface
Renderers convert the parsed `Book` into specific output formats:

```go
type Renderer interface {
    RenderMetadata(book *Book) (interface{}, error)
    RenderContent(book *Book) (interface{}, error)
}
```

### Package Structure

```
github.com/vpoluyaktov/biblio-ebook-parser/
├── parser/                  # Core interfaces and registry
│   ├── parser.go           # Parser interface, Book/Metadata/Content types
│   ├── registry.go         # Parser registry
│   ├── element.go          # Element types and interfaces
│   └── author.go           # Author type
├── formats/                # Format-specific parsers
│   ├── epub/
│   │   ├── epub.go         # EPUB parser implementation
│   │   ├── metadata.go     # EPUB metadata extraction
│   │   ├── content.go      # EPUB content extraction
│   │   ├── toc.go          # TOC parsing (NCX, nav.xhtml)
│   │   ├── epub_test.go    # EPUB parser tests
│   │   └── testdata/       # EPUB test files
│   └── fb2/
│       ├── fb2.go          # FB2 parser implementation
│       ├── metadata.go     # FB2 metadata extraction
│       ├── content.go      # FB2 content extraction
│       ├── sanitize.go     # XML sanitization
│       ├── fb2_test.go     # FB2 parser tests
│       └── testdata/       # FB2 test files
├── renderer/               # Output renderers
│   ├── html/
│   │   ├── html.go         # HTML renderer for web readers
│   │   └── html_test.go
│   ├── plaintext/
│   │   ├── plaintext.go    # PlainText renderer for TTS
│   │   ├── periods.go      # Period insertion logic
│   │   └── plaintext_test.go
│   └── ssml/               # Future: SSML renderer
│       └── ssml.go
└── testdata/               # Shared test fixtures
```

## Implementation Plan

### Phase 1: Core Infrastructure ✅
- [x] Create GitHub repository
- [x] Initialize Go module
- [x] Define core interfaces (Parser, Renderer, Element)
- [x] Implement parser registry
- [x] Create README and Specification

### Phase 2: EPUB Parser Migration
- [ ] Extract EPUB parsing logic from biblio-ebooks-catalog
- [ ] Implement element-based content extraction
- [ ] Support NCX and nav.xhtml TOC formats
- [ ] Handle anchor-based chapter splitting
- [ ] Add comprehensive tests with real EPUB files
- [ ] Test coverage: metadata, TOC, content, cover images

### Phase 3: FB2 Parser Migration
- [ ] Extract FB2 parsing logic from biblio-ebooks-catalog
- [ ] Implement XML sanitization (UTF-8 fixing, encoding support)
- [ ] Support nested sections with depth tracking
- [ ] Handle Windows-1251 and other encodings
- [ ] Add comprehensive tests with real FB2 files
- [ ] Test coverage: metadata, sections, encoding, malformed XML

### Phase 4: HTML Renderer
- [ ] Implement HTML renderer for web readers
- [ ] Convert elements to HTML structure
- [ ] Preserve formatting and structure
- [ ] Match current biblio-ebooks-catalog output
- [ ] Add tests comparing output

### Phase 5: PlainText Renderer
- [ ] Implement PlainText renderer for TTS
- [ ] Convert elements to plain text
- [ ] Add period insertion logic
- [ ] Support SSML marker insertion
- [ ] Match current biblio-audiobook-builder-tts output
- [ ] Add tests comparing output

### Phase 6: Integration - biblio-ebooks-catalog
- [ ] Create feature branch: `feature/unified-ebook-parser`
- [ ] Add dependency on biblio-ebook-parser
- [ ] Replace internal parser with shared library + HTML renderer
- [ ] Update API handlers
- [ ] Run integration tests
- [ ] Verify web reader functionality
- [ ] Update Specification.md
- [ ] Create PR and merge

### Phase 7: Integration - biblio-audiobook-builder-tts
- [ ] Create feature branch: `feature/unified-ebook-parser`
- [ ] Add dependency on biblio-ebook-parser
- [ ] Replace internal parser with shared library + PlainText renderer
- [ ] Configure renderer for TTS (periods, markers)
- [ ] Run integration tests
- [ ] Verify audiobook generation
- [ ] Update Specification.md
- [ ] Create PR and merge

### Phase 8: Future Enhancements
- [ ] Add MOBI format support
- [ ] Add AZW3 format support
- [ ] Add PDF text extraction
- [ ] Implement SSML renderer
- [ ] Add format auto-detection
- [ ] Performance optimizations
- [ ] Streaming parser for large files

## Design Decisions

### Why Element-Based Content Model?

Instead of storing raw HTML or plain text, we use typed elements:

**Advantages:**
- Single parsing pass, multiple output formats
- Renderers can apply format-specific transformations
- Easier to test and validate
- Type-safe operations on content

**Example:**
```go
// Parser extracts structure
chapter.Elements = []Element{
    &Heading{Text: "Chapter 1", Level: 1},
    &Paragraph{Text: "Once upon a time..."},
    &Image{Alt: "Map", Href: "images/map.jpg"},
}

// HTML renderer produces:
<h1>Chapter 1</h1>
<p>Once upon a time...</p>
<img src="images/map.jpg" alt="Map">

// PlainText renderer produces:
Chapter 1

Once upon a time.

[Image: Map]
```

### Why Separate Renderers?

Different use cases require different output:
- **Web reader**: Needs HTML with structure, images, formatting
- **TTS**: Needs plain text with periods, SSML markers, no images
- **API**: Might need JSON, Markdown, or other formats

Separating rendering from parsing allows each consumer to get exactly what they need.

### Error Handling Strategy

1. **Metadata parsing**: Fail fast on critical errors (file not found, invalid format)
2. **Content parsing**: Best-effort with fallbacks (missing TOC → use spine, malformed XML → sanitize)
3. **Encoding issues**: Auto-detect and convert (Windows-1251, UTF-8, etc.)
4. **Malformed files**: Sanitize and attempt recovery before failing

## Testing Strategy

### Unit Tests
- Test each parser independently
- Test each renderer independently
- Mock data for edge cases

### Integration Tests
- Real EPUB/FB2 files from various sources
- Compare output with current implementations
- Test encoding variations
- Test malformed files

### Regression Tests
- Ensure bug fixes don't regress
- Test files that previously caused issues

### Performance Tests
- Benchmark parsing speed
- Memory usage profiling
- Large file handling

## Migration Strategy

### Backward Compatibility
During migration, both old and new implementations will coexist:

1. Add biblio-ebook-parser as dependency
2. Create adapter layer if needed
3. Run tests with both implementations
4. Compare outputs for consistency
5. Switch to new implementation
6. Remove old code

### Rollback Plan
If issues arise:
1. Feature branches allow easy rollback
2. Old implementation remains until verified
3. Can run both in parallel for comparison

## Success Criteria

- [ ] All existing EPUB/FB2 parsing tests pass
- [ ] Web reader displays books correctly
- [ ] TTS audiobook generation works correctly
- [ ] No regression in functionality
- [ ] Code duplication eliminated
- [ ] New formats can be added easily
- [ ] Performance is equal or better

## Current Status

**Phase 1: Core Infrastructure** - ✅ In Progress
- Repository created
- Go module initialized
- README and Specification drafted
- Next: Implement core interfaces
