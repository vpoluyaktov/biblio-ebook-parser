package plaintext

import (
	"strings"

	"github.com/vpoluyaktov/biblio-ebook-parser/parser"
)

// Renderer converts parsed books to plain text format for TTS
type Renderer struct {
	Config Config
}

// Config holds configuration for plain text rendering
type Config struct {
	AddPeriods    bool // Add periods to paragraphs that don't end with punctuation
	InsertMarkers bool // Insert SSML markers for TTS pauses
	NormalizeText bool // Normalize text for speech synthesis
}

// NewRenderer creates a new plain text renderer
func NewRenderer(config Config) *Renderer {
	return &Renderer{Config: config}
}

// Book represents plain text book content for TTS
type Book struct {
	Title        string
	Author       string
	Series       string
	SeriesNumber string
	Description  string
	Chapters     []Chapter
	Metadata     map[string]string
}

// Chapter represents a plain text chapter
type Chapter struct {
	Title    string
	Content  string
	ID       string
	TOCDepth int
}

// RenderMetadata converts book metadata to a simple map
func (r *Renderer) RenderMetadata(book *parser.Book) (interface{}, error) {
	metadata := map[string]string{
		"title":       book.Metadata.Title,
		"language":    book.Metadata.Language,
		"description": book.Metadata.Description,
	}

	if len(book.Metadata.Authors) > 0 {
		metadata["author"] = book.Metadata.Authors[0].FullName()
	}

	if book.Metadata.Series != "" {
		metadata["series"] = book.Metadata.Series
	}

	return metadata, nil
}

// RenderContent converts book content to plain text format
func (r *Renderer) RenderContent(book *parser.Book) (interface{}, error) {
	result := &Book{
		Title:       book.Metadata.Title,
		Series:      book.Metadata.Series,
		Description: book.Metadata.Description,
		Chapters:    make([]Chapter, 0, len(book.Content.Chapters)),
		Metadata: map[string]string{
			"description": book.Metadata.Description,
		},
	}

	if len(book.Metadata.Authors) > 0 {
		result.Author = book.Metadata.Authors[0].FullName()
	}

	for _, ch := range book.Content.Chapters {
		plainText := r.elementsToPlainText(ch.Elements)
		
		if r.Config.AddPeriods {
			plainText = addPeriods(plainText)
		}

		result.Chapters = append(result.Chapters, Chapter{
			Title:    ch.Title,
			Content:  plainText,
			ID:       ch.ID,
			TOCDepth: ch.Level,
		})
	}

	return result, nil
}

func (r *Renderer) elementsToPlainText(elements []parser.Element) string {
	var text strings.Builder

	for _, elem := range elements {
		switch e := elem.(type) {
		case *parser.Heading:
			text.WriteString("\n")
			text.WriteString(e.Text)
			if r.Config.InsertMarkers {
				text.WriteString("{{TITLE_BREAK}}")
			}
			text.WriteString("\n\n")

		case *parser.Paragraph:
			text.WriteString(e.Text)
			text.WriteString("\n\n")

		case *parser.Image:
			if e.Alt != "" {
				text.WriteString("[Image: ")
				text.WriteString(e.Alt)
				text.WriteString("]\n\n")
			}

		case *parser.Table:
			if e.Caption != "" {
				text.WriteString("[Table: ")
				text.WriteString(e.Caption)
				text.WriteString("]\n\n")
			} else {
				text.WriteString("[Table]\n\n")
			}

		case *parser.EmptyLine:
			text.WriteString("\n")

		case *parser.Epigraph:
			for _, p := range e.Paragraphs {
				text.WriteString("    ") // Indent epigraphs
				text.WriteString(p.Text)
				text.WriteString("\n\n")
			}
		}
	}

	return strings.TrimSpace(text.String())
}
