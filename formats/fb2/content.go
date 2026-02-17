package fb2

import (
	"html"
	"regexp"
	"strings"

	"github.com/vpoluyaktov/biblio-ebook-parser/parser"
)

func sectionToElements(section fb2Section) []parser.Element {
	elements := []parser.Element{}

	// Add title as heading if present
	if section.Title.Content != "" {
		titleText := fb2XMLToText(section.Title.Content)
		if strings.TrimSpace(titleText) != "" {
			elements = append(elements, &parser.Heading{
				Text:  strings.TrimSpace(titleText),
				Level: 2,
			})
		}
	}

	// Add epigraphs
	for _, epigraph := range section.Epigraphs {
		epigraphParas := []parser.Paragraph{}
		for _, p := range epigraph.Paragraphs {
			text := fb2XMLToText(p.Content)
			if strings.TrimSpace(text) != "" {
				epigraphParas = append(epigraphParas, parser.Paragraph{
					Text: strings.TrimSpace(text),
					HTML: p.Content,
				})
			}
		}
		if len(epigraphParas) > 0 {
			elements = append(elements, &parser.Epigraph{
				Paragraphs: epigraphParas,
			})
		}
	}

	// Add paragraphs
	for _, p := range section.Paragraphs {
		text := fb2XMLToText(p.Content)
		if strings.TrimSpace(text) != "" {
			elements = append(elements, &parser.Paragraph{
				Text: strings.TrimSpace(text),
				HTML: p.Content,
			})
		}
	}

	return elements
}

func fb2XMLToText(xmlContent string) string {
	if xmlContent == "" {
		return ""
	}

	text := xmlContent

	// Remove nested section tags
	reFB2Section := regexp.MustCompile(`(?is)<section[^>]*>.*?</section>`)
	for {
		newText := reFB2Section.ReplaceAllString(text, "")
		if newText == text {
			break
		}
		text = newText
	}

	// Handle special elements
	reFB2Table := regexp.MustCompile(`(?i)<table[^>]*>.*?</table>`)
	reFB2Image := regexp.MustCompile(`(?i)<image[^>]*/?>`)
	reFB2EmptyLine := regexp.MustCompile(`(?i)<empty-line\s*/?>`)
	reFB2Link := regexp.MustCompile(`(?is)<a[^>]*>.*?</a>`)

	text = reFB2Table.ReplaceAllString(text, "\n[Table]\n")
	text = reFB2Image.ReplaceAllString(text, "\n[Image]\n")
	text = reFB2EmptyLine.ReplaceAllString(text, "\n")
	text = reFB2Link.ReplaceAllString(text, "")

	// Handle paragraphs and titles
	reFB2PClose := regexp.MustCompile(`(?i)</p>`)
	reFB2POpen := regexp.MustCompile(`(?i)<p[^>]*>`)
	reFB2TitleClose := regexp.MustCompile(`(?i)</title>`)
	reFB2TitleOpen := regexp.MustCompile(`(?i)<title[^>]*>`)
	reFB2SubClose := regexp.MustCompile(`(?i)</subtitle>`)
	reFB2SubOpen := regexp.MustCompile(`(?i)<subtitle[^>]*>`)

	text = reFB2PClose.ReplaceAllString(text, "\n")
	text = reFB2POpen.ReplaceAllString(text, "")
	text = reFB2TitleClose.ReplaceAllString(text, "\n")
	text = reFB2TitleOpen.ReplaceAllString(text, "\n")
	text = reFB2SubClose.ReplaceAllString(text, "\n")
	text = reFB2SubOpen.ReplaceAllString(text, "\n")

	// Remove remaining XML tags
	reFB2Tags := regexp.MustCompile(`<[^>]+>`)
	text = reFB2Tags.ReplaceAllString(text, "")

	// Decode HTML entities
	text = html.UnescapeString(text)

	// Clean up whitespace
	text = strings.ReplaceAll(text, "\u00A0", " ")
	reFB2Spaces := regexp.MustCompile(`[ \t]+`)
	reFB2Newlines := regexp.MustCompile(`\n{2,}`)
	text = reFB2Spaces.ReplaceAllString(text, " ")
	text = reFB2Newlines.ReplaceAllString(text, "\n")

	return strings.TrimSpace(text)
}
