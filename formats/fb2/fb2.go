package fb2

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/vpoluyaktov/biblio-ebook-parser/parser"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/unicode"
)

// Parser implements the parser.Parser interface for FB2 files
type Parser struct {
	TOCMaxDepth int
	ParseNotes  bool
}

// NewParser creates a new FB2 parser
func NewParser() *Parser {
	return &Parser{
		TOCMaxDepth: 3,
		ParseNotes:  false,
	}
}

// Format returns the format identifier
func (p *Parser) Format() string {
	return "fb2"
}

// Parse extracts book structure from an FB2 file
func (p *Parser) Parse(filePath string) (*parser.Book, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read FB2: %w", err)
	}

	return p.parseFromBytes(data)
}

// ParseReader extracts book structure from an io.ReaderAt
func (p *Parser) ParseReader(r io.ReaderAt, size int64) (*parser.Book, error) {
	data := make([]byte, size)
	_, err := r.ReadAt(data, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read FB2: %w", err)
	}

	return p.parseFromBytes(data)
}

func (p *Parser) parseFromBytes(data []byte) (*parser.Book, error) {
	// Check if it's a ZIP file (FB2.ZIP)
	if len(data) > 4 && bytes.Equal(data[0:4], []byte{0x50, 0x4B, 0x03, 0x04}) {
		return p.parseFromZip(data)
	}

	// Parse FB2 XML - try with original data first to preserve charset
	var fb2 fb2Document
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charsetReader
	decoder.Strict = false

	if err := decoder.Decode(&fb2); err != nil {
		// If that fails, try with sanitized data
		sanitizedData := sanitizeFB2XML(data)
		decoder2 := xml.NewDecoder(bytes.NewReader(sanitizedData))
		decoder2.CharsetReader = charsetReader
		decoder2.Strict = false

		if err2 := decoder2.Decode(&fb2); err2 != nil {
			return nil, fmt.Errorf("failed to parse FB2: %w", err)
		}
	}

	book := &parser.Book{}

	// Extract metadata
	book.Metadata = extractMetadata(fb2)

	// Extract content
	book.Content = p.extractContent(fb2)

	return book, nil
}

func (p *Parser) parseFromZip(data []byte) (*parser.Book, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP: %w", err)
	}

	var fb2File *zip.File
	for _, f := range zipReader.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".fb2") {
			fb2File = f
			break
		}
	}

	if fb2File == nil {
		return nil, fmt.Errorf("no FB2 file found in archive")
	}

	rc, err := fb2File.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open FB2 file: %w", err)
	}
	defer rc.Close()

	fb2Data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read FB2 file: %w", err)
	}

	return p.parseFromBytes(fb2Data)
}

func extractMetadata(fb2 fb2Document) parser.Metadata {
	metadata := parser.Metadata{}

	metadata.Title = strings.TrimSpace(fb2.Description.TitleInfo.BookTitle)
	metadata.Language = strings.TrimSpace(fb2.Description.TitleInfo.Lang)

	// Description from annotation
	annotation := strings.Join(fb2.Description.TitleInfo.Annotation.Paragraphs, "\n\n")
	metadata.Description = strings.TrimSpace(annotation)

	// Series
	metadata.Series = strings.TrimSpace(fb2.Description.TitleInfo.Sequence.Name)
	metadata.SeriesIndex = parseSeriesNumber(fb2.Description.TitleInfo.Sequence.Number)

	// Genres
	metadata.Genres = fb2.Description.TitleInfo.Genres

	// Author
	author := parser.Author{
		FirstName:  strings.TrimSpace(fb2.Description.TitleInfo.Author.FirstName),
		LastName:   strings.TrimSpace(fb2.Description.TitleInfo.Author.LastName),
		MiddleName: strings.TrimSpace(fb2.Description.TitleInfo.Author.MiddleName),
	}
	if !author.IsEmpty() {
		metadata.Authors = []parser.Author{author}
	}

	// Cover image
	var coverID string
	for _, img := range fb2.Description.TitleInfo.Coverpage.Images {
		href := img.Href
		if href == "" {
			href = img.XlinkHref
		}
		if href == "" {
			href = img.LHref
		}
		if href != "" {
			coverID = strings.TrimPrefix(href, "#")
			break
		}
	}

	if coverID != "" {
		for _, binary := range fb2.Binaries {
			if binary.ID == coverID {
				decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(binary.Data))
				if err == nil {
					metadata.CoverData = decoded
					metadata.CoverType = binary.ContentType
					if metadata.CoverType == "" {
						if bytes.HasPrefix(decoded, []byte{0xFF, 0xD8, 0xFF}) {
							metadata.CoverType = "image/jpeg"
						} else if bytes.HasPrefix(decoded, []byte{0x89, 0x50, 0x4E, 0x47}) {
							metadata.CoverType = "image/png"
						} else {
							metadata.CoverType = "image/jpeg"
						}
					}
				}
				break
			}
		}
	}

	return metadata
}

