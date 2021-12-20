package tailer

import (
	"path/filepath"

	"github.com/gobwas/glob"
)

type Provider struct {
	list    []string
	lastErr error
}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) SearchFiles(patterns []string) *Provider {
	if len(patterns) == 0 {
		return p
	}

	for _, pattern := range patterns {
		paths, err := filepath.Glob(pattern)
		if err != nil {
			p.lastErr = err
			continue
		}
		p.list = append(p.list, paths...)
	}

	p.list = unique(p.list)
	return p
}

func (p *Provider) IgnoreFiles(patterns []string) *Provider {
	if len(patterns) == 0 {
		return p
	}

	var ignores []glob.Glob
	for _, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			p.lastErr = err
			continue
		}
		ignores = append(ignores, g)
	}

	pass := []string{}
	for _, path := range p.list {
		matched := false
		func() {
			for _, g := range ignores {
				if g.Match(path) {
					matched = true
					return
				}
			}
		}()
		if !matched {
			pass = append(pass, path)
		}
	}

	p.list = pass
	return p
}

func (p *Provider) Result() (list []string, lastErr error) {
	return p.list, p.lastErr
}

func unique(slice []string) []string {
	var res []string
	keys := make(map[string]interface{})
	for _, str := range slice {
		if _, ok := keys[str]; !ok {
			keys[str] = nil
			res = append(res, str)
		}
	}
	return res
}
