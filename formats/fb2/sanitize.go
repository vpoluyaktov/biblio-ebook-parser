package fb2

import (
	"regexp"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

func sanitizeFB2XML(data []byte) []byte {
	if !utf8.Valid(data) {
		data = fixInvalidUTF8(data)
	}

	data = removeIllegalXMLChars(data)
	data = fixUnescapedAmpersands(data)
	data = fixMalformedTags(data)

	return data
}

func fixInvalidUTF8(data []byte) []byte {
	result := make([]byte, 0, len(data))
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			if data[0] >= 0x80 {
				decoded := charmap.Windows1251.DecodeByte(data[0])
				result = utf8.AppendRune(result, decoded)
			} else {
				result = append(result, ' ')
			}
			data = data[1:]
		} else {
			result = utf8.AppendRune(result, r)
			data = data[size:]
		}
	}
	return result
}

func removeIllegalXMLChars(data []byte) []byte {
	result := make([]byte, 0, len(data))
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == 0x9 || r == 0xA || r == 0xD || (r >= 0x20 && r <= 0xD7FF) || (r >= 0xE000 && r <= 0xFFFD) || (r >= 0x10000 && r <= 0x10FFFF) {
			result = utf8.AppendRune(result, r)
		} else {
			result = append(result, ' ')
		}
		data = data[size:]
	}
	return result
}

func fixUnescapedAmpersands(data []byte) []byte {
	result := make([]byte, 0, len(data))
	i := 0
	for i < len(data) {
		if data[i] == '&' {
			// Check if this is a valid entity
			remaining := string(data[i:])
			if regexp.MustCompile(`^&(amp|lt|gt|quot|apos);`).MatchString(remaining) ||
				regexp.MustCompile(`^&#[0-9]+;`).MatchString(remaining) ||
				regexp.MustCompile(`^&#x[0-9a-fA-F]+;`).MatchString(remaining) {
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
