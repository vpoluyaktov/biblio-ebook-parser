package parser

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
)

// FastExtractor defines the interface for fast metadata/cover/annotation extraction
// without parsing the full book content.
//
// This interface allows format-specific implementations to provide optimized extraction
// methods that read only the necessary parts of the ebook file, making them much faster
// than full parsing when you only need specific metadata.
type FastExtractor interface {
	ExtractCoverFromFile(filePath string) ([]byte, string, error)
	ExtractCoverFromReader(r io.ReaderAt, size int64) ([]byte, string, error)
	ExtractAnnotationFromFile(filePath string) (string, error)
	ExtractAnnotationFromReader(r io.ReaderAt, size int64) (string, error)
	ExtractMetadataFromFile(filePath string) (Metadata, error)
	ExtractMetadataFromReader(r io.ReaderAt, size int64) (Metadata, error)
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

// ExtractCoverFromFile extracts only the cover image from an ebook file without parsing the full content.
// This is much faster than Parse() when you only need the cover.
// Supported formats: EPUB, FB2
func ExtractCoverFromFile(filePath string) ([]byte, string, error) {
	format := detectFormat(filePath)
	extractor, err := getExtractor(format)
	if err != nil {
		return nil, "", err
	}
	return extractor.ExtractCoverFromFile(filePath)
}

// ExtractCoverFromReader extracts only the cover image from an ebook reader without parsing the full content.
func ExtractCoverFromReader(r io.ReaderAt, size int64, format string) ([]byte, string, error) {
	extractor, err := getExtractor(format)
	if err != nil {
		return nil, "", err
	}
	return extractor.ExtractCoverFromReader(r, size)
}

// ExtractAnnotationFromFile extracts only the description/annotation from an ebook file without parsing the full content.
func ExtractAnnotationFromFile(filePath string) (string, error) {
	format := detectFormat(filePath)
	extractor, err := getExtractor(format)
	if err != nil {
		return "", err
	}
	return extractor.ExtractAnnotationFromFile(filePath)
}

// ExtractAnnotationFromReader extracts only the description/annotation from an ebook reader without parsing the full content.
func ExtractAnnotationFromReader(r io.ReaderAt, size int64, format string) (string, error) {
	extractor, err := getExtractor(format)
	if err != nil {
		return "", err
	}
	return extractor.ExtractAnnotationFromReader(r, size)
}

// ExtractMetadataFromFile extracts only metadata from an ebook file without parsing the full content.
func ExtractMetadataFromFile(filePath string) (Metadata, error) {
	format := detectFormat(filePath)
	extractor, err := getExtractor(format)
	if err != nil {
		return Metadata{}, err
	}
	return extractor.ExtractMetadataFromFile(filePath)
}

// ExtractMetadataFromReader extracts only metadata from an ebook reader without parsing the full content.
func ExtractMetadataFromReader(r io.ReaderAt, size int64, format string) (Metadata, error) {
	extractor, err := getExtractor(format)
	if err != nil {
		return Metadata{}, err
	}
	return extractor.ExtractMetadataFromReader(r, size)
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
