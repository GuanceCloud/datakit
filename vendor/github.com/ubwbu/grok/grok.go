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

type Grok struct{}

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

		pattern = strings.Replace(pattern, values[0], buffer.String(), -1)
	}

	return pattern, nil
}

func DenormalizePatternsFromMap(m map[string]string, denormalized ...map[string]string) (map[string]string, error) {
	patternDeps := graph{}
	de := []map[string]string{}
	de = append(de, denormalized...)
	result := map[string]string{}
	de = append(de, result)
	for k, v := range m {
		keys := []string{}
		for _, key := range normal.FindAllStringSubmatch(v, -1) {
			names := strings.Split(key[1], ":")
			syntax := names[0]
			flag := false
			for _, v := range de {
				if _, ok := v[syntax]; ok {
					flag = true
					break
				}
			}
			if !flag {
				if _, ok := m[syntax]; ok {
					flag = true
				}
			}
			if !flag {
				return nil, fmt.Errorf("no pattern found for %%{%s}", syntax)
			}

			keys = append(keys, syntax)
		}
		patternDeps[k] = keys
	}
	order, _ := sortGraph(patternDeps)
	for _, key := range reverseList(order) {
		if rst, err := DenormalizePattern(m[key], de...); err == nil {
			result[key] = rst
		} else {
			return nil, err
		}
	}
	return result, nil
}

func DenormalizePatternsFromPath(path string) (map[string]string, error) {
	if fi, err := os.Stat(path); err == nil {
		if fi.IsDir() {
			path = path + "/*"
		}
	} else {
		return nil, fmt.Errorf("invalid path : %s", path)
	}

	// only one error can be raised, when pattern is malformed
	// pattern is hard-coded "/*" so we ignore err
	files, _ := filepath.Glob(path)

	var filePatterns = map[string]string{}
	for _, fileName := range files {
		file, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(bufio.NewReader(file))

		for scanner.Scan() {
			l := scanner.Text()
			if len(l) > 0 && l[0] != '#' {
				names := strings.SplitN(l, " ", 2)
				filePatterns[names[0]] = names[1]
			}
		}

		_ = file.Close()
	}

	return DenormalizePatternsFromMap(filePatterns, nil)
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
