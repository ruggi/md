package types

type Layout struct {
	Page
	Content string
}

type Page struct {
	Title string
	Path  string
}

type FileData struct {
	Page
	Pages map[string][]Page
}
