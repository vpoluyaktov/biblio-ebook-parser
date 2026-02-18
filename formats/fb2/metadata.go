package fb2

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/vpoluyaktov/biblio-ebook-parser/parser"
)

// ExtractCoverOnly extracts only the cover image from an FB2 file without parsing the full content.
// This is much faster than Parse() when you only need the cover.
func ExtractCoverOnly(filePath string) ([]byte, string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read FB2: %w", err)
	}

	return extractCoverFromBytes(data)
}

// ExtractCoverOnlyReader extracts only the cover image from an FB2 reader without parsing the full content.
func ExtractCoverOnlyReader(r io.ReaderAt, size int64) ([]byte, string, error) {
	data := make([]byte, size)
	_, err := r.ReadAt(data, 0)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read FB2: %w", err)
	}

	return extractCoverFromBytes(data)
}

// ExtractAnnotationOnly extracts only the description/annotation from an FB2 file without parsing the full content.
func ExtractAnnotationOnly(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read FB2: %w", err)
	}

	return extractAnnotationFromBytes(data)
}

// ExtractAnnotationOnlyReader extracts only the description/annotation from an FB2 reader without parsing the full content.
func ExtractAnnotationOnlyReader(r io.ReaderAt, size int64) (string, error) {
	data := make([]byte, size)
	_, err := r.ReadAt(data, 0)
	if err != nil {
		return "", fmt.Errorf("failed to read FB2: %w", err)
	}

	return extractAnnotationFromBytes(data)
}

// ExtractMetadataOnly extracts only metadata from an FB2 file without parsing the full content.
func ExtractMetadataOnly(filePath string) (parser.Metadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return parser.Metadata{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return parser.Metadata{}, fmt.Errorf("failed to read FB2: %w", err)
	}

	return extractMetadataFromBytes(data)
}

// ExtractMetadataOnlyReader extracts only metadata from an FB2 reader without parsing the full content.
func ExtractMetadataOnlyReader(r io.ReaderAt, size int64) (parser.Metadata, error) {
	data := make([]byte, size)
	_, err := r.ReadAt(data, 0)
	if err != nil {
		return parser.Metadata{}, fmt.Errorf("failed to read FB2: %w", err)
	}

	return extractMetadataFromBytes(data)
}

func extractCoverFromBytes(data []byte) ([]byte, string, error) {
	var doc fb2Document
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charsetReader
	decoder.Strict = false

	if err := decoder.Decode(&doc); err != nil {
		return nil, "", fmt.Errorf("failed to parse FB2: %w", err)
	}

	metadata := extractMetadata(doc)
	return metadata.CoverData, metadata.CoverType, nil
}

func extractAnnotationFromBytes(data []byte) (string, error) {
	var doc fb2Document
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charsetReader
	decoder.Strict = false

	if err := decoder.Decode(&doc); err != nil {
		return "", fmt.Errorf("failed to parse FB2: %w", err)
	}

	metadata := extractMetadata(doc)
	return metadata.Description, nil
}

func extractMetadataFromBytes(data []byte) (parser.Metadata, error) {
	var doc fb2Document
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charsetReader
	decoder.Strict = false

	if err := decoder.Decode(&doc); err != nil {
		return parser.Metadata{}, fmt.Errorf("failed to parse FB2: %w", err)
	}

	return extractMetadata(doc), nil
}
