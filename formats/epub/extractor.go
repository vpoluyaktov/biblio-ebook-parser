package epub

import (
	"io"

	"github.com/vpoluyaktov/biblio-ebook-parser/parser"
)

// Extractor implements the FastExtractor interface for EPUB files
type Extractor struct{}

// ExtractCover extracts only the cover image from an EPUB file
func (e *Extractor) ExtractCover(filePath string) ([]byte, string, error) {
	return ExtractCoverOnly(filePath)
}

// ExtractCoverReader extracts only the cover image from an EPUB reader
func (e *Extractor) ExtractCoverReader(r io.ReaderAt, size int64) ([]byte, string, error) {
	return ExtractCoverOnlyReader(r, size)
}

// ExtractAnnotation extracts only the annotation from an EPUB file
func (e *Extractor) ExtractAnnotation(filePath string) (string, error) {
	return ExtractAnnotationOnly(filePath)
}

// ExtractAnnotationReader extracts only the annotation from an EPUB reader
func (e *Extractor) ExtractAnnotationReader(r io.ReaderAt, size int64) (string, error) {
	return ExtractAnnotationOnlyReader(r, size)
}

// ExtractMetadata extracts only metadata from an EPUB file
func (e *Extractor) ExtractMetadata(filePath string) (parser.Metadata, error) {
	return ExtractMetadataOnly(filePath)
}

// ExtractMetadataReader extracts only metadata from an EPUB reader
func (e *Extractor) ExtractMetadataReader(r io.ReaderAt, size int64) (parser.Metadata, error) {
	return ExtractMetadataOnlyReader(r, size)
}
