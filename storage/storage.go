package storage

// Storage provides an easy interface for store and retrieve data.
type Storage interface {
	Save(key string, value interface{})
	Load(key string) interface{}
}
