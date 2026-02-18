package epub

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/vpoluyaktov/biblio-ebook-parser/parser"
)

// ExtractCoverOnly extracts only the cover image from an EPUB file without parsing the full content.
// This is much faster than Parse() when you only need the cover.
func ExtractCoverOnly(filePath string) ([]byte, string, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer r.Close()

	return extractCoverFromZip(&r.Reader)
}

// ExtractCoverOnlyReader extracts only the cover image from an EPUB reader without parsing the full content.
func ExtractCoverOnlyReader(r io.ReaderAt, size int64) ([]byte, string, error) {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open EPUB as zip: %w", err)
	}

	return extractCoverFromZip(zipReader)
}

// ExtractAnnotationOnly extracts only the description/annotation from an EPUB file without parsing the full content.
func ExtractAnnotationOnly(filePath string) (string, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer r.Close()

	return extractAnnotationFromZip(&r.Reader)
}

// ExtractAnnotationOnlyReader extracts only the description/annotation from an EPUB reader without parsing the full content.
func ExtractAnnotationOnlyReader(r io.ReaderAt, size int64) (string, error) {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return "", fmt.Errorf("failed to open EPUB as zip: %w", err)
	}

	return extractAnnotationFromZip(zipReader)
}

func extractCoverFromZip(zr *zip.Reader) ([]byte, string, error) {
	// Find and parse container.xml
	containerFile, err := findFileInZip(zr, "META-INF/container.xml")
	if err != nil {
		return nil, "", fmt.Errorf("container.xml not found: %w", err)
	}

	var container epubContainer
	if err := parseXMLFromZipFile(containerFile, &container); err != nil {
		return nil, "", fmt.Errorf("failed to parse container.xml: %w", err)
	}

	// Find and parse the package file (content.opf)
	packageFile, err := findFileInZip(zr, container.RootFile.FullPath)
	if err != nil {
		return nil, "", fmt.Errorf("package file not found: %w", err)
	}

	var pkg epubPackage
	if err := parseXMLFromZipFile(packageFile, &pkg); err != nil {
		return nil, "", fmt.Errorf("failed to parse package file: %w", err)
	}

	// Extract cover image
	baseDir := filepath.Dir(container.RootFile.FullPath)
	coverHref := extractCoverHref(pkg, baseDir)
	if coverHref == "" {
		return nil, "", nil
	}

	coverFile, err := findFileInZip(zr, coverHref)
	if err != nil {
		return nil, "", nil
	}

	rc, err := coverFile.Open()
	if err != nil {
		return nil, "", err
	}
	defer rc.Close()

	coverData, err := io.ReadAll(rc)
	if err != nil {
		return nil, "", err
	}

	coverType := "image/jpeg"
	if strings.HasSuffix(strings.ToLower(coverHref), ".png") {
		coverType = "image/png"
	}

	return coverData, coverType, nil
}

func extractAnnotationFromZip(zr *zip.Reader) (string, error) {
	// Find and parse container.xml
	containerFile, err := findFileInZip(zr, "META-INF/container.xml")
	if err != nil {
		return "", fmt.Errorf("container.xml not found: %w", err)
	}

	var container epubContainer
	if err := parseXMLFromZipFile(containerFile, &container); err != nil {
		return "", fmt.Errorf("failed to parse container.xml: %w", err)
	}

	// Find and parse the package file (content.opf)
	packageFile, err := findFileInZip(zr, container.RootFile.FullPath)
	if err != nil {
		return "", fmt.Errorf("package file not found: %w", err)
	}

	var pkg epubPackage
	if err := parseXMLFromZipFile(packageFile, &pkg); err != nil {
		return "", fmt.Errorf("failed to parse package file: %w", err)
	}

	// Return description from metadata
	annotation := strings.TrimSpace(pkg.Metadata.Description)
	if annotation == "" && len(pkg.Metadata.Subjects) > 0 {
		annotation = strings.Join(pkg.Metadata.Subjects, ", ")
	}

	return annotation, nil
}

// ExtractMetadataOnly extracts only metadata from an EPUB file without parsing the full content.
func ExtractMetadataOnly(filePath string) (parser.Metadata, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return parser.Metadata{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return parser.Metadata{}, err
	}

	return ExtractMetadataOnlyReader(f, stat.Size())
}

// ExtractMetadataOnlyReader extracts only metadata from an EPUB reader without parsing the full content.
func ExtractMetadataOnlyReader(r io.ReaderAt, size int64) (parser.Metadata, error) {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return parser.Metadata{}, fmt.Errorf("failed to open EPUB as zip: %w", err)
	}

	// Find and parse container.xml
	containerFile, err := findFileInZip(zipReader, "META-INF/container.xml")
	if err != nil {
		return parser.Metadata{}, fmt.Errorf("container.xml not found: %w", err)
	}

	var container epubContainer
	if err := parseXMLFromZipFile(containerFile, &container); err != nil {
		return parser.Metadata{}, fmt.Errorf("failed to parse container.xml: %w", err)
	}

	// Find and parse the package file (content.opf)
	packageFile, err := findFileInZip(zipReader, container.RootFile.FullPath)
	if err != nil {
		return parser.Metadata{}, fmt.Errorf("package file not found: %w", err)
	}

	var pkg epubPackage
	if err := parseXMLFromZipFile(packageFile, &pkg); err != nil {
		return parser.Metadata{}, fmt.Errorf("failed to parse package file: %w", err)
	}

	return extractMetadata(pkg, container.RootFile.FullPath, zipReader), nil
}
