package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	fileHeader = `//此文件由.grok_rule_convert.go文件自动生成，请不要手动修改

package patterns

var GlobalPatterns = map[string][][]string {
`
	fileTailer = `}`

	contHeader = `    "%s" : {
`
	contTailer = `    },

`
)

var (
	convertMap []map[string]string

	rege = regexp.MustCompile(`%{(\w+(?::\w+(?::\w+)?)?)}`)
)

func main() {
	convertMap = make([]map[string]string, 100)
	dir, _ := os.Getwd()
	Convert(filepath.Join(dir, "patterns"), filepath.Join(dir, "pattern.go"))
}

func Convert(inputDir, outputFile string) {
	text := fileHeader
	mkConvertMap(inputDir)
	files, _ := getFilesByDir(inputDir)

	for _, file := range files {
		text += fmt.Sprintf(contHeader, file)
		content, _ := ioutil.ReadFile(filepath.Join(inputDir, file))
		s := bufio.NewScanner(strings.NewReader(string(content)))
		for s.Scan() {
			str := s.Text()

			lineElem := strings.SplitN(str, " ", 2)
			if len(lineElem) < 2 {
				continue
			}

			cvtMap := convertMap[len(lineElem[0])]
			if cvtMap == nil {
				continue
			}

			v, ok := cvtMap[lineElem[0]]
			if !ok {
				continue
			}

			str = strings.TrimSpace(lineElem[1])
			found := make(map[string]int)
			for _, key := range rege.FindAllStringSubmatch(str, -1) {
				names := strings.Split(key[1], ":")
				found[names[0]] = 1
			}

			//先把长模式大写串转换成小写串
			for i := len(convertMap) - 1; i >= 0; i-- {
				cvtMap := convertMap[i]
				for k, v := range cvtMap {
					if _, ok := found[k]; ok {
						str = strings.Replace(str, k, v, -1)
					}
				}
			}
			text += fmt.Sprintf("		{\"%v\", `%v`},\n", v, str)
		}
		text += contTailer
	}
	text += fileTailer

	ioutil.WriteFile(outputFile, []byte(text), 0666)
}

func mkConvertMap(inputDir string) {
	files, _ := getFilesByDir(inputDir)

	for _, file := range files {
		content, _ := ioutil.ReadFile(filepath.Join(inputDir, file))
		s := bufio.NewScanner(strings.NewReader(string(content)))
		for s.Scan() {
			str := s.Text()
			lineElem := strings.Split(str, " ")
			if len(lineElem) < 2 {
				continue
			}
			if lineElem[0][0] == '#' {
				continue
			}

			nPattern := "_" + strings.ToLower(lineElem[0])
			patternLen := len(lineElem[0])
			if convertMap[patternLen] == nil {
				convertMap[patternLen] = make(map[string]string)
			}
			convertMap[patternLen][lineElem[0]] = nPattern
		}
	}
}

func getFilesByDir(dir string) ([]string, error) {
	files := make([]string, 0)
	fileInfoList, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for i := range fileInfoList {
		if !fileInfoList[i].IsDir() {
			files = append(files, fileInfoList[i].Name())
		}
	}

	return files, nil
}
