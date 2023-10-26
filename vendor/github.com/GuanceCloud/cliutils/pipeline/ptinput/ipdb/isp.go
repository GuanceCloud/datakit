// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ipdb

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const (
	IPV4Len       = 4
	FileSeparator = " "
)

// var ipdbInstance ipdb.IPdb

var IspValid = map[string]string{
	"chinanet": "中国电信",
	"cmcc":     "中国移动",
	"unicom":   "中国联通",
	"tietong":  "中国铁通",
	"cernet":   "教育网",
	"cstnet":   "科技网",
	"drpeng":   "鹏博士",
	"googlecn": "谷歌中国",
}

func MergeISP(from, to string) error {
	files, err := os.ReadDir(from)
	if err != nil {
		return err
	}

	var content []string

	for _, f := range files {
		file := f.Name()

		// 去掉统计信息文件
		if !strings.HasSuffix(file, ".txt") {
			continue
		}

		// 去掉ipv6文件
		if strings.HasSuffix(file, "6.txt") {
			continue
		}

		isp := strings.TrimSuffix(file, ".txt")
		if _, ok := IspValid[isp]; !ok {
			continue
		}

		fd, err := os.Open(filepath.Clean(filepath.Join(from, file)))
		if err != nil {
			return err
		}
		defer fd.Close() //nolint:errcheck,gosec

		scanner := bufio.NewScanner(fd)
		for scanner.Scan() {
			c := fmt.Sprintf("%v%v%v", scanner.Text(), FileSeparator, isp)
			content = append(content, c)
		}
	}

	return os.WriteFile(to, []byte(strings.Join(content, "\n")), os.ModePerm)
}

func BuildContryCity(csvFile, outputFile string) error {
	d := make(map[string]map[string][]string)
	found := make(map[string]uint8)

	f, err := os.Open(filepath.Clean(csvFile))
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec

	w := csv.NewReader(f)
	data, err := w.ReadAll()
	if err != nil {
		return err
	}

	for _, ip := range data {
		contry := ip[3]
		province := ip[4]
		city := ip[5]
		if contry == "-" || city == "-" {
			continue
		}

		uniKey := fmt.Sprintf("%v%v%v", contry, province, city)
		if _, ok := found[uniKey]; ok {
			continue
		}

		c, ok := d[contry]
		if !ok {
			c = make(map[string][]string)
			d[contry] = c
		}

		c[province] = append(c[province], city)
		found[uniKey] = 0
	}

	r, err := yaml.Marshal(d)
	if err != nil {
		return err
	}

	return os.WriteFile(outputFile, r, os.ModePerm)
}
