package cover

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

//go:embed fonts/*.ttf
var fontsFS embed.FS

//go:embed images/BookCover.png
var coverTemplateFS embed.FS

const (
	coverWidth  = 300
	coverHeight = 426

	// Frame boundaries (scaled from 845x1196 template with ~60px borders)
	// Added extra padding to ensure text never touches the ornate border
	frameLeft   = 50
	frameRight  = 250
	frameTop    = 35
	frameBottom = 391
	frameWidth  = frameRight - frameLeft // ~230px usable width
	frameHeight = frameBottom - frameTop // ~356px usable height
)

// Bright gold/yellow color for better contrast with the dark background
var goldColor = color.RGBA{255, 225, 140, 255}

var (
	boldFont    *truetype.Font
	italicFont  *truetype.Font
	templateImg image.Image
)

func init() {
	boldData, err := fontsFS.ReadFile("fonts/Cormorant-Bold.ttf")
	if err != nil {
		panic("failed to load bold font: " + err.Error())
	}
	boldFont, err = truetype.Parse(boldData)
	if err != nil {
		panic("failed to parse bold font: " + err.Error())
	}

	italicData, err := fontsFS.ReadFile("fonts/Cormorant-Italic.ttf")
	if err != nil {
		panic("failed to load italic font: " + err.Error())
	}
	italicFont, err = truetype.Parse(italicData)
	if err != nil {
		panic("failed to parse italic font: " + err.Error())
	}

	// Load the cover template
	templateData, err := coverTemplateFS.ReadFile("images/BookCover.png")
	if err != nil {
		panic("failed to load cover template: " + err.Error())
	}
	templateImg, _, err = image.Decode(bytes.NewReader(templateData))
	if err != nil {
		panic("failed to decode cover template: " + err.Error())
	}
}

// GeneratePlaceholder creates a book cover image with title and author
// using the embedded template image
func GeneratePlaceholder(title, author string) ([]byte, error) {
	dc := gg.NewContext(coverWidth, coverHeight)

	// Draw the template image scaled to fit
	if templateImg != nil {
		dc.DrawImageAnchored(templateImg, coverWidth/2, coverHeight/2, 0.5, 0.5)
		// Scale the template to fit our cover dimensions
		scaleX := float64(coverWidth) / float64(templateImg.Bounds().Dx())
		scaleY := float64(coverHeight) / float64(templateImg.Bounds().Dy())
		dc.Clear()
		dc.Push()
		dc.Scale(scaleX, scaleY)
		dc.DrawImage(templateImg, 0, 0)
		dc.Pop()
	} else {
		// Fallback: draw a simple brown background if template not loaded
		dc.SetColor(color.RGBA{92, 51, 46, 255})
		dc.DrawRectangle(0, 0, coverWidth, coverHeight)
		dc.Fill()
	}

	// Draw author at the top
	drawAuthor(dc, author)

	// Draw title in the center
	drawTitle(dc, title)

	// Encode to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dc.Image(), &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func drawTitle(dc *gg.Context, title string) {
	if boldFont == nil {
		return
	}

	// Remove surrounding quotes if present
	title = strings.Trim(title, `"'`)
	title = strings.TrimPrefix(title, "\u00AB") // «
	title = strings.TrimSuffix(title, "\u00BB") // »
	title = strings.TrimPrefix(title, "\u201E") // „
	title = strings.TrimSuffix(title, "\u201C") // "
	title = strings.TrimPrefix(title, "\u201C") // "
	title = strings.TrimSuffix(title, "\u201D") // "

	// Calculate font size based on title length (larger sizes for readability)
	fontSize := 38.0
	if len(title) > 60 {
		fontSize = 24.0
	} else if len(title) > 40 {
		fontSize = 28.0
	} else if len(title) > 25 {
		fontSize = 32.0
	}

	face := truetype.NewFace(boldFont, &truetype.Options{Size: fontSize})
	dc.SetFontFace(face)
	dc.SetColor(goldColor)

	// Wrap text to fit within the frame with padding
	maxWidth := float64(frameWidth) - 40
	lines := wrapText(dc, title, maxWidth)

	// Center title vertically in the frame area, shifted down by 10%
	lineHeight := fontSize * 1.3
	totalHeight := float64(len(lines)) * lineHeight
	centerY := float64(frameTop+frameBottom)/2 + float64(frameHeight)*0.10
	startY := centerY - totalHeight/2 + lineHeight/2

	for i, line := range lines {
		y := startY + float64(i)*lineHeight
		dc.DrawStringAnchored(line, float64(coverWidth)/2, y, 0.5, 0.5)
	}
}

func drawAuthor(dc *gg.Context, author string) {
	if italicFont == nil || author == "" {
		return
	}

	fontSize := 24.0
	face := truetype.NewFace(italicFont, &truetype.Options{Size: fontSize})
	dc.SetFontFace(face)
	dc.SetColor(goldColor)

	// Wrap author text to fit inside the frame with padding
	maxWidth := float64(frameWidth) - 20
	lines := wrapText(dc, author, maxWidth)

	// Position author at the top of the frame area, shifted down by 10%
	lineHeight := fontSize * 1.3
	startY := float64(frameTop) + 45 + float64(frameHeight)*0.10

	for i, line := range lines {
		if i >= 2 { // Limit to 2 lines for author
			break
		}
		y := startY + float64(i)*lineHeight
		dc.DrawStringAnchored(line, float64(coverWidth)/2, y, 0.5, 0.5)
	}
}

func wrapText(dc *gg.Context, text string, maxWidth float64) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		w, _ := dc.MeasureString(testLine)
		if w > maxWidth && currentLine != "" {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// Limit to reasonable number of lines
	if len(lines) > 6 {
		lines = lines[:6]
		lines[5] = lines[5] + "..."
	}

	return lines
}

// GeneratePlaceholderImage returns an image.Image instead of bytes
func GeneratePlaceholderImage(title, author string) (image.Image, error) {
	data, err := GeneratePlaceholder(title, author)
	if err != nil {
		return nil, err
	}
	return jpeg.Decode(bytes.NewReader(data))
}
