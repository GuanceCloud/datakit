package grok

import (
	"fmt"
)

type nodeP struct {
	cnt   string
	ptn   *GrokPattern
	cNode []string
}

type path struct {
	m map[string]struct{}
	l []string
}

func (p *path) String() {
}

func runTree(m map[string]*nodeP) (map[string]*GrokPattern, map[string]string) {
	ret := map[string]*GrokPattern{}
	invalid := map[string]string{}
	pt := &path{
		m: map[string]struct{}{},
		l: []string{},
	}
	for name, v := range m {
		if err := dfs(ret, m, name, v, pt); err != nil {
			invalid[name] = err.Error()
		}
	}
	return ret, invalid
}

func dfs(deP map[string]*GrokPattern, top map[string]*nodeP, startName string, start *nodeP, pt *path) error {
	if _, ok := pt.m[startName]; ok {
		lineStr := ""
		for _, k := range pt.l {
			lineStr += k + " -> "
		}
		lineStr += startName
		return fmt.Errorf("circular dependency: pattern %s", lineStr)
	}

	pt.m[startName] = struct{}{}
	pt.l = append(pt.l, startName)
	defer func() {
		delete(pt.m, startName)
		pt.l = pt.l[:len(pt.l)-1]
	}()

	if _, ok := deP[startName]; ok {
		return nil
	} else if len(start.cNode) == 0 {
		if ptn, err := DenormalizePattern(start.cnt, PatternStorage{deP}); err != nil {
			return err
		} else {
			deP[startName] = ptn
		}
		return nil
	}

	for _, name := range start.cNode {
		cNode, ok := top[name]
		if !ok || cNode == nil {
			return fmt.Errorf("no pattern found for %%{%s}", name)
		}

		// 完成此 node 的编译
		if err := dfs(deP, top, name, cNode, pt); err != nil {
			return err
		}
	}

	if ptn, err := DenormalizePattern(start.cnt, PatternStorage{deP}); err != nil {
		return err
	} else {
		deP[startName] = ptn
	}

	return nil
}
