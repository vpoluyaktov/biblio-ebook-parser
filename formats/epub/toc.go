package epub

import (
	"archive/zip"
	"io"
	"path/filepath"
	"regexp"
	"strings"
)

func extractTOCEntries(zr *zip.Reader, packageBaseDir string, manifestMap map[string]string, manifestMediaTypeMap map[string]string, spineTOCID string) []epubTOCEntry {
	tocIDs := make([]string, 0, 4)
	if spineTOCID != "" {
		tocIDs = append(tocIDs, spineTOCID)
	}
	for id, mediaType := range manifestMediaTypeMap {
		if mediaType == "application/x-dtbncx+xml" || (mediaType == "application/xhtml+xml" && strings.Contains(strings.ToLower(id), "nav")) {
			tocIDs = append(tocIDs, id)
		}
	}

	for _, tocID := range tocIDs {
		tocHref, ok := manifestMap[tocID]
		if !ok {
			continue
		}
		tocPath := normalizeEPUBPath(packageBaseDir, tocHref)
		tocFile, err := findFileInZip(zr, tocPath)
		if err != nil {
			continue
		}

		mediaType := manifestMediaTypeMap[tocID]
		tocBaseDir := filepath.Dir(tocPath)
		if mediaType == "application/x-dtbncx+xml" {
			entries, err := parseNCXTOCEntries(tocFile, tocBaseDir)
			if err == nil && len(entries) > 0 {
				return entries
			}
			continue
		}
		if mediaType == "application/xhtml+xml" {
			entries, err := parseNavXHTMLTOCEntries(tocFile, tocBaseDir)
			if err == nil && len(entries) > 0 {
				return entries
			}
		}
	}

	return nil
}

func parseNCXTOCEntries(f *zip.File, tocBaseDir string) ([]epubTOCEntry, error) {
	var ncx struct {
		NavMap struct {
			NavPoints []ncxNavPoint `xml:"navPoint"`
		} `xml:"navMap"`
	}
	if err := parseXMLFromZipFile(f, &ncx); err != nil {
		return nil, err
	}

	entries := make([]epubTOCEntry, 0, len(ncx.NavMap.NavPoints))
	collectNCXTOCEntries(ncx.NavMap.NavPoints, tocBaseDir, &entries)
	return entries, nil
}

type ncxNavPoint struct {
	NavLabel struct {
		Text string `xml:"text"`
	} `xml:"navLabel"`
	Content struct {
		Src string `xml:"src,attr"`
	} `xml:"content"`
	NavPoints []ncxNavPoint `xml:"navPoint"`
}

func collectNCXTOCEntries(points []ncxNavPoint, tocBaseDir string, out *[]epubTOCEntry) {
	for _, point := range points {
		title := strings.TrimSpace(stripHTMLTags(point.NavLabel.Text))
		src := strings.TrimSpace(point.Content.Src)
		if title != "" && src != "" {
			filePath, anchor := splitEPUBHref(src)
			*out = append(*out, epubTOCEntry{
				Title:  title,
				Path:   normalizeEPUBPath(tocBaseDir, filePath),
				Anchor: anchor,
			})
		}
		if len(point.NavPoints) > 0 {
			collectNCXTOCEntries(point.NavPoints, tocBaseDir, out)
		}
	}
}

func parseNavXHTMLTOCEntries(f *zip.File, tocBaseDir string) ([]epubTOCEntry, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	// Lenient fallback parser for nav.xhtml when XML namespaces are inconsistent
	re := regexp.MustCompile(`(?is)<a[^>]*href\s*=\s*"([^"]+)"[^>]*>(.*?)</a>`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	entries := make([]epubTOCEntry, 0, len(matches))
	for _, m := range matches {
		href := strings.TrimSpace(m[1])
		title := strings.TrimSpace(stripHTMLTags(m[2]))
		if href == "" || title == "" {
			continue
		}
		filePath, anchor := splitEPUBHref(href)
		entries = append(entries, epubTOCEntry{
			Title:  title,
			Path:   normalizeEPUBPath(tocBaseDir, filePath),
			Anchor: anchor,
		})
	}

	return entries, nil
}

func splitEPUBHref(href string) (string, string) {
	href = strings.TrimSpace(href)
	if href == "" {
		return "", ""
	}
	parts := strings.SplitN(href, "#", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.TrimSpace(parts[1])
}
