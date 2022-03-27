// Package grok used to parses grok patterns in Go
package grok

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	valid    = regexp.MustCompile(`^\w+([-.]\w+)*(:([-.\w]+)(:(string|float|int))?)?$`)
	normal   = regexp.MustCompile(`%{([\w-.]+(?::[\w-.]+(?::[\w-.]+)?)?)}`)
	symbolic = regexp.MustCompile(`\W`)
)

type Grok struct {
	GlobalDenormalizedPatterns map[string]string

	DenormalizedPatterns map[string]string
	CompliedGrokRe       map[string]map[string]*GrokRegexp
}

func DenormalizePattern(pattern string, denormalized ...map[string]string) (string, error) {
	for _, values := range normal.FindAllStringSubmatch(pattern, -1) {
		if !valid.MatchString(values[1]) {
			return "", fmt.Errorf("invalid pattern %%{%s}", values[1])
		}
		names := strings.Split(values[1], ":")

		syntax, alias := names[0], names[0]
		if len(names) > 1 {
			alias = symbolic.ReplaceAllString(names[1], "_")
		}

		ok := false
		storedPattern := ""
		for _, denormalized := range denormalized {
			storedPattern, ok = denormalized[syntax]
			if ok {
				break
			}
		}
		if !ok {
			return "", fmt.Errorf("no pattern found for %%{%s}", syntax)
		}

		var buffer bytes.Buffer
		if len(names) > 1 {
			buffer.WriteString("(?P<")
			buffer.WriteString(alias)
			buffer.WriteString(">")
			buffer.WriteString(storedPattern)
			buffer.WriteString(")")
		} else {
			buffer.WriteString("(")
			buffer.WriteString(storedPattern)
			buffer.WriteString(")")
		}

		pattern = strings.ReplaceAll(pattern, values[0], buffer.String())
	}

	return pattern, nil
}

func LoadPatternsFromPath(path string) (map[string]string, error) {
	if fi, err := os.Stat(path); err == nil {
		if fi.IsDir() {
			path += "/*"
		}
	} else {
		return nil, fmt.Errorf("invalid path : %s", path)
	}

	// only one error can be raised, when pattern is malformed
	// pattern is hard-coded "/*" so we ignore err
	files, _ := filepath.Glob(path)

	filePatterns := map[string]string{}
	for _, fileName := range files {
		// TODO limit filepath range
		// nolint:gosec
		file, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(bufio.NewReader(file))

		for scanner.Scan() {
			l := scanner.Text()
			if len(l) > 0 && l[0] != '#' {
				names := strings.SplitN(l, " ", 2)
				if len(names) == 2 {
					filePatterns[names[0]] = names[1]
				}
			}
		}

		_ = file.Close()
	}
	return filePatterns, nil
}

// DenormalizePatternsFromMap denormalize pattern from map,
// will return a valid pattern:value map and an invalid pattern:error map.
func DenormalizePatternsFromMap(m map[string]string, denormalized ...map[string]string) (map[string]string, map[string]string) {
	patternDeps := map[string]*nodeP{}

	for key, value := range m {
		node := &nodeP{
			cnt:   value,
			cNode: []string{},
		}

		// sub pattern
		for _, key := range normal.FindAllStringSubmatch(value, -1) {
			names := strings.Split(key[1], ":")
			syntax := names[0]

			if _, ok := m[syntax]; ok {
			} else { // 取 denormalized 的
				for _, v := range denormalized {
					if deV, ok := v[syntax]; ok {
						node.cNode = append(node.cNode, syntax)
						patternDeps[syntax] = &nodeP{
							cnt: deV,
						}
						break
					}
				}
			}
			node.cNode = append(node.cNode, syntax)
		}
		patternDeps[key] = node
	}

	return runTree(patternDeps)
}

func CompilePattern(pattern string, denormalized ...map[string]string) (*GrokRegexp, error) {
	denormalizedP, err := DenormalizePattern(pattern, denormalized...)
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(denormalizedP)
	if err != nil {
		return nil, err
	}
	return &GrokRegexp{
		Pattern:             pattern,
		DenormalizedPattern: denormalizedP,
		Re:                  re,
	}, nil
}
