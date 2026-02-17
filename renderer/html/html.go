package html

import (
	"fmt"
	"strings"

	"github.com/vpoluyaktov/biblio-ebook-parser/parser"
)

// Renderer converts parsed books to HTML format for web readers
type Renderer struct {
	Config Config
}

// Config holds configuration for HTML rendering
type Config struct {
	PreserveStructure bool // Preserve HTML structure from original
}

// NewRenderer creates a new HTML renderer
func NewRenderer(config Config) *Renderer {
	return &Renderer{Config: config}
}

// BookContent represents HTML-formatted book content for web readers
type BookContent struct {
	Title    string    `json:"title"`
	Author   string    `json:"author"`
	Format   string    `json:"format"`
	Chapters []Chapter `json:"chapters"`
}

// Chapter represents an HTML chapter
type Chapter struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// RenderMetadata converts book metadata to a simple map
func (r *Renderer) RenderMetadata(book *parser.Book) (interface{}, error) {
	metadata := map[string]interface{}{
		"title":       book.Metadata.Title,
		"language":    book.Metadata.Language,
		"description": book.Metadata.Description,
		"genres":      book.Metadata.Genres,
		"series":      book.Metadata.Series,
		"seriesIndex": book.Metadata.SeriesIndex,
	}

	if len(book.Metadata.Authors) > 0 {
		authors := make([]string, len(book.Metadata.Authors))
		for i, author := range book.Metadata.Authors {
			authors[i] = author.FullName()
		}
		metadata["authors"] = authors
	}

	if book.Metadata.CoverData != nil {
		metadata["hasCover"] = true
		metadata["coverType"] = book.Metadata.CoverType
	}

	return metadata, nil
}

// RenderContent converts book content to HTML format
func (r *Renderer) RenderContent(book *parser.Book) (interface{}, error) {
	content := &BookContent{
		Title:    book.Metadata.Title,
		Format:   "html",
		Chapters: make([]Chapter, 0, len(book.Content.Chapters)),
	}

	if len(book.Metadata.Authors) > 0 {
		content.Author = book.Metadata.Authors[0].FullName()
	}

	for _, ch := range book.Content.Chapters {
		htmlContent := r.elementsToHTML(ch.Elements)
		content.Chapters = append(content.Chapters, Chapter{
			ID:      ch.ID,
			Title:   ch.Title,
			Content: htmlContent,
		})
	}

	return content, nil
}

func (r *Renderer) elementsToHTML(elements []parser.Element) string {
	var html strings.Builder

	for _, elem := range elements {
		switch e := elem.(type) {
		case *parser.Heading:
			level := e.Level
			if level < 1 {
				level = 1
			}
			if level > 6 {
				level = 6
			}
			html.WriteString(fmt.Sprintf("<h%d>%s</h%d>\n", level, htmlEscape(e.Text), level))

		case *parser.Paragraph:
			if r.Config.PreserveStructure && e.HTML != "" {
				html.WriteString(e.HTML)
				html.WriteString("\n")
			} else {
				html.WriteString("<p>")
				html.WriteString(htmlEscape(e.Text))
				html.WriteString("</p>\n")
			}

		case *parser.Image:
			alt := htmlEscape(e.Alt)
			if e.Href != "" {
				html.WriteString(fmt.Sprintf(`<img src="%s" alt="%s">`, htmlEscape(e.Href), alt))
			} else {
				html.WriteString(fmt.Sprintf(`<p><em>[Image: %s]</em></p>`, alt))
			}
			html.WriteString("\n")

		case *parser.Table:
			caption := htmlEscape(e.Caption)
			if caption != "" {
				html.WriteString(fmt.Sprintf("<p><em>[Table: %s]</em></p>\n", caption))
			} else {
				html.WriteString("<p><em>[Table]</em></p>\n")
			}

		case *parser.EmptyLine:
			html.WriteString("<br/>\n")

		case *parser.Epigraph:
			html.WriteString(`<blockquote class="epigraph">`)
			html.WriteString("\n")
			for _, p := range e.Paragraphs {
				html.WriteString("<p>")
				html.WriteString(htmlEscape(p.Text))
				html.WriteString("</p>\n")
			}
			html.WriteString("</blockquote>\n")
		}
	}

	return html.String()
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
