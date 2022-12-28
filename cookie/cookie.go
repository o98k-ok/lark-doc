package cookie

import "time"

type Filter interface {
	Filter() []CookieItem
}

type CookieItem struct {
	Name    string
	Value   string
	Path    string    // optional
	Domain  string    // optional
	Expires time.Time // optional
}
