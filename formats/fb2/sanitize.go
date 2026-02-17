package fb2

import "regexp"

func sanitizeFB2XML(data []byte) []byte {
	// Don't do any sanitization for now - let XML decoder handle everything
	// including charset conversion via charsetReader
	return data
}

func fixUnescapedAmpersands(data []byte) []byte {
	result := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		if data[i] == '&' {
			// Check if this is a valid entity - work with bytes directly
			// to avoid charset corruption from string conversion
			remaining := data[i:]
			if isValidEntity(remaining) {
				// Valid entity, keep as-is
				result = append(result, data[i])
			} else {
				// Invalid/unescaped ampersand, escape it
				result = append(result, []byte("&amp;")...)
				i++
				continue
			}
		} else {
			result = append(result, data[i])
		}
		i++
	}
	return result
}

// isValidEntity checks if bytes start with a valid XML entity (ASCII-only check)
func isValidEntity(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	// Check for &amp; &lt; &gt; &quot; &apos;
	if len(data) >= 5 && string(data[:5]) == "&amp;" {
		return true
	}
	if len(data) >= 4 && string(data[:4]) == "&lt;" {
		return true
	}
	if len(data) >= 4 && string(data[:4]) == "&gt;" {
		return true
	}
	if len(data) >= 6 && string(data[:6]) == "&quot;" {
		return true
	}
	if len(data) >= 6 && string(data[:6]) == "&apos;" {
		return true
	}
	// Check for &#123; or &#xAB; (numeric entities)
	if data[1] == '#' {
		for j := 2; j < len(data) && j < 12; j++ {
			if data[j] == ';' {
				return true
			}
			if j == 2 && data[j] == 'x' {
				continue // hex entity
			}
			if !((data[j] >= '0' && data[j] <= '9') ||
				(data[j] >= 'a' && data[j] <= 'f') ||
				(data[j] >= 'A' && data[j] <= 'F')) {
				return false
			}
		}
	}
	return false
}

func fixMalformedTags(data []byte) []byte {
	// Fix tags starting with numbers, dots, or dashes
	reInvalidTagStart := regexp.MustCompile(`<([0-9]|\.\.\.|--?[^a-zA-Z>])`)
	data = reInvalidTagStart.ReplaceAllFunc(data, func(match []byte) []byte {
		return append([]byte("&lt;"), match[1:]...)
	})

	// Fix unescaped < followed by non-ASCII characters
	result := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		if data[i] == '<' {
			// Check if this is a valid XML tag start
			if i+1 >= len(data) {
				// Bare < at end of file
				result = append(result, []byte("&lt;")...)
				i++
				continue
			}
			nextByte := data[i+1]
			// Valid tag starts: a-z, A-Z, /, !, ?, _
			isValidTagStart := (nextByte >= 'a' && nextByte <= 'z') ||
				(nextByte >= 'A' && nextByte <= 'Z') ||
				nextByte == '/' || nextByte == '!' || nextByte == '?' || nextByte == '_'

			if !isValidTagStart {
				// Invalid tag start - escape the <
				result = append(result, []byte("&lt;")...)
				i++
				continue
			}
		}
		result = append(result, data[i])
		i++
	}

	return result
}
