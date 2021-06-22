package zabbix

import (
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	influxm "github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var tbls = []string{"history", "history_uint", "trends", "trends_uint"}

const (
	MaxGatherInterval = 30 * time.Minute
	MinGatherInterval = 10 * time.Second
)

type ZabbixInput struct {
	Zabbix
}

type ZabbixParam struct {
	input  ZabbixInput
	output ZabbixOutput
	log    *logger.Logger
}

func (z *ZabbixParam) gather() {
	var start, stop time.Time
	var d time.Duration
	var err error

	switch z.input.Interval.(type) {
	case int64:
		d = time.Duration(z.input.Interval.(int64)) * time.Second
	case string:
		d, err = time.ParseDuration(z.input.Interval.(string))
		if err != nil {
			z.log.Errorf("parse interval err: %s", err.Error())
			return
		}
	default:
		z.log.Errorf("interval type unsupported")
		return
	}

	d = config.ProtectedInterval(MinGatherInterval, MaxGatherInterval, d)
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	start = z.getStartDate()
	for {
		select {
		case <-ticker.C:
			stop = z.getStopDate(start)

			for _, tbl := range tbls {
				select {
				case <-datakit.Exit.Wait():
					z.log.Info("input zabbix exit")
					return
				default:
				}

				_, err := z.gatherData(start, stop, tbl, false)
				if err != nil {
					io.FeedLastError(inputName, err.Error())
					z.log.Errorf("gatherData %s", err.Error())
				}
			}

			start = stop
			z.updateRegistry(stop)

		case <-datakit.Exit.Wait():
			z.log.Info("input zabbix exit")
			return
		}
	}
}

func (z *ZabbixParam) gatherData(start, stop time.Time, tblName string, isTest bool) ([]byte, error) {
	starttimestr := strconv.FormatInt(start.Unix(), 10)
	endtimestr := strconv.FormatInt(stop.Unix(), 10)

	ext := NewExtracter(z.input.DbType, z.input.DbAddress, tblName, starttimestr, endtimestr)
	if err := ext.Extract(); err != nil {
		return nil, err
	}

	rowcount := len(ext.Result)
	z.log.Info("table:%s from %s to %s gather %d rows", tblName,
		start.Format(time.RFC3339), stop.Format(time.RFC3339), rowcount)

	if rowcount != 0 {
		inlineData := strings.Join(ext.Result[:], "\n")
		pt := []byte(inlineData)
		if isTest {
			return pt, nil
		}
		pts, _, err := ParseInfluxPts(pt)
		if err != nil {
			return nil, err
		}

		err = z.ProcessPts(pts)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func ParseInfluxPts(data []byte) ([]*influxdb.Point, int, error) {
	points, err := influxm.ParsePoints(data)
	if err != nil {
		return nil, 0, err
	}

	bp := []*influxdb.Point{}
	counts := 0
	for _, pt := range points {
		measurement := string(pt.Name())
		mapTags := map[string]string{}
		for _, tag := range pt.Tags() {
			mapTags[string(tag.Key)] = string(tag.Value)
		}

		fields, _ := pt.Fields()
		pt, err := influxdb.NewPoint(measurement, mapTags, fields, pt.Time())
		if err != nil {
			continue
		}
		counts++
		bp = append(bp, pt)
	}

	return bp, counts, nil
}
