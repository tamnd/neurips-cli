package cli

import (
	"io"

	"github.com/tamnd/neurips-cli/pkg/render"
)

// Format is an output rendering format.
type Format = render.Format

const (
	FormatTable = render.FormatTable
	FormatJSON  = render.FormatJSON
	FormatJSONL = render.FormatJSONL
	FormatCSV   = render.FormatCSV
	FormatTSV   = render.FormatTSV
	FormatURL   = render.FormatURL
	FormatRaw   = render.FormatRaw
)

// NewRenderer builds a renderer writing to w.
func NewRenderer(w io.Writer, format Format, fields []string, noHeader bool, tmpl string) *render.Renderer {
	return render.New(w, format, fields, noHeader, tmpl)
}
