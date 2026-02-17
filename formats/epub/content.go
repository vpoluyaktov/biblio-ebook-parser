package epub

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vpoluyaktov/biblio-ebook-parser/parser"
)

func extractContent(zr *zip.Reader, baseDir string, pkg epubPackage) parser.Content {
	content := parser.Content{
		Chapters: []parser.Chapter{},
	}

	// Build manifest map
	manifestMap := make(map[string]string)
	manifestMediaTypeMap := make(map[string]string)
	for _, item := range pkg.Manifest.Items {
		manifestMap[item.ID] = item.Href
		manifestMediaTypeMap[item.ID] = item.MediaType
	}

	// Try TOC-based extraction first
	tocChapters := extractChaptersFromTOC(zr, baseDir, manifestMap, manifestMediaTypeMap, pkg.Spine.TOC)
	if len(tocChapters) > 0 {
		content.Chapters = tocChapters
		return content
	}

	// Fallback to spine-based extraction
	for i, itemRef := range pkg.Spine.ItemRefs {
		href, ok := manifestMap[itemRef.IDRef]
		if !ok {
			continue
		}

		fullPath := normalizeEPUBPath(baseDir, href)
		chapterFile, err := findFileInZip(zr, fullPath)
		if err != nil {
			continue
		}

		rc, err := chapterFile.Open()
		if err != nil {
			continue
		}

		chapterData, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}

		htmlContent := string(chapterData)
		defaultTitle := fmt.Sprintf("Chapter %d", i+1)
		chapterTitle := extractChapterTitle(htmlContent, defaultTitle)

		elements := htmlToElements(htmlContent)
		content.Chapters = append(content.Chapters, parser.Chapter{
			ID:       itemRef.IDRef,
			Title:    strings.TrimSpace(chapterTitle),
			Level:    0,
			Elements: elements,
		})
	}

	return content
}

func extractChaptersFromTOC(zr *zip.Reader, packageBaseDir string, manifestMap map[string]string, manifestMediaTypeMap map[string]string, spineTOCID string) []parser.Chapter {
	entries := extractTOCEntries(zr, packageBaseDir, manifestMap, manifestMediaTypeMap, spineTOCID)
	if len(entries) == 0 {
		return nil
	}

	htmlCache := make(map[string]string)
	chapters := make([]parser.Chapter, 0, len(entries))

	for i, entry := range entries {
		if entry.Path == "" || strings.TrimSpace(entry.Title) == "" {
			continue
		}

		htmlContent, ok := htmlCache[entry.Path]
		if !ok {
			chapterFile, err := findFileInZip(zr, entry.Path)
			if err != nil {
				continue
			}
			rc, err := chapterFile.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}
			htmlContent = string(data)
			htmlCache[entry.Path] = htmlContent
		}

		start := findAnchorStart(htmlContent, entry.Anchor)
		end := len(htmlContent)
		if i+1 < len(entries) && entries[i+1].Path == entry.Path {
			nextStart := findAnchorStart(htmlContent, entries[i+1].Anchor)
			if nextStart > start {
				end = nextStart
			}
		}
		if start < 0 || start >= len(htmlContent) {
			start = 0
		}
		if end <= start || end > len(htmlContent) {
			end = len(htmlContent)
		}

		segment := strings.TrimSpace(htmlContent[start:end])
		if segment == "" {
			continue
		}

		title := strings.TrimSpace(entry.Title)
		title = extractChapterTitle(segment, title)

		elements := htmlToElements(segment)
		chapters = append(chapters, parser.Chapter{
			ID:       fmt.Sprintf("toc-%d", i+1),
			Title:    title,
			Level:    0,
			Elements: elements,
		})
	}

	return chapters
}

