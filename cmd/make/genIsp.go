package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
)

const (
	Gen_template = `package ip2isp

var (
    Ip2IspDb = map[string]string {
%s    } 
)
`
)
var (
	Ip2IspDb = map[string]string{}

	IspValid = map[string]string{
		"chinanet" : "中国电信",
		"cmcc"     : "中国移动",
		"unicom"   : "中国联通",
		"tietong"  : "中国铁通",
		"cernet"   : "教育网"  ,
		"cstnet"   : "科技网"  ,
		"drpeng"   : "鹏博士"  ,
		"googlecn" : "谷歌中国",
	}
)

func genIspFile(inputDir, outputFile string) error {
	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		file := f.Name()

		//去掉统计信息文件
		if !strings.HasSuffix(file, ".txt") {
			continue
		}

		//去掉ipv6文件
		if strings.HasSuffix(file, "6.txt") {
			continue
		}

		isp := strings.TrimSuffix(file, ".txt")
		if _, ok := IspValid[isp]; !ok {
			continue
		}

		fd, err := os.Open(filepath.Join(inputDir, file))
		if err != nil {
			return err
		}
		defer fd.Close()

		scanner := bufio.NewScanner(fd)
		for scanner.Scan() {
			ipBitStr, err := ip2isp.ParseIpCIDR(scanner.Text())
			if err != nil {
				continue
			}
			Ip2IspDb[ipBitStr] = IspValid[isp]
		}
	}
	OutputFileFlush(outputFile)
	return nil
}

func OutputFileFlush(outputFile string) {
	tmp := ""
	fd, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("create file %s %v", outputFile, err)
		return
	}
	defer fd.Close()

	for k, v := range Ip2IspDb {
		tmp += fmt.Sprintf("\t\t\"%s\" : \"%s\",\n", k, v)
	}

	content := fmt.Sprintf(Gen_template, tmp)
	fd.WriteString(content)
}

func main() {
	curDir, _ := os.Getwd()
	inputDir   := filepath.Join(curDir, "china-operator-ip")
	outputFile := filepath.Join(curDir, "process", "ip2isp", "ip2isp.go")
	genIspFile(inputDir, outputFile)
}
