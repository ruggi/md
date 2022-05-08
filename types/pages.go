package types

import (
	"time"
)

type Layout struct {
	Page
	Content string
}

type Breadcrumb struct {
	Path string
	Base string
}

type Page struct {
	Title       string
	Path        string
	Breadcrumbs []Breadcrumb
	Base        string
	Dir         string
	IsDir       bool
	Parent      string
	Date        time.Time
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

type PagesByName []Page

func (p PagesByName) Len() int {
	return len(p)
}

func (p PagesByName) Less(a, b int) bool {
	return p[a].Path < p[b].Path
}

func (p PagesByName) Swap(a, b int) {
	p[a], p[b] = p[b], p[a]
}
