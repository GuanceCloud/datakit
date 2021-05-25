package scanport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	DNSName string = `^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`
)

var (
	l         *logger.Logger
	inputName = "scanport"
	rxDNSName = regexp.MustCompile(DNSName)
)

func (_ *Scanport) Catalog() string {
	return "network"
}

func (_ *Scanport) SampleConfig() string {
	return configSample
}

func (_ *Scanport) Description() string {
	return ""
}

func (_ *Scanport) Gather() error {
	return nil
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Scanport{}
	})
}

func (s *Scanport) Run() {
	l = logger.SLogger("scanport")

	l.Info("scanport input started...")

	s.checkCfg()

	tick := time.NewTicker(s.IntervalDuration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			s.handle()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (s *Scanport) checkCfg() {
	// 采集频度
	s.IntervalDuration = 10 * time.Minute

	if s.Interval != "" {
		du, err := time.ParseDuration(s.Interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10m", s.Interval, err.Error())
		} else {
			s.IntervalDuration = du
		}
	}
}

// handle
func (s *Scanport) handle() {
	ips, err := s.getAllIp()
	if err != nil {
		l.Errorf("convert config ip fail error %v", err.Error())
		return
	}

	var lines [][]byte

	//扫所有的ip
	for i := 0; i < len(ips); i++ {
		tags := make(map[string]string)
		fields := make(map[string]interface{})
		tm := time.Now()

		ports := s.GetIpOpenPort(ips[i])
		if len(ports) > 0 {
			b, err := json.Marshal(ports)
			if err != nil {
				l.Errorf("get all open port error %v", err)
			}

			var resStr = string(b)
			tags["ip"] = ips[i]

			if isDNSName(s.Targets) {
				tags["domain"] = s.Targets
			}

			fields["openPort"] = resStr[1 : len(resStr)-1]

			pt, err := io.MakeMetric("scanport", tags, fields, tm)
			if err != nil {
				l.Errorf("make metric point error %v", err)
			}

			lines = append(lines, pt)

			err = io.NamedFeed([]byte(pt), datakit.Metric, inputName)
			if err != nil {
				l.Errorf("push metric point error %v", err)
			}
		}
	}

	s.resData = bytes.Join(lines, []byte("\n"))
}

//获取所有ip
func (s *Scanport) getAllIp() ([]string, error) {
	var (
		ips []string
	)

	targets := []string{"127.0.0.1"}

	if len(s.Targets) != 0 {
		//处理 ","号 如 80,81,88 或 80,88-100
		targets = strings.Split(strings.Trim(s.Targets, ","), ",")
	}

	for _, target := range targets {
		target = strings.TrimSpace(target)
		if isIP(target) {
			ips = append(ips, target)
		} else if isDNSName(target) {
			// 添加ip
			ip, _ := parseDNS(target)
			ips = append(ips, ip)
		} else if isCIDR(target) {
			// crdr解析
			cidrIp, _ := parseCIDR(target)
			ips = append(ips, cidrIp...)
		} else {
			return ips, errors.New("config target not right")
		}
	}

	return ips, nil
}

//获取所有端口
func (s *Scanport) getAllPort() ([]int, error) {
	var ports []int

	portArr := []string{"80"}

	if len(s.Port) != 0 {
		//处理 ","号 如 80,81,88 或 80,88-100
		portArr = strings.Split(strings.Trim(s.Port, ","), ",")
	}

	for _, v := range portArr {
		portArr2 := strings.Split(strings.Trim(v, "-"), "-")
		startPort, err := filterPort(portArr2[0])
		if err != nil {
			continue
		}
		//第一个端口先添加
		ports = append(ports, startPort)
		if len(portArr2) > 1 {
			//添加第一个后面的所有端口
			endPort, _ := filterPort(portArr2[1])
			if endPort > startPort {
				for i := 1; i <= endPort-startPort; i++ {
					ports = append(ports, startPort+i)
				}
			}
		}
	}
	//去重复
	ports = arrayUnique(ports)

	return ports, nil
}

//查看端口号是否打开
func (s *Scanport) isOpen(ip string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), time.Millisecond*time.Duration(s.Timeout))
	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			l.Errorf("too many open files" + err.Error())
			os.Exit(1)
		}
		return false
	}
	_ = conn.Close()
	return true
}

//端口合法性过滤
func filterPort(str string) (int, error) {
	port, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	if port < 1 || port > 65535 {
		return 0, errors.New("port out of range")
	}
	return port, nil
}

//数组去重
func arrayUnique(arr []int) []int {
	var newArr []int
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return newArr
}

//获取开放端口号
func (s *Scanport) GetIpOpenPort(ip string) []int {
	var (
		total     int
		pageCount int
		num       int
		openPorts []int
		mutex     sync.Mutex
	)
	ports, _ := s.getAllPort()
	total = len(ports)
	if total < s.Process {
		pageCount = total
	} else {
		pageCount = s.Process
	}

	num = int(math.Ceil(float64(total) / float64(pageCount)))

	l.Info(fmt.Sprintf("%v 【%v】scan port total:%v，goroutine count:%v，peer goroutine handle count:%v，timeout:%vms", time.Now().Format("2006-01-02 15:04:05"), ip, total, pageCount, num, s.Timeout))
	start := time.Now()
	all := map[int][]int{}
	for i := 1; i <= pageCount; i++ {
		for j := 0; j < num; j++ {
			tmp := (i-1)*num + j
			if tmp < total {
				all[i] = append(all[i], ports[tmp])
			}
		}
	}

	wg := sync.WaitGroup{}
	for k, v := range all {
		wg.Add(1)
		go func(value []int, key int) {
			defer wg.Done()
			var tmpPorts []int
			for i := 0; i < len(value); i++ {
				opened := s.isOpen(ip, value[i])
				if opened {
					tmpPorts = append(tmpPorts, value[i])
				}
			}
			mutex.Lock()
			openPorts = append(openPorts, tmpPorts...)
			mutex.Unlock()
			if len(tmpPorts) > 0 {
				l.Info(fmt.Sprintf("%v 【%v】goroutine%v complete，cost time： %.3fs，open ports： %v", time.Now().Format("2006-01-02 15:04:05"), ip, key, time.Since(start).Seconds(), tmpPorts))
			}
		}(v, k)
	}
	wg.Wait()

	sort.Ints(openPorts)

	l.Info(fmt.Sprintf("%v 【%v】scan finish，cost time%.3fs , open ports:%v", time.Now().Format("2006-01-02 15:04:05"), ip, time.Since(start).Seconds(), openPorts))

	return openPorts
}

// IsCIDR check if the string is an valid CIDR notiation (IPV4 & IPV6)
func isCIDR(str string) bool {
	_, _, err := net.ParseCIDR(str)
	return err == nil
}

// IsIP checks if a string is either IP version 4 or 6. Alias for `net.ParseIP`
func isIP(str string) bool {
	return net.ParseIP(str) != nil
}

// IsDNSName will validate the given string as a DNS name
func isDNSName(str string) bool {
	if str == "" || len(strings.Replace(str, ".", "", -1)) > 255 {
		// constraints already violated
		return false
	}
	return !isIP(str) && rxDNSName.MatchString(str)
}

func parseDNS(domain string) (string, error) {
	addr, err := net.ResolveIPAddr("ip", domain)
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}

// cidr解析
func parseCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