func (p *Parser) extractContent(fb2 fb2Document) parser.Content {
	content := parser.Content{
		Chapters: []parser.Chapter{},
	}

	chapterNum := 1
	for _, body := range fb2.Bodies {
		// Skip notes and comments unless configured
		if body.Name == "notes" || body.Name == "comments" {
			if !p.ParseNotes {
				continue
			}
		}

		// Add body title as chapter if present
		if body.Title.Content != "" {
			titleText := fb2XMLToText(body.Title.Content)
			elements := []parser.Element{
				&parser.Heading{Text: titleText, Level: 1},
			}
			content.Chapters = append(content.Chapters, parser.Chapter{
				ID:       fmt.Sprintf("body-title-%d", chapterNum),
				Title:    titleText,
				Level:    0,
				Elements: elements,
			})
			chapterNum++
		}

		// Process sections
		for _, section := range body.Sections {
			p.addSections(&content, section, 0, &chapterNum)
		}
	}

	return content
}

func (p *Parser) addSections(content *parser.Content, section fb2Section, depth int, chapterNum *int) {
	depth++
	if depth > p.TOCMaxDepth {
		return
	}

	title := fb2XMLToText(section.Title.Content)
	if title == "" {
		title = fmt.Sprintf("Chapter %d", *chapterNum)
	}

	elements := sectionToElements(section)

	// Only add if has content or no nested sections
	hasNestedSections := len(section.Sections) > 0
	hasContent := len(elements) > 0

	if hasContent || !hasNestedSections {
		content.Chapters = append(content.Chapters, parser.Chapter{
			ID:       fmt.Sprintf("section-%d", *chapterNum),
			Title:    strings.TrimSpace(title),
			Level:    depth - 1,
			Elements: elements,
		})
		*chapterNum++
	}

	// Process nested sections
	for _, subsection := range section.Sections {
		p.addSections(content, subsection, depth, chapterNum)
	}
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	charset = strings.ToLower(charset)

	switch charset {
	case "windows-1251":
		return charmap.Windows1251.NewDecoder().Reader(input), nil
	case "windows-1252":
		return charmap.Windows1252.NewDecoder().Reader(input), nil
	case "iso-8859-1", "latin1":
		return charmap.ISO8859_1.NewDecoder().Reader(input), nil
	case "koi8-r":
		return charmap.KOI8R.NewDecoder().Reader(input), nil
	case "koi8-u":
		return charmap.KOI8U.NewDecoder().Reader(input), nil
	case "utf-8", "":
		return input, nil
	case "utf-16", "utf-16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder().Reader(input), nil
	case "utf-16be":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder().Reader(input), nil
	default:
		enc, err := ianaindex.IANA.Encoding(charset)
		if err != nil {
			return input, nil
		}
		if enc == nil {
			return input, nil
		}
		return enc.NewDecoder().Reader(input), nil
	}
}

func parseSeriesNumber(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	if n, err := strconv.Atoi(s); err == nil {
		if n > 0 {
			return n
		}
		return 1
	}

	return 1
}

// XML structures for FB2 parsing

type fb2Document struct {
	XMLName     xml.Name `xml:"FictionBook"`
	Description struct {
		TitleInfo struct {
			Author struct {
				FirstName  string `xml:"first-name"`
				LastName   string `xml:"last-name"`
				MiddleName string `xml:"middle-name"`
			} `xml:"author"`
			BookTitle  string   `xml:"book-title"`
			Genres     []string `xml:"genre"`
			Lang       string   `xml:"lang"`
			Annotation struct {
				Paragraphs []string `xml:"p"`
			} `xml:"annotation"`
			Sequence struct {
				Name   string `xml:"name,attr"`
				Number string `xml:"number,attr"`
			} `xml:"sequence"`
			Coverpage struct {
				Images []fb2Image `xml:"image"`
			} `xml:"coverpage"`
		} `xml:"title-info"`
	} `xml:"description"`
	Bodies   []fb2Body   `xml:"body"`
	Binaries []fb2Binary `xml:"binary"`
}

type fb2Body struct {
	Name     string       `xml:"name,attr"`
	Title    fb2Title     `xml:"title"`
	Sections []fb2Section `xml:"section"`
}

type fb2Section struct {
	Title      fb2Title      `xml:"title"`
	Paragraphs []fb2Para     `xml:"p"`
	Epigraphs  []fb2Epigraph `xml:"epigraph"`
	Sections   []fb2Section  `xml:"section"`
}

type fb2Title struct {
	Content string `xml:",innerxml"`
}

type fb2Para struct {
	Content string `xml:",innerxml"`
}

type fb2Epigraph struct {
	Paragraphs []fb2Para `xml:"p"`
}

type fb2Image struct {
	Href      string `xml:"href,attr"`
	XlinkHref string `xml:"http://www.w3.org/1999/xlink href,attr"`
	LHref     string `xml:"http://www.gribuser.ru/xml/fictionbook/2.0 href,attr"`
}

type fb2Binary struct {
	ID          string `xml:"id,attr"`
	ContentType string `xml:"content-type,attr"`
	Data        string `xml:",chardata"`
}
