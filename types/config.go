package types

type Config struct {
	SyntaxHighlight SyntaxHighlightConfig
	Ignore          []string
	Hooks           Hooks
	NoWatch         []string
}

type SyntaxHighlightConfig struct {
	Enabled     bool
	Style       string
	LineNumbers bool
}

type Hooks struct {
	Build Hook
}

type Hook struct {
	Before []string
	After  []string
}
