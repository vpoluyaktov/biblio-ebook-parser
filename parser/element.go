package parser

import "strings"

// ElementType represents the type of content element
type ElementType int

const (
	ElementTypeParagraph ElementType = iota
	ElementTypeHeading
	ElementTypeImage
	ElementTypeTable
	ElementTypeEmptyLine
	ElementTypeEpigraph
)

// Element represents a content building block
type Element interface {
	Type() ElementType
	CharCount() int
	WordCount() int
}

// Paragraph represents a text paragraph
type Paragraph struct {
	Text string
	HTML string // Original HTML if available
}

func (p *Paragraph) Type() ElementType { return ElementTypeParagraph }
func (p *Paragraph) CharCount() int    { return len(p.Text) }
func (p *Paragraph) WordCount() int    { return len(strings.Fields(p.Text)) }

// Heading represents a section heading
type Heading struct {
	Text  string
	Level int // 1-6 for h1-h6
}

func (h *Heading) Type() ElementType { return ElementTypeHeading }
func (h *Heading) CharCount() int    { return len(h.Text) }
func (h *Heading) WordCount() int    { return len(strings.Fields(h.Text)) }

// Image represents an image reference
type Image struct {
	Alt  string
	Href string
	Data []byte // Embedded image data if available
}

func (i *Image) Type() ElementType { return ElementTypeImage }
func (i *Image) CharCount() int    { return 0 }
func (i *Image) WordCount() int    { return 0 }

// Table represents a table (content not parsed, just placeholder)
type Table struct {
	Caption string
}

func (t *Table) Type() ElementType { return ElementTypeTable }
func (t *Table) CharCount() int    { return 0 }
func (t *Table) WordCount() int    { return 0 }

// EmptyLine represents a line break or spacing
type EmptyLine struct{}

func (e *EmptyLine) Type() ElementType { return ElementTypeEmptyLine }
func (e *EmptyLine) CharCount() int    { return 0 }
func (e *EmptyLine) WordCount() int    { return 0 }

// Epigraph represents an epigraph section
type Epigraph struct {
	Paragraphs []Paragraph
}

func (e *Epigraph) Type() ElementType { return ElementTypeEpigraph }
func (e *Epigraph) CharCount() int {
	total := 0
	for _, p := range e.Paragraphs {
		total += p.CharCount()
	}
	return total
}
func (e *Epigraph) WordCount() int {
	total := 0
	for _, p := range e.Paragraphs {
		total += p.WordCount()
	}
	return total
}
