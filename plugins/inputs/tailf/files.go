package tailf

import (
	"github.com/gobwas/glob"
	"github.com/mattn/go-zglob"
)

// FIXME:
// 可以不作为 Tailf 成员函数，仅仅是为了打印那两条 Warnning

func (t *Tailf) getFileList(filesGlob, ignoreGlob []string) []string {
	var matches = make(map[string]interface{})

	files := t.fileList(filesGlob)
	ignores := t.ignoreList(ignoreGlob)

	for _, f := range files {
		for _, g := range ignores {
			if g.Match(f) {
				goto next
			}
		}
		matches[f] = nil
	next:
	}

	// unique
	var list []string
	for matche := range matches {
		list = append(list, matche)
	}

	return list
}

func (t *Tailf) fileList(filesGlob []string) []string {
	var matches []string
	for _, f := range filesGlob {
		matche, err := zglob.Glob(f)
		if err != nil {
			t.log.Warnf("logfile %s, %s", f, err)
			continue
		}
		matches = append(matches, matche...)
	}
	return matches
}

func (t *Tailf) ignoreList(ignoreGlob []string) []glob.Glob {
	var globs []glob.Glob
	for _, ig := range ignoreGlob {
		g, err := glob.Compile(ig)
		if err != nil {
			t.log.Errorf("ignore glob, %s", err)
			continue
		}
		globs = append(globs, g)
	}
	return globs
}
