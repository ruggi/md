package engine

import "io"

type Engine interface {
	Convert(r io.Reader, w io.Writer) error
	SetSyntaxHighlight(enabled bool, style string)
}
