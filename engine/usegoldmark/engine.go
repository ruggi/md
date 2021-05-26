package usegoldmark

import (
	"io"
	"io/ioutil"

	chromahtml "github.com/alecthomas/chroma/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type Engine struct {
	goldmark goldmark.Markdown
}

type EngineConf struct {
}

func NewEngine(conf EngineConf) *Engine {
	return &Engine{
		goldmark: goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
				html.WithXHTML(),
				html.WithUnsafe(),
			),
		),
	}
}

func (e *Engine) Convert(r io.Reader, w io.Writer) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return e.goldmark.Convert(data, w)
}

func (e *Engine) SetSyntaxHighlight(enabled bool, style string) {
	if !enabled {
		return
	}
	e.goldmark.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(highlighting.NewHTMLRenderer(
				highlighting.WithStyle(style),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			), 200),
		),
	)
}
