package epub

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/vpoluyaktov/biblio-ebook-parser/parser"
)

// Parser implements the parser.Parser interface for EPUB files
type Parser struct{}

// NewParser creates a new EPUB parser
func NewParser() *Parser {
	return &Parser{}
}

func init() {
	// Register fast extraction functions
	parser.RegisterEPUBExtractors(
		ExtractCoverOnly,
		ExtractCoverOnlyReader,
		ExtractAnnotationOnly,
		ExtractAnnotationOnlyReader,
		ExtractMetadataOnly,
		ExtractMetadataOnlyReader,
	)
}

// Format returns the format identifier
func (p *Parser) Format() string {
	return "epub"
}

// Parse extracts book structure from an EPUB file
func (p *Parser) Parse(filePath string) (*parser.Book, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer r.Close()

	return p.parseFromZip(&r.Reader)
}

// ParseReader extracts book structure from an io.ReaderAt
func (p *Parser) ParseReader(r io.ReaderAt, size int64) (*parser.Book, error) {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB as zip: %w", err)
	}

	return p.parseFromZip(zipReader)
}

func (p *Parser) parseFromZip(zr *zip.Reader) (*parser.Book, error) {
	// Find and parse container.xml
	containerFile, err := findFileInZip(zr, "META-INF/container.xml")
	if err != nil {
		return nil, fmt.Errorf("container.xml not found: %w", err)
	}

	var container epubContainer
	if err := parseXMLFromZipFile(containerFile, &container); err != nil {
		return nil, fmt.Errorf("failed to parse container.xml: %w", err)
	}

	// Find and parse the package file (content.opf)
	packageFile, err := findFileInZip(zr, container.RootFile.FullPath)
	if err != nil {
		return nil, fmt.Errorf("package file not found: %w", err)
	}

	var pkg epubPackage
	if err := parseXMLFromZipFile(packageFile, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package file: %w", err)
	}

	book := &parser.Book{}

	// Extract metadata
	book.Metadata = extractMetadata(pkg, container.RootFile.FullPath, zr)

	// Extract content
	baseDir := filepath.Dir(container.RootFile.FullPath)
	book.Content = extractContent(zr, baseDir, pkg)

	return book, nil
}

func extractMetadata(pkg epubPackage, rootFilePath string, zr *zip.Reader) parser.Metadata {
	metadata := parser.Metadata{}

	// Title
	if len(pkg.Metadata.Titles) > 0 {
		metadata.Title = strings.TrimSpace(pkg.Metadata.Titles[0])
	}

	// Authors
	metadata.Authors = parseAuthors(pkg.Metadata.Creators)

	// Language
	if len(pkg.Metadata.Languages) > 0 {
		lang := strings.TrimSpace(pkg.Metadata.Languages[0])
		if len(lang) > 2 {
			lang = lang[:2]
		}
		metadata.Language = lang
	}

	// Description
	metadata.Description = strings.TrimSpace(pkg.Metadata.Description)
	if metadata.Description == "" && len(pkg.Metadata.Subjects) > 0 {
		metadata.Description = strings.Join(pkg.Metadata.Subjects, ", ")
	}

	// Series and genres from Calibre metadata
	for _, meta := range pkg.Metadata.Metas {
		switch meta.Name {
		case "calibre:series":
			metadata.Series = strings.TrimSpace(meta.Content)
		case "calibre:series_index":
			fmt.Sscanf(meta.Content, "%d", &metadata.SeriesIndex)
		}
	}

	// Genres from subjects
	metadata.Genres = pkg.Metadata.Subjects

	// Extract cover image
	baseDir := filepath.Dir(rootFilePath)
	coverHref := extractCoverHref(pkg, baseDir)
	if coverHref != "" {
		coverFile, err := findFileInZip(zr, coverHref)
		if err == nil {
			rc, err := coverFile.Open()
			if err == nil {
				defer rc.Close()
				coverData, err := io.ReadAll(rc)
				if err == nil {
					metadata.CoverData = coverData
					if strings.HasSuffix(strings.ToLower(coverHref), ".png") {
						metadata.CoverType = "image/png"
					} else {
						metadata.CoverType = "image/jpeg"
					}
				}
			}
		}
	}

	return metadata
}

