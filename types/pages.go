package types

import "time"

type Layout struct {
	Page
	Content string
}

type Page struct {
	Title string
	Path  string
	Date  time.Time
}

type FileData struct {
	Page
	Pages map[string][]Page
}

type PagesByDate []Page

func (p PagesByDate) Len() int {
	return len(p)
}

func (p PagesByDate) Less(a, b int) bool {
	return p[a].Date.Before(p[b].Date)
}

func (p PagesByDate) Swap(a, b int) {
	p[a], p[b] = p[b], p[a]
}
