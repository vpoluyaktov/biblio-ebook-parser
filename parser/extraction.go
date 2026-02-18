package parser

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
)

// FastExtractor defines the interface for fast metadata/cover/annotation extraction
// without parsing the full book content
type FastExtractor interface {
	ExtractCover(filePath string) ([]byte, string, error)
	ExtractCoverReader(r io.ReaderAt, size int64) ([]byte, string, error)
	ExtractAnnotation(filePath string) (string, error)
	ExtractAnnotationReader(r io.ReaderAt, size int64) (string, error)
	ExtractMetadata(filePath string) (Metadata, error)
	ExtractMetadataReader(r io.ReaderAt, size int64) (Metadata, error)
}

var (
	extractors   = make(map[string]FastExtractor)
	extractorsMu sync.RWMutex
)

// RegisterExtractor registers a fast extractor for a specific format
func RegisterExtractor(format string, extractor FastExtractor) {
	extractorsMu.Lock()
	defer extractorsMu.Unlock()
	extractors[format] = extractor
}

// getExtractor returns the extractor for a given format
func getExtractor(format string) (FastExtractor, error) {
	extractorsMu.RLock()
	defer extractorsMu.RUnlock()

	extractor, ok := extractors[format]
	if !ok {
		return nil, fmt.Errorf("no extractor registered for format: %s", format)
	}
	return extractor, nil
}

// ExtractCover extracts only the cover image from an ebook file without parsing the full content.
// This is much faster than Parse() when you only need the cover.
// Supported formats: EPUB, FB2
func ExtractCover(filePath string) ([]byte, string, error) {
	format := detectFormat(filePath)
	extractor, err := getExtractor(format)
	if err != nil {
		return nil, "", err
	}
	return extractor.ExtractCover(filePath)
}

// ExtractCoverReader extracts only the cover image from an ebook reader without parsing the full content.
func ExtractCoverReader(r io.ReaderAt, size int64, format string) ([]byte, string, error) {
	extractor, err := getExtractor(format)
	if err != nil {
		return nil, "", err
	}
	return extractor.ExtractCoverReader(r, size)
}

// ExtractAnnotation extracts only the description/annotation from an ebook file without parsing the full content.
func ExtractAnnotation(filePath string) (string, error) {
	format := detectFormat(filePath)
	extractor, err := getExtractor(format)
	if err != nil {
		return "", err
	}
	return extractor.ExtractAnnotation(filePath)
}

// ExtractAnnotationReader extracts only the description/annotation from an ebook reader without parsing the full content.
func ExtractAnnotationReader(r io.ReaderAt, size int64, format string) (string, error) {
	extractor, err := getExtractor(format)
	if err != nil {
		return "", err
	}
	return extractor.ExtractAnnotationReader(r, size)
}

// ExtractMetadata extracts only metadata from an ebook file without parsing the full content.
func ExtractMetadata(filePath string) (Metadata, error) {
	format := detectFormat(filePath)
	extractor, err := getExtractor(format)
	if err != nil {
		return Metadata{}, err
	}
	return extractor.ExtractMetadata(filePath)
}

// ExtractMetadataReader extracts only metadata from an ebook reader without parsing the full content.
func ExtractMetadataReader(r io.ReaderAt, size int64, format string) (Metadata, error) {
	extractor, err := getExtractor(format)
	if err != nil {
		return Metadata{}, err
	}
	return extractor.ExtractMetadataReader(r, size)
}

// detectFormat detects the ebook format from file extension
func detectFormat(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".epub":
		return "epub"
	case ".fb2":
		return "fb2"
	case ".zip":
		// Could be fb2.zip or epub.zip, need to check
		if strings.HasSuffix(strings.ToLower(filePath), ".fb2.zip") {
			return "fb2"
		} else if strings.HasSuffix(strings.ToLower(filePath), ".epub.zip") {
			return "epub"
		}
		return "unknown"
	default:
		return "unknown"
	}
}
