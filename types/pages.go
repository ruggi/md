package types

type Layout struct {
	Title   string
	Content string
}

type Page struct {
	Title string
	Path  string
}

type FileData struct {
	Title string
	Pages map[string][]Page
}
