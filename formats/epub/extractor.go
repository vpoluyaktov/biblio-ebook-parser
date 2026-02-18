package epub

import (
	"io"

	"github.com/vpoluyaktov/biblio-ebook-parser/parser"
)

// Extractor implements the FastExtractor interface for EPUB files
type Extractor struct{}

// ExtractCoverFromFile extracts only the cover image from an EPUB file
func (e *Extractor) ExtractCoverFromFile(filePath string) ([]byte, string, error) {
	return ExtractCoverOnly(filePath)
}

// ExtractCoverFromReader extracts only the cover image from an EPUB reader
func (e *Extractor) ExtractCoverFromReader(r io.ReaderAt, size int64) ([]byte, string, error) {
	return ExtractCoverOnlyReader(r, size)
}

// ExtractAnnotationFromFile extracts only the annotation from an EPUB file
func (e *Extractor) ExtractAnnotationFromFile(filePath string) (string, error) {
	return ExtractAnnotationOnly(filePath)
}

// ExtractAnnotationFromReader extracts only the annotation from an EPUB reader
func (e *Extractor) ExtractAnnotationFromReader(r io.ReaderAt, size int64) (string, error) {
	return ExtractAnnotationOnlyReader(r, size)
}

// ExtractMetadataFromFile extracts only metadata from an EPUB file
func (e *Extractor) ExtractMetadataFromFile(filePath string) (parser.Metadata, error) {
	return ExtractMetadataOnly(filePath)
}

// ExtractMetadataFromReader extracts only metadata from an EPUB reader
func (e *Extractor) ExtractMetadataFromReader(r io.ReaderAt, size int64) (parser.Metadata, error) {
	return ExtractMetadataOnlyReader(r, size)
}
