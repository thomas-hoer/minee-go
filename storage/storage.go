package storage

// Storage provides an easy interface for store and retreive data.
type Storage interface {
	Save(key string, value interface{})
	Load(key string) interface{}
}