func htmlToElements(htmlContent string) []parser.Element {
	elements := []parser.Element{}

	// Remove head, script, style tags
	reHead := regexp.MustCompile(`(?is)<head[^>]*>.*?</head>`)
	reScript := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	reStyle := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)

	htmlContent = reHead.ReplaceAllString(htmlContent, "")
	htmlContent = reScript.ReplaceAllString(htmlContent, "")
	htmlContent = reStyle.ReplaceAllString(htmlContent, "")

	// Extract headings (match each level separately since Go regexp doesn't support backreferences)
	headingPatterns := []struct {
		pattern *regexp.Regexp
		level   int
	}{
		{regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`), 1},
		{regexp.MustCompile(`(?is)<h2[^>]*>(.*?)</h2>`), 2},
		{regexp.MustCompile(`(?is)<h3[^>]*>(.*?)</h3>`), 3},
		{regexp.MustCompile(`(?is)<h4[^>]*>(.*?)</h4>`), 4},
		{regexp.MustCompile(`(?is)<h5[^>]*>(.*?)</h5>`), 5},
		{regexp.MustCompile(`(?is)<h6[^>]*>(.*?)</h6>`), 6},
	}

	for _, hp := range headingPatterns {
		matches := hp.pattern.FindAllStringSubmatch(htmlContent, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				text := strings.TrimSpace(stripHTMLTags(match[1]))
				if text != "" {
					elements = append(elements, &parser.Heading{
						Text:  text,
						Level: hp.level,
					})
				}
			}
		}
	}

	// Extract paragraphs
	reParagraph := regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	paragraphMatches := reParagraph.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range paragraphMatches {
		if len(match) >= 2 {
			text := stripHTMLTags(match[1])
			if strings.TrimSpace(text) != "" {
				elements = append(elements, &parser.Paragraph{
					Text: strings.TrimSpace(text),
					HTML: match[0],
				})
			}
		}
	}

	// If no structured content found, treat entire content as one paragraph
	if len(elements) == 0 {
		text := stripHTMLTags(htmlContent)
		if strings.TrimSpace(text) != "" {
			elements = append(elements, &parser.Paragraph{
				Text: strings.TrimSpace(text),
				HTML: htmlContent,
			})
		}
	}

	return elements
}

func extractChapterTitle(htmlContent, fallback string) string {
	headingPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`),
		regexp.MustCompile(`(?is)<h2[^>]*>(.*?)</h2>`),
	}
	for _, pattern := range headingPatterns {
		matches := pattern.FindStringSubmatch(htmlContent)
		if len(matches) < 2 {
			continue
		}
		title := strings.TrimSpace(stripHTMLTags(matches[1]))
		if title != "" {
			return title
		}
	}

	titlePattern := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	titleMatches := titlePattern.FindStringSubmatch(htmlContent)
	if len(titleMatches) >= 2 {
		title := strings.TrimSpace(stripHTMLTags(titleMatches[1]))
		if title != "" {
			return title
		}
	}

	return fallback
}

func findAnchorStart(htmlContent, anchor string) int {
	if anchor == "" {
		return 0
	}
	quotedAnchor := regexp.QuoteMeta(anchor)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?is)<[^>]*\sid\s*=\s*"` + quotedAnchor + `"[^>]*>`),
		regexp.MustCompile(`(?is)<[^>]*\sname\s*=\s*"` + quotedAnchor + `"[^>]*>`),
		regexp.MustCompile(`(?is)<[^>]*\sid\s*=\s*'` + quotedAnchor + `'[^>]*>`),
		regexp.MustCompile(`(?is)<[^>]*\sname\s*=\s*'` + quotedAnchor + `'[^>]*>`),
	}
	for _, pattern := range patterns {
		loc := pattern.FindStringIndex(htmlContent)
		if loc != nil {
			return loc[0]
		}
	}
	return 0
}

func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func normalizeEPUBPath(baseDir, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	if i := strings.Index(href, "?"); i >= 0 {
		href = href[:i]
	}
	return filepath.ToSlash(filepath.Clean(filepath.Join(baseDir, href)))
}
