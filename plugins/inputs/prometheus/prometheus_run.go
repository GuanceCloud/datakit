package prometheus

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

func mkUrl(hostStr string) string {
	var prefix string = "http://"
	var suffix string = "/metrics"
	if strings.Contains(hostStr, prefix) {
		return fmt.Sprintf("%s%s", hostStr, suffix)
	} else {
		return fmt.Sprintf("%s%s%s", prefix, hostStr, suffix)
	}
}

func (p *PrometheusParam) gather() {
	url := mkUrl(p.input.host)
	for {
		select {
		case <-stopChan:
			p.output.cfun()
			return
		case <-p.output.ctx.Done():
			p.output.cfun()
			return
		default:
		}

		p.getMetrics(url)

		err := internal.SleepContext(p.output.ctx, time.Duration(p.input.interval)*time.Second)
		if err != nil {
			log.Printf("W! [Prometheus] %s", err.Error())
		}
	}
}

func (p *PrometheusParam) getMetrics(url string) {
	rowsCnt := 0
	pointCnt := 0
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("W! [Prometheus] %s", err.Error())
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("W! [Prometheus] Get %s %d", url, resp.StatusCode)
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		lineText := scanner.Text()
		if lineText[0:2] == "# " {
			continue
		}
		rowsCnt += 1

		err := p.parseLine(lineText)
		if err != nil {
			log.Printf("W! [Prometheus] %s Parse %s Err: %s", url, lineText, err.Error())
		} else {
			pointCnt += 1
		}
	}

	log.Printf("I! [Prometheus] %s generate %d points by reading %d lines", url, pointCnt, rowsCnt)
}

func (p *PrometheusParam) parseLine(line string) error {
	var tStr string
	fields := make(map[string]interface{})

	lineSplit := strings.Split(line, " ")
	contSplit := len(lineSplit)

	if contSplit < 2 || contSplit > 3 {
		return fmt.Errorf("split error")
	}

	metic, tags, err := parseMetricTags(lineSplit[0])
	if err != nil {
		return err
	}

	val, err := parseValue(lineSplit[1])
	if err != nil {
		return err
	}
	fields["value"] = val

	if contSplit == 3 {
		tStr = lineSplit[2]
	} else {
		tStr = ""
	}
	t, err := parseTime(tStr)
	if err != nil {
		return err
	}

	pointMetric, err := metric.New(metic, tags, fields, t)
	if err != nil {
		return err
	}
	p.output.acc.AddMetric(pointMetric)
	return nil
}

func parseMetricTags(str string) (string, map[string]string, error) {
	tags := make(map[string]string)
	strLen := len(str)
	index := strings.Index(str, "{")
	if index == -1 {
		return str, tags, nil
	}

	if str[strLen-1] != '}' {
		return "", tags, fmt.Errorf("missed }")
	}

	metric := str[0:index]
	tagStr := str[index+1 : strLen-1]
	tgs := strings.Split(tagStr, ",")
	for _, tg := range tgs {
		if len(tg) == 0 {
			continue
		}

		ts := strings.Split(tg, "=")
		if len(ts) != 2 {
			return "", tags, fmt.Errorf("mismatch %s", tg)
		}
		key := ts[0]
		val := strings.Trim(ts[1], `"`)
		tags[key] = val
	}

	return metric, tags, nil
}

func parseValue(vStr string) (float64, error) {
	val, err := strconv.ParseFloat(vStr, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func parseTime(tStr string) (time.Time, error) {
	if tStr == "" {
		return time.Now(), nil
	}

	val, err := strconv.ParseFloat(tStr, 64)
	if err != nil {
		return time.Unix(0, 0), err
	}
	valInt := int64(val * 1e6)
	return time.Unix(valInt/int64(time.Millisecond), valInt%int64(time.Millisecond)), nil
}
