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

	yaml "gopkg.in/yaml.v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	IpV4_Len       = 4
	FILE_SEPERATOR = " "
)

var (
	IspValid = map[string]string{
		"chinanet": "中国电信",
		"cmcc":     "中国移动",
		"unicom":   "中国联通",
		"tietong":  "中国铁通",
		"cernet":   "教育网",
		"cstnet":   "科技网",
		"drpeng":   "鹏博士",
		"googlecn": "谷歌中国",
	}
)

var (
	l               = logger.DefaultSLogger("ip2isp")
	Ip2IspDb        = map[string]string{}
	BitConvTemplate = []string{
		"00000000",
		"00000001",
		"00000010",
		"00000011",
		"00000100",
		"00000101",
		"00000110",
		"00000111",
		"00001000",
		"00001001",
		"00001010",
		"00001011",
		"00001100",
		"00001101",
		"00001110",
		"00001111",
		"00010000",
		"00010001",
		"00010010",
		"00010011",
		"00010100",
		"00010101",
		"00010110",
		"00010111",
		"00011000",
		"00011001",
		"00011010",
		"00011011",
		"00011100",
		"00011101",
		"00011110",
		"00011111",
		"00100000",
		"00100001",
		"00100010",
		"00100011",
		"00100100",
		"00100101",
		"00100110",
		"00100111",
		"00101000",
		"00101001",
		"00101010",
		"00101011",
		"00101100",
		"00101101",
		"00101110",
		"00101111",
		"00110000",
		"00110001",
		"00110010",
		"00110011",
		"00110100",
		"00110101",
		"00110110",
		"00110111",
		"00111000",
		"00111001",
		"00111010",
		"00111011",
		"00111100",
		"00111101",
		"00111110",
		"00111111",
		"01000000",
		"01000001",
		"01000010",
		"01000011",
		"01000100",
		"01000101",
		"01000110",
		"01000111",
		"01001000",
		"01001001",
		"01001010",
		"01001011",
		"01001100",
		"01001101",
		"01001110",
		"01001111",
		"01010000",
		"01010001",
		"01010010",
		"01010011",
		"01010100",
		"01010101",
		"01010110",
		"01010111",
		"01011000",
		"01011001",
		"01011010",
		"01011011",
		"01011100",
		"01011101",
		"01011110",
		"01011111",
		"01100000",
		"01100001",
		"01100010",
		"01100011",
		"01100100",
		"01100101",
		"01100110",
		"01100111",
		"01101000",
		"01101001",
		"01101010",
		"01101011",
		"01101100",
		"01101101",
		"01101110",
		"01101111",
		"01110000",
		"01110001",
		"01110010",
		"01110011",
		"01110100",
		"01110101",
		"01110110",
		"01110111",
		"01111000",
		"01111001",
		"01111010",
		"01111011",
		"01111100",
		"01111101",
		"01111110",
		"01111111",
		"10000000",
		"10000001",
		"10000010",
		"10000011",
		"10000100",
		"10000101",
		"10000110",
		"10000111",
		"10001000",
		"10001001",
		"10001010",
		"10001011",
		"10001100",
		"10001101",
		"10001110",
		"10001111",
		"10010000",
		"10010001",
		"10010010",
		"10010011",
		"10010100",
		"10010101",
		"10010110",
		"10010111",
		"10011000",
		"10011001",
		"10011010",
		"10011011",
		"10011100",
		"10011101",
		"10011110",
		"10011111",
		"10100000",
		"10100001",
		"10100010",
		"10100011",
		"10100100",
		"10100101",
		"10100110",
		"10100111",
		"10101000",
		"10101001",
		"10101010",
		"10101011",
		"10101100",
		"10101101",
		"10101110",
		"10101111",
		"10110000",
		"10110001",
		"10110010",
		"10110011",
		"10110100",
		"10110101",
		"10110110",
		"10110111",
		"10111000",
		"10111001",
		"10111010",
		"10111011",
		"10111100",
		"10111101",
		"10111110",
		"10111111",
		"11000000",
		"11000001",
		"11000010",
		"11000011",
		"11000100",
		"11000101",
		"11000110",
		"11000111",
		"11001000",
		"11001001",
		"11001010",
		"11001011",
		"11001100",
		"11001101",
		"11001110",
		"11001111",
		"11010000",
		"11010001",
		"11010010",
		"11010011",
		"11010100",
		"11010101",
		"11010110",
		"11010111",
		"11011000",
		"11011001",
		"11011010",
		"11011011",
		"11011100",
		"11011101",
		"11011110",
		"11011111",
		"11100000",
		"11100001",
		"11100010",
		"11100011",
		"11100100",
		"11100101",
		"11100110",
		"11100111",
		"11101000",
		"11101001",
		"11101010",
		"11101011",
		"11101100",
		"11101101",
		"11101110",
		"11101111",
		"11110000",
		"11110001",
		"11110010",
		"11110011",
		"11110100",
		"11110101",
		"11110110",
		"11110111",
		"11111000",
		"11111001",
		"11111010",
		"11111011",
		"11111100",
		"11111101",
		"11111110",
		"11111111",
	}
)

func ParseIpCIDR(ipCidr string) (string, error) {
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
	if len(ipBytes) != IpV4_Len {
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
	if len(Ip2IspDb) == 0 {
		return "unknown"
	}

	for i := 32; i > 0; i-- {
		ipCidr := fmt.Sprintf("%s/%v", ip, i)
		ipBitStr, _ := ParseIpCIDR(ipCidr)
		if v, ok := Ip2IspDb[ipBitStr]; ok {
			return v
		}
	}
	return "unknown"
}

func Init(f string) error {
	m := make(map[string]string)

	if !datakit.FileExist(f) {
		l.Warnf("%v not found", f)
		return nil
	}

	fd, err := os.Open(f)
	if err != nil {
		return err
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		contents := strings.Split(scanner.Text(), FILE_SEPERATOR)
		if len(contents) != 2 {
			continue
		}

		ipBitStr, err := ParseIpCIDR(contents[0])
		if err != nil {
			continue
		}
		m[ipBitStr] = contents[1]
	}

	if len(m) != 0 {
		Ip2IspDb = m
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

		fd, err := os.Open(filepath.Join(from, file))
		if err != nil {
			return err
		}
		defer fd.Close()

		scanner := bufio.NewScanner(fd)
		for scanner.Scan() {
			c := fmt.Sprintf("%v%v%v", scanner.Text(), FILE_SEPERATOR, isp)
			content = append(content, c)
		}
	}

	return ioutil.WriteFile(to, []byte(strings.Join(content, "\n")), 0x666)
}

func BuildContryCity(csvFile, outputFile string) error {
	d := make(map[string]map[string][]string)
	found := make(map[string]uint8)

	f, err := os.Open(csvFile)
	if err != nil {
		return err
	}
	defer f.Close()

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

	return ioutil.WriteFile(outputFile, r, 0x666)
}
