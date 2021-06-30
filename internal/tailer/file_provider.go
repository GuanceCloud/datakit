package tailer

import (
	"github.com/gobwas/glob"
	"github.com/mattn/go-zglob"
)

type FileList struct {
	list []string
}

func NewFileList(pathNames []string) *FileList {
	var matches []string

	for _, name := range pathNames {
		matche, err := zglob.Glob(name)
		if err != nil {
			continue
		}
		matches = append(matches, matche...)
	}

	return &FileList{list: unique(matches)}
}

func (f *FileList) Ignore(pathNames []string) *FileList {
	// FIXME: need to optimize this codes
	var ignores []glob.Glob
	for _, name := range pathNames {
		g, err := glob.Compile(name)
		if err != nil {
			continue
		}
		ignores = append(ignores, g)
	}

	for index, name := range f.list {
		func() {
			for _, g := range ignores {
				if g.Match(name) {
					remove(f.list, index)
					return
				}
			}
		}()
	}

	return f
}

func (f *FileList) List() []string {
	return f.list
}

func unique(slice []string) (array []string) {
	keys := make(map[string]interface{})
	for _, str := range slice {
		if _, ok := keys[str]; !ok {
			keys[str] = nil
			array = append(array, str)
		}
	}
	return
}

func remove(s []string, index int) []string {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}
