package engine

import (
	"io"

	"github.com/ruggi/md/types"
)

type Engine interface {
	Convert(r io.Reader, w io.Writer) error
	SetSyntaxHighlight(types.SyntaxHighlightConfig)
}
