package doc

type Query interface {
	Query(query string) []Entity
}

type Entity struct {
	Title  string
	Intro  string
	Url    string
	Owner  string
	ViewTS uint32
	EditTs uint32
	Extra  interface{}
}
