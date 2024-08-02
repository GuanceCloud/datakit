// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package elasticsearch

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	gcPoint "github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) setIptLastCOInfo(url string, m *customerObjectMeasurement) {
	ipt.CustomerObjectMap[url] = m
}

func (ipt *Input) fixIptLastCOInfo(url string, version string, uptime int) {
	ipt.CustomerObjectMap[url].fields["version"] = version
	ipt.CustomerObjectMap[url].fields["uptime"] = strconv.Itoa(uptime)
}

func getPointsFromMeasurement(ms []inputs.MeasurementV2) []*gcPoint.Point {
	pts := []*gcPoint.Point{}
	for _, m := range ms {
		pts = append(pts, m.Point())
	}

	return pts
}

func buildURL(baseURL, relativePath string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	relativePath = strings.TrimLeft(relativePath, "/")
	return baseURL + "/" + relativePath
}

type CatNode struct {
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	Version string `json:"version"`
	Uptime  int    `json:"uptime"`
}

func fetchNodesData(url string) ([]*CatNode, error) {
	// url := "http://localhost:9200/_cat/nodes?v&h=ip,port,version,uptime".
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			l.Errorf("error closing response body: %v", err)
		}
	}(resp.Body)

	var nodes []*CatNode
	scanner := bufio.NewScanner(resp.Body)
	// Skip the header line
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		// Parse the line into fields
		fields := strings.Fields(line)
		if len(fields) != 4 {
			return nil, fmt.Errorf("unexpected line format: %s", line)
		}

		port, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, err
		}

		node := &CatNode{
			IP:      fields[0],
			Port:    port,
			Version: fields[2],
			Uptime:  parseUptime(fields[3]),
		}
		l.Debugf("Node IP: %s, Port: %d, Version: %s, Uptime: %d seconds\n", node.IP, node.Port, node.Version, node.Uptime)

		nodes = append(nodes, node)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}

func parseUptime(uptime string) int {
	var seconds float64
	fmt.Printf("uptime before is %s\n", uptime)

	if len(uptime) == 0 {
		fmt.Printf("uptime is empty\n")
		return 0
	}

	var multiplier float64
	unit := uptime[len(uptime)-1:]

	switch unit {
	case "s":
		multiplier = 1.0
	case "m":
		multiplier = 60.0
	case "h":
		multiplier = 3600.0
	case "d":
		multiplier = 86400.0
	default:
		fmt.Printf("unknown uptime format: %s\n", uptime)
		return 0
	}

	val, err := strconv.ParseFloat(uptime[:len(uptime)-1], 64)
	if err != nil {
		fmt.Printf("uptime parse error: %v\n", err)
		return 0
	}
	seconds = val * multiplier

	fmt.Printf("uptime is %.0f seconds\n", seconds)
	return int(seconds)
}

func (ipt *Input) getPtsByNodes(nodes []*CatNode) []*gcPoint.Point {
	ms := []inputs.MeasurementV2{}
	for _, node := range nodes {
		url := fmt.Sprintf("%s:%d", node.IP, node.Port)
		fields := map[string]interface{}{
			"display_name": fmt.Sprintf("%s:%d", node.IP, node.Port),
			"version":      node.Version,
			"uptime":       fmt.Sprintf("%d", node.Uptime),
		}
		tags := map[string]string{
			"name":          fmt.Sprintf("%s-%s:%d", inputName, node.IP, node.Port),
			"host":          node.IP,
			"ip":            fmt.Sprintf("%s:%d", node.IP, node.Port),
			"col_co_status": "OK",
		}
		m := &customerObjectMeasurement{
			name:     "database",
			tags:     tags,
			fields:   fields,
			election: ipt.Election,
		}
		l.Debugf("m linepoint:", m.Point().LineProto())
		ms = append(ms, m)

		if ipt.CustomerObjectMap[url] == nil {
			ipt.setIptLastCOInfo(url, m)
		} else {
			ipt.fixIptLastCOInfo(url, node.Version, node.Uptime)
		}
	}
	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		l.Debugf("pts: %v", pts)
		return pts
	}
	return []*gcPoint.Point{}
}

func (ipt *Input) collectCustomerObjectMeasurement() {
	for _, u := range ipt.Servers {
		uu := buildURL(u, "/_cat/nodes?v&h=ip,port,version,uptime")
		nodes, err := fetchNodesData(uu)
		if err != nil {
			l.Errorf("fail to fetch nodes: %s", err.Error())
			continue
		}
		pts := ipt.getPtsByNodes(nodes)
		ipt.FeedCoPts(pts)
	}
}

func (ipt *Input) getCoPointByColErr(u string, errStr string) []*gcPoint.Point {
	var ms []inputs.MeasurementV2
	uu, _ := url.Parse(u)
	h, p, err := net.SplitHostPort(uu.Host)
	var host string
	var port int
	if err == nil {
		host = h
		port, _ = strconv.Atoi(p)
	} else {
		l.Errorf("failed to split host and port: %s", err)
	}
	fields := map[string]interface{}{
		"display_name": fmt.Sprintf("%s:%d", host, port),
	}
	tags := map[string]string{
		"reason":        errStr,
		"name":          fmt.Sprintf("%s-%s:%d", inputName, host, port),
		"host":          host,
		"ip":            fmt.Sprintf("%s:%d", host, port),
		"col_co_status": "NotOK",
	}
	m := &customerObjectMeasurement{
		name:     "database",
		tags:     tags,
		fields:   fields,
		election: ipt.Election,
	}
	ms = append(ms, m)
	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts
	}
	return []*gcPoint.Point{}
}

func (ipt *Input) FeedCoPts(pts []*gcPoint.Point) {
	if err := ipt.feeder.FeedV2(gcPoint.CustomObject, pts,
		dkio.WithCollectCost(time.Since(time.Now())),
		dkio.WithElection(ipt.Election),
		dkio.WithInputName(inputName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(gcPoint.CustomObject),
		)
		l.Errorf("feed : %s", err)
	}
}

func (ipt *Input) FeedCoByErr(url string, err error) {
	pts := ipt.getCoPointByColErr(url, err.Error())
	if err := ipt.feeder.FeedV2(gcPoint.CustomObject, pts,
		dkio.WithCollectCost(time.Since(time.Now())),
		dkio.WithElection(ipt.Election),
		dkio.WithInputName(inputName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(gcPoint.CustomObject),
		)
		l.Errorf("feed : %s", err)
	}
}
