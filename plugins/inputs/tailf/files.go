package tailf

import (
	"github.com/gobwas/glob"
	"github.com/mattn/go-zglob"
)

func getFileList(filesGlob, ignoreGlob []string) []string {
	var matches = make(map[string]interface{})

	files := fileList(filesGlob)
	ignores := ignoreList(ignoreGlob)

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

func fileList(filesGlob []string) []string {
	var matches []string
	for _, f := range filesGlob {
		matche, err := zglob.Glob(f)
		if err != nil {
			l.Warnf("logfile %s, %s", f, err)
			continue
		}
		matches = append(matches, matche...)
	}
	return matches
}

func ignoreList(ignoreGlob []string) []glob.Glob {
	var globs []glob.Glob
	for _, ig := range ignoreGlob {
		g, err := glob.Compile(ig)
		if err != nil {
			l.Errorf("ignore glob, %s", err)
			continue
		}
		globs = append(globs, g)
	}
	return globs
}
