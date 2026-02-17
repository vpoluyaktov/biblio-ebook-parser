package formats

import (
	"github.com/vpoluyaktov/biblio-ebook-parser/formats/epub"
	"github.com/vpoluyaktov/biblio-ebook-parser/formats/fb2"
	"github.com/vpoluyaktov/biblio-ebook-parser/parser"
)

func init() {
	// Register EPUB parser
	parser.Register("epub", epub.NewParser())
	parser.Register("epub.zip", epub.NewParser())

	// Register FB2 parser
	parser.Register("fb2", fb2.NewParser())
	parser.Register("fb2.zip", fb2.NewParser())
}
