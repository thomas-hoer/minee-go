package minee

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type staticProvider struct {
	root  string
	cache map[string][]byte
}

func (s *staticProvider) Get(name string) []byte {
	if val, ok := s.cache[name]; ok {
		return val
	}
	return nil
}

func (s *staticProvider) init() {
	s.cache = make(map[string][]byte)
	filepath.Walk(s.root, func(path string, info fs.FileInfo, err error) error {
		key := strings.ReplaceAll(strings.Replace(path, s.root, "", 1), `\`, "/")
		dat, _ := os.ReadFile(path)
		s.cache[key] = dat
		return nil
	})
}
