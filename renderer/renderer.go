package renderer

import "github.com/vpoluyaktov/biblio-ebook-parser/parser"

// Renderer converts a parsed Book into a specific output format
type Renderer interface {
	// RenderMetadata converts book metadata to the target format
	RenderMetadata(book *parser.Book) (interface{}, error)

	// RenderContent converts book content to the target format
	RenderContent(book *parser.Book) (interface{}, error)
}
