package plaintext

import "strings"

// addPeriods adds periods at the end of paragraphs that don't have punctuation
func addPeriods(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			result = append(result, "")
			continue
		}
		
		// Skip marker lines (TITLE_BREAK, etc.)
		if strings.Contains(line, "{{") && strings.Contains(line, "}}") {
			result = append(result, line)
			continue
		}
		
		// Get last rune to handle multi-byte characters
		runes := []rune(line)
		if len(runes) == 0 {
			result = append(result, line)
			continue
		}
		
		lastRune := runes[len(runes)-1]
		
		// Check for sentence-ending punctuation (including curly quotes)
		if lastRune != '.' && lastRune != '?' && lastRune != '!' &&
			lastRune != ':' && lastRune != '"' && lastRune != 0x201C && lastRune != 0x201D {
			// Check for ellipsis
			if !strings.HasSuffix(line, "...") {
				line = line + "."
			}
		}
		
		result = append(result, line)
	}
	
	return strings.Join(result, "\n")
}
