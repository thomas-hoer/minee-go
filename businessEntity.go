package minee

type Page struct {
	FileName string
}
type Component struct {
}
type Context struct {
}

type BusinessEntity struct {
	Name            string
	Type            string
	Instanceable    bool
	ContextRoot     string
	Page            *Page
	Component       *Component
	AllowedSubTypes []string

	Unmarshal     func([]byte) (interface{}, error)
	Marshal       func(interface{}) ([]byte, error)
	OnPostBefore  func(context Context, data interface{}) (interface{}, string)
	OnPostCompute func(context Context, data interface{}) interface{}
	OnPostAfter   func(context Context, data interface{}) interface{}
}
