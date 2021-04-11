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

func (s *staticProvider) init() error {
	s.cache = make(map[string][]byte)
	return filepath.WalkDir(s.root, func(path string, info fs.DirEntry, err error) error {
		if info.IsDir() {
			return nil
		}
		key := strings.Replace(strings.ReplaceAll(path, `\`, "/"), s.root, "", 1)
		dat, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		s.cache[key] = dat
		return nil
	})
}