func parseAuthors(creators []epubCreator) []parser.Author {
	var authors []parser.Author

	for _, creator := range creators {
		// Skip if not an author (role might be editor, illustrator, etc.)
		if creator.Role != "" && creator.Role != "aut" {
			continue
		}

		name := strings.TrimSpace(creator.Name)
		if name == "" {
			continue
		}

		author := parser.Author{}

		// Try to parse "LastName, FirstName" format
		if strings.Contains(name, ",") {
			parts := strings.SplitN(name, ",", 2)
			author.LastName = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				// FirstName might contain middle name
				nameParts := strings.Fields(strings.TrimSpace(parts[1]))
				if len(nameParts) > 0 {
					author.FirstName = nameParts[0]
				}
				if len(nameParts) > 1 {
					author.MiddleName = strings.Join(nameParts[1:], " ")
				}
			}
		} else {
			// Try to parse "FirstName LastName" format
			nameParts := strings.Fields(name)
			if len(nameParts) == 1 {
				author.LastName = nameParts[0]
			} else if len(nameParts) == 2 {
				author.FirstName = nameParts[0]
				author.LastName = nameParts[1]
			} else if len(nameParts) > 2 {
				author.FirstName = nameParts[0]
				author.MiddleName = strings.Join(nameParts[1:len(nameParts)-1], " ")
				author.LastName = nameParts[len(nameParts)-1]
			}
		}

		if !author.IsEmpty() {
			authors = append(authors, author)
		}
	}

	return authors
}

func extractCoverHref(pkg epubPackage, baseDir string) string {
	// Look for items that might be cover images
	for _, item := range pkg.Manifest.Items {
		id := strings.ToLower(item.ID)
		href := strings.ToLower(item.Href)
		if (strings.Contains(id, "cover") || strings.Contains(href, "cover")) &&
			(item.MediaType == "image/jpeg" || item.MediaType == "image/png" ||
				item.MediaType == "image/jpg") {
			return filepath.Join(baseDir, item.Href)
		}
	}

	return ""
}

func findFileInZip(zr *zip.Reader, name string) (*zip.File, error) {
	for _, f := range zr.File {
		if f.Name == name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

func parseXMLFromZipFile(f *zip.File, v interface{}) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	return xml.Unmarshal(data, v)
}

// XML structures for EPUB parsing

type epubContainer struct {
	XMLName  xml.Name `xml:"container"`
	RootFile struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

type epubPackage struct {
	XMLName  xml.Name     `xml:"package"`
	Metadata epubMetadata `xml:"metadata"`
	Manifest struct {
		Items []epubManifestItem `xml:"item"`
	} `xml:"manifest"`
	Spine struct {
		TOC      string `xml:"toc,attr"`
		ItemRefs []struct {
			IDRef string `xml:"idref,attr"`
		} `xml:"itemref"`
	} `xml:"spine"`
}

type epubMetadata struct {
	Titles      []string      `xml:"title"`
	Creators    []epubCreator `xml:"creator"`
	Languages   []string      `xml:"language"`
	Subjects    []string      `xml:"subject"`
	Description string        `xml:"description"`
	Metas       []epubMeta    `xml:"meta"`
}

type epubCreator struct {
	Name   string `xml:",chardata"`
	FileAs string `xml:"file-as,attr"`
	Role   string `xml:"role,attr"`
}

type epubMeta struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

type epubManifestItem struct {
	ID        string `xml:"id,attr"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}

type epubTOCEntry struct {
	Title  string
	Path   string
	Anchor string
}
