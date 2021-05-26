package types

type Config struct {
	SyntaxHighlight SyntaxHighlightConfig
}

type SyntaxHighlightConfig struct {
	Enabled     bool
	Style       string
	LineNumbers bool
}
