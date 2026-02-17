package parser

import "io"

// Parser defines the interface for ebook parsers
type Parser interface {
	// Parse extracts book structure from a file path
	Parse(filePath string) (*Book, error)

	// ParseReader extracts book structure from an io.ReaderAt
	ParseReader(r io.ReaderAt, size int64) (*Book, error)

	// Format returns the format identifier (e.g., "epub", "fb2")
	Format() string
}

// Book represents a parsed ebook with metadata and content
type Book struct {
	Metadata Metadata
	Content  Content
}

// Metadata represents format-agnostic book metadata
type Metadata struct {
	Title       string
	Authors     []Author
	Language    string
	Description string
	Genres      []string
	Series      string
	SeriesIndex int
	CoverData   []byte
	CoverType   string // MIME type (e.g., "image/jpeg", "image/png")
}

// Content represents the structured content of a book
type Content struct {
	Chapters []Chapter
}

// Chapter represents a book chapter or section
type Chapter struct {
	ID       string
	Title    string
	Level    int       // TOC depth (0 = top level, 1 = subsection, etc.)
	Elements []Element // Content elements
}

// GetTotalCharacters returns the total character count across all chapters
func (b *Book) GetTotalCharacters() int {
	total := 0
	for _, ch := range b.Content.Chapters {
		for _, elem := range ch.Elements {
			total += elem.CharCount()
		}
	}
	return total
}

// GetTotalWords returns the approximate word count across all chapters
func (b *Book) GetTotalWords() int {
	total := 0
	for _, ch := range b.Content.Chapters {
		for _, elem := range ch.Elements {
			total += elem.WordCount()
		}
	}
	return total
}
