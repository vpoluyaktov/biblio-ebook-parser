package parser

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// ExtractCover extracts only the cover image from an ebook file without parsing the full content.
// This is much faster than Parse() when you only need the cover.
// Supported formats: EPUB, FB2
func ExtractCover(filePath string) ([]byte, string, error) {
	format := detectFormat(filePath)

	// Use format-specific fast extraction
	switch format {
	case "epub":
		if extractEPUBCover == nil {
			return nil, "", fmt.Errorf("EPUB extractor not registered")
		}
		return extractEPUBCover(filePath)
	case "fb2":
		if extractFB2Cover == nil {
			return nil, "", fmt.Errorf("FB2 extractor not registered")
		}
		return extractFB2Cover(filePath)
	default:
		return nil, "", fmt.Errorf("cover extraction not supported for format: %s", format)
	}
}

// ExtractCoverReader extracts only the cover image from an ebook reader without parsing the full content.
func ExtractCoverReader(r io.ReaderAt, size int64, format string) ([]byte, string, error) {
	// Use format-specific fast extraction
	switch format {
	case "epub":
		if extractEPUBCoverReader == nil {
			return nil, "", fmt.Errorf("EPUB extractor not registered")
		}
		return extractEPUBCoverReader(r, size)
	case "fb2":
		if extractFB2CoverReader == nil {
			return nil, "", fmt.Errorf("FB2 extractor not registered")
		}
		return extractFB2CoverReader(r, size)
	default:
		return nil, "", fmt.Errorf("cover extraction not supported for format: %s", format)
	}
}

// ExtractAnnotation extracts only the description/annotation from an ebook file without parsing the full content.
func ExtractAnnotation(filePath string) (string, error) {
	format := detectFormat(filePath)

	// Use format-specific fast extraction
	switch format {
	case "epub":
		if extractEPUBAnnotation == nil {
			return "", fmt.Errorf("EPUB extractor not registered")
		}
		return extractEPUBAnnotation(filePath)
	case "fb2":
		if extractFB2Annotation == nil {
			return "", fmt.Errorf("FB2 extractor not registered")
		}
		return extractFB2Annotation(filePath)
	default:
		return "", fmt.Errorf("annotation extraction not supported for format: %s", format)
	}
}

// ExtractAnnotationReader extracts only the description/annotation from an ebook reader without parsing the full content.
func ExtractAnnotationReader(r io.ReaderAt, size int64, format string) (string, error) {
	// Use format-specific fast extraction
	switch format {
	case "epub":
		if extractEPUBAnnotationReader == nil {
			return "", fmt.Errorf("EPUB extractor not registered")
		}
		return extractEPUBAnnotationReader(r, size)
	case "fb2":
		if extractFB2AnnotationReader == nil {
			return "", fmt.Errorf("FB2 extractor not registered")
		}
		return extractFB2AnnotationReader(r, size)
	default:
		return "", fmt.Errorf("annotation extraction not supported for format: %s", format)
	}
}

// ExtractMetadata extracts only metadata from an ebook file without parsing the full content.
func ExtractMetadata(filePath string) (Metadata, error) {
	format := detectFormat(filePath)

	// Use format-specific fast extraction
	switch format {
	case "epub":
		if extractEPUBMetadata == nil {
			return Metadata{}, fmt.Errorf("EPUB extractor not registered")
		}
		return extractEPUBMetadata(filePath)
	case "fb2":
		if extractFB2Metadata == nil {
			return Metadata{}, fmt.Errorf("FB2 extractor not registered")
		}
		return extractFB2Metadata(filePath)
	default:
		return Metadata{}, fmt.Errorf("metadata extraction not supported for format: %s", format)
	}
}

// ExtractMetadataReader extracts only metadata from an ebook reader without parsing the full content.
func ExtractMetadataReader(r io.ReaderAt, size int64, format string) (Metadata, error) {
	// Use format-specific fast extraction
	switch format {
	case "epub":
		if extractEPUBMetadataReader == nil {
			return Metadata{}, fmt.Errorf("EPUB extractor not registered")
		}
		return extractEPUBMetadataReader(r, size)
	case "fb2":
		if extractFB2MetadataReader == nil {
			return Metadata{}, fmt.Errorf("FB2 extractor not registered")
		}
		return extractFB2MetadataReader(r, size)
	default:
		return Metadata{}, fmt.Errorf("metadata extraction not supported for format: %s", format)
	}
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

// These functions will be implemented by format-specific packages
var (
	extractEPUBCover            func(string) ([]byte, string, error)
	extractEPUBCoverReader      func(io.ReaderAt, int64) ([]byte, string, error)
	extractEPUBAnnotation       func(string) (string, error)
	extractEPUBAnnotationReader func(io.ReaderAt, int64) (string, error)
	extractEPUBMetadata         func(string) (Metadata, error)
	extractEPUBMetadataReader   func(io.ReaderAt, int64) (Metadata, error)

	extractFB2Cover            func(string) ([]byte, string, error)
	extractFB2CoverReader      func(io.ReaderAt, int64) ([]byte, string, error)
	extractFB2Annotation       func(string) (string, error)
	extractFB2AnnotationReader func(io.ReaderAt, int64) (string, error)
	extractFB2Metadata         func(string) (Metadata, error)
	extractFB2MetadataReader   func(io.ReaderAt, int64) (Metadata, error)
)

// RegisterEPUBExtractors registers EPUB-specific extraction functions
func RegisterEPUBExtractors(
	cover func(string) ([]byte, string, error),
	coverReader func(io.ReaderAt, int64) ([]byte, string, error),
	annotation func(string) (string, error),
	annotationReader func(io.ReaderAt, int64) (string, error),
	metadata func(string) (Metadata, error),
	metadataReader func(io.ReaderAt, int64) (Metadata, error),
) {
	extractEPUBCover = cover
	extractEPUBCoverReader = coverReader
	extractEPUBAnnotation = annotation
	extractEPUBAnnotationReader = annotationReader
	extractEPUBMetadata = metadata
	extractEPUBMetadataReader = metadataReader
}

// RegisterFB2Extractors registers FB2-specific extraction functions
func RegisterFB2Extractors(
	cover func(string) ([]byte, string, error),
	coverReader func(io.ReaderAt, int64) ([]byte, string, error),
	annotation func(string) (string, error),
	annotationReader func(io.ReaderAt, int64) (string, error),
	metadata func(string) (Metadata, error),
	metadataReader func(io.ReaderAt, int64) (Metadata, error),
) {
	extractFB2Cover = cover
	extractFB2CoverReader = coverReader
	extractFB2Annotation = annotation
	extractFB2AnnotationReader = annotationReader
	extractFB2Metadata = metadata
	extractFB2MetadataReader = metadataReader
}
