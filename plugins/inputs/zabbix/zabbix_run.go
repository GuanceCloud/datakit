package zabbix

import (
	"fmt"
	"log"
	"time"
	"context"
	"strconv"
	"strings"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	influxm "github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type ZabbixInput struct {
	provider      string
	address       string
	tablename     string
	interval      int
	hoursperbatch int
}

type ZabbixParam struct {
	input  ZabbixInput
	output ZabbixOutput
}

func (p *ZabbixParam) gather() error {
	for {
		select {
		case <-stopChan:
			p.output.cfun()
			return nil
		case <-p.output.ctx.Done():
			p.output.cfun()
			return context.Canceled
		default:
		}

		err := p.gatherData()
		if err != nil {
			log.Printf("W! [Zabbix] %s", err.Error())
			return err
		}

		err = internal.SleepContext(p.output.ctx, time.Duration(p.input.interval)*time.Second)
		if err != nil {
			log.Printf("W! [Zabbix] %s", err.Error())
			return err
		}
	}
	return nil
}

func (p *ZabbixParam) gatherData() error {

	var infoLogs []string
	var currTable string = p.input.tablename
	var currTableForLog string = RightPad(currTable, " ", 12-len(currTable))

	// read registry
	if err := ReadRegistry(registryPath, &mapTables); err != nil {
		return err
	}

	// set times
	starttimereg := GetValueFromKey(mapTables, currTable)
	startimerfc, err := time.Parse("2006-01-02T15:04:05", starttimereg)
	if err != nil {
		startimerfc, err = time.Parse(time.RFC3339, starttimereg)
		if err != nil {
			return err
		}
	}
	var starttimestr string = strconv.FormatInt(startimerfc.Unix(), 10)
	var endtimetmp time.Time = startimerfc.Add(time.Hour * time.Duration(p.input.hoursperbatch))
	var endtimestr string = strconv.FormatInt(endtimetmp.Unix(), 10)

	//
	// <--  Extract
	//
	infoLogs = append(infoLogs,
		fmt.Sprintf(
			"----------- | %s | [%v --> %v[",
			currTableForLog,
			startimerfc.Format("2006-01-02 15:04:00"),
			endtimetmp.Format("2006-01-02 15:04:00")))

	//start watcher
	startwatch := time.Now()
	ext := NewExtracter(
		p.input.provider,
		p.input.address,
		currTable,
		starttimestr,
		endtimestr)

	if err := ext.Extract(); err != nil {
		return err
	}

	// count rows
	var rowcount int = len(ext.Result)
	infoLogs = append(infoLogs,
		fmt.Sprintf(
			"<-- Extract | %s | %v rows in %s",
			currTableForLog,
			rowcount,
			time.Since(startwatch)))

	// set max clock time
	var maxclock time.Time = startimerfc
	if ext.Maxclock.IsZero() == false {
		maxclock = ext.Maxclock
	}
	// no row
	if rowcount == 0 {
		infoLogs = append(infoLogs,
			fmt.Sprintf(
				"--> Load    | %s | No data",
				currTableForLog))
	} else {
		startwatch = time.Now()
		inlineData := ""
		inlineData = strings.Join(ext.Result[:], "\n")

		pts, _ := ParseInfluxPts([]byte(inlineData))
		err := p.output.ProcessPts(pts)
		if err != nil {
			log.Printf("W! [Zabbix] %s", err.Error())
		}

		infoLogs = append(infoLogs,
			fmt.Sprintf(
				"--> Load    | %s | %v rows in %s",
				currTableForLog,
				rowcount,
				time.Since(startwatch)))
	}

	infoLogs = append(infoLogs,
		fmt.Sprintf("--- Waiting | %s | %v sec ",
			currTableForLog,
			p.input.interval))
	// print all log messages
	log.Printf("I! [%s] %s\n", pluginName, infoLogs)

	// Save in registry
	saveMaxTime(currTable, startimerfc, maxclock, p.input.hoursperbatch)

	return nil
}

func ParseInfluxPts(data []byte) ([]*influxdb.Point, int) {
	points, err := influxm.ParsePoints(data)
	if err != nil {
		return nil, 0
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
		pt, err := influxdb.NewPoint(
			measurement,
			mapTags,
			fields,
			pt.Time(), //job.Ts, //m1,
		)

		if err != nil {
			continue
		}
		counts++
		bp = append(bp, pt)
	}
	return bp, counts
}

func saveMaxTime(tablename string, starttime time.Time, maxtime time.Time, duration int) {

	var timetosave time.Time

	// if maxtime is greater than now, keep the maxclock returned
	if (starttime.Add(time.Hour * time.Duration(duration))).After(time.Now()) {
		timetosave = maxtime
	} else {
		timetosave = starttime.Add(time.Hour * time.Duration(duration))
	}
	//fmt.Printf("|| %s   %s  %s\n", starttime.Format(time.RFC3339), maxtime.Format(time.RFC3339), timetosave.Format(time.RFC3339))
	SaveRegistry(registryPath, tablename, timetosave.Format(time.RFC3339))
}
