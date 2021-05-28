package types

type Config struct {
	SyntaxHighlight SyntaxHighlightConfig
	Ignore          []string
}

type SyntaxHighlightConfig struct {
	Enabled     bool
	Style       string
	LineNumbers bool
}
