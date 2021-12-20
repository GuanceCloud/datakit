// Package ip2isp append ISP info to IP address
package ip2isp

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	yaml "gopkg.in/yaml.v2"
)

const (
	IPV4Len       = 4
	FileSeparator = " "
)

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

var (
	l        = logger.DefaultSLogger("ip2isp")
	IP2ISPDB = map[string]string{}
)

func ParseIPCIDR(ipCidr string) (string, error) {
	var err error
	var cidrLen int64 = 32

	ipCidrs := strings.Split(ipCidr, "/")
	if len(ipCidrs) == 2 {
		cidrLen, err = strconv.ParseInt(ipCidrs[1], 10, 8)
		if err != nil {
			return "", err
		}
	}

	ipBytes := strings.Split(ipCidrs[0], ".")
	if len(ipBytes) != IPV4Len {
		return "", fmt.Errorf("invalid ip address")
	}
	ipBitStr := ""
	for _, ipByteStr := range ipBytes {
		ip, err := strconv.ParseInt(ipByteStr, 10, 16)
		if err != nil {
			return "", err
		}
		if cidrLen >= 8 {
			ipBitStr += BitConvTemplate[ip]
		} else {
			ipBitStr += BitConvTemplate[ip][0:cidrLen]
		}
		cidrLen -= 8
		if cidrLen <= 0 {
			break
		}
	}
	return ipBitStr, nil
}

func SearchIsp(ip string) string {
	if len(IP2ISPDB) == 0 {
		return "unknown"
	}

	for i := 32; i > 0; i-- {
		ipCidr := fmt.Sprintf("%s/%v", ip, i)
		ipBitStr, _ := ParseIPCIDR(ipCidr)
		if v, ok := IP2ISPDB[ipBitStr]; ok {
			return v
		}
	}
	return "unknown"
}

func Init(f string) error {
	l = logger.SLogger("ip2isp")

	l.Debugf("setup ipdb from %s", f)

	m := make(map[string]string)

	if !datakit.FileExist(f) {
		l.Warnf("%v not found", f)
		return nil
	}

	fd, err := os.Open(filepath.Clean(f))
	if err != nil {
		return err
	}
	defer fd.Close() //nolint:errcheck,gosec

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		contents := strings.Split(scanner.Text(), FileSeparator)
		if len(contents) != 2 {
			continue
		}

		ipBitStr, err := ParseIPCIDR(contents[0])
		if err != nil {
			continue
		}
		m[ipBitStr] = contents[1]
	}

	if len(m) != 0 {
		IP2ISPDB = m
		l.Infof("found new %d rules", len(m))
	} else {
		l.Infof("no rules founded")
	}

	return nil
}

func MergeIsp(from, to string) error {
	files, err := ioutil.ReadDir(from)
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

	return ioutil.WriteFile(to, []byte(strings.Join(content, "\n")), datakit.ConfPerm)
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

	return ioutil.WriteFile(outputFile, r, datakit.ConfPerm)
}
