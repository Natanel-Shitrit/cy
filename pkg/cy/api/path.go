package api

import (
	_ "embed"
	"path/filepath"

	"github.com/cfoust/cy/pkg/janet"
)

//go:embed docs-path.md
var DOCS_PATH string

type PathModule struct{}

var _ janet.Documented = (*PathModule)(nil)

func (i *PathModule) Documentation() string {
	return DOCS_PATH
}

func (p *PathModule) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

func (p *PathModule) Base(path string) string {
	return filepath.Base(path)
}

func (p *PathModule) Join(elem []string) string {
	return filepath.Join(elem...)
}

func (p *PathModule) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}
