package main

import (
	"bufio"
	"io/ioutil"
	"path/filepath"
	"strings"
)

var (
	convertMap []map[string]string
)

func main() {
	convertMap = make([]map[string]string, 100)
	Convert(`D:\go\src\github.com\vjeantet\grok\patterns`, `D:\go\src\github.com\vjeantet\grok\pattern`)
}



func Convert(inputDir, outputDir string) {
	genConvertMap(inputDir)

	files, _ := getFilesByDir(inputDir)

	for _, file := range files {
		content, _ := ioutil.ReadFile(filepath.Join(inputDir, file))
		s := bufio.NewScanner(strings.NewReader(string(content)))
		newContent := ""
		for s.Scan() {
			str := s.Text()
			//先把长模式大写串转换成小写串
			for i:= len(convertMap) -1; i >=0 ;i-- {
				cvtMap := convertMap[i]
				for k, v := range cvtMap {
					str = strings.Replace(str, k, v, -1)
				}
			}
			str += "\r\n"
			newContent += str
		}
		ioutil.WriteFile(filepath.Join(outputDir, file), []byte(newContent), 0666)
	}
}

func genConvertMap(inputDir string) {
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

			nPattern := "_"+strings.ToLower(lineElem[0])
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
	fileInfoList,err := ioutil.ReadDir(dir)
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
