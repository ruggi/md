package settings

const (
	MDDir  = ".md"
	OutDir = "out"
)

var (
	SourceFileExtensions = map[string]bool{
		".md":       true,
		".markdown": true,
	}
)
