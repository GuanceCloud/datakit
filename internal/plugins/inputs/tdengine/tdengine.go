// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tdengine is input for tdengine database
package tdengine

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/araddon/dateparse"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// restResult 请求 tdengine 返回的数据结构.
type restResult struct {
	// 请求状态 200~300：ok ， 300~5XX 错误.
	Status string `json:"status"`

	// 逐步丢弃 以columnMata 为主，因为 head 数组并不能表示列的类型和长度属性.
	Head []string `json:"head"`

	// 表头列表说明：[[id,3,2],[end_point 8 134],...] (列名，类型，长度).
	ColumnMeta [][]interface{} `json:"column_meta"`

	// columnMata 对应的值.
	Data [][]interface{} `json:"data"`

	// 返回的行数.
	Rows int `json:"rows"`
}

type tdEngine struct {
	user, pw string
	adapter  string
	basic    string
	stop     chan struct{}
	upstream bool

	Ipt *Input
}

func newTDEngine(ipt *Input) *tdEngine {
	if !strings.HasPrefix(ipt.AdapterEndpoint, "http://") && !strings.HasPrefix(ipt.AdapterEndpoint, "https://") {
		ipt.AdapterEndpoint = "http://" + ipt.AdapterEndpoint
	}
	return &tdEngine{
		user:     ipt.User,
		pw:       ipt.Password,
		adapter:  ipt.AdapterEndpoint,
		basic:    UserToBase64(ipt.User, ipt.Password),
		stop:     make(chan struct{}),
		upstream: true,
		Ipt:      ipt,
	}
}

func (t *tdEngine) Stop() {
	select {
	case <-t.stop:
	default:
		close(t.stop)
	}
}

func (t *tdEngine) CheckHealth(sql selectSQL) bool {
	_, err := query(t.adapter, t.basic, "", []byte(sql.sql))
	return err == nil // When err = nil, TD is health and can subsequent operations.
}

func (t *tdEngine) getSTablesNum(alignTS int64) []*point.Point {
	// show databases.
	databases := t.getDatabase()

	// show $database.stables.
	pts := make([]*point.Point, 0)
	for i := 0; i < len(databases); i++ {
		msm := &Measurement{
			name: "td_database",
			tags: map[string]string{
				"database_name": databases[i].name,
				"created_time":  databases[i].creatTime,
			},
			fields: map[string]interface{}{},
			ts:     alignTS,
		}
		sql := "show " + databases[i].name + ".stables"
		body, err := query(t.adapter, t.basic, "", []byte(sql))
		if err != nil {
			l.Errorf("query data: %v", err)
			continue
		}

		var res restResult
		if err = json.Unmarshal(body, &res); err != nil {
			l.Errorf("parse json error: %v", err)
			return pts
		}

		msm.fields["stable_count"] = res.Rows

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTimestamp(msm.ts))

		if t.Ipt.Election {
			msm.tags = inputs.MergeTagsWrapper(msm.tags, t.Ipt.Tagger.ElectionTags(), t.Ipt.Tags, t.Ipt.AdapterEndpoint)
		} else {
			msm.tags = inputs.MergeTagsWrapper(msm.tags, t.Ipt.Tagger.HostTags(), t.Ipt.Tags, t.Ipt.AdapterEndpoint)
		}

		pt := point.NewPointV2(msm.name,
			append(point.NewTags(msm.tags), point.NewKVs(msm.fields)...),
			opts...)
		pts = append(pts, pt)
	}

	return pts
}

func (t *tdEngine) run() {
	msmC := make(chan []*point.Point, 100)
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_tdengine"})
	for _, m := range metrics {
		func(metric metric) {
			g.Go(func(ctx context.Context) error {
				ticker := time.NewTicker(metric.TimeSeries)
				defer ticker.Stop()

				lastTS := ntp.Now()
				for {
					t.Ipt.alignTS = lastTS.UnixNano()

					select {
					case <-datakit.Exit.Wait():
						return nil
					case <-t.stop:
						return nil
					case tt := <-ticker.C:
						lastTS = inputs.AlignTime(tt, lastTS, metric.TimeSeries)
						if !t.upstream {
							l.Debugf("not leader, skipped")
							continue
						}

						l.Debugf("start to run selectSQL,metricName = %s", metric.metricName)
						for _, sql := range metric.MetricList {
							body, err := query(t.adapter, t.basic, "", []byte(sql.sql))
							if err != nil {
								continue
							}
							var res restResult
							if err := json.Unmarshal(body, &res); err != nil {
								l.Error("parse json error: ", err)
								continue
							}
							if sql.plugInFun != nil {
								msmC <- sql.plugInFun.resToMeasurement(metric.metricName, res, sql, t.Ipt)
							} else {
								msmC <- makeMeasurements(metric.metricName, res, sql, t.Ipt)
							}
						}
					}
				}
			})
		}(m)
		time.Sleep(time.Second / 5) // 交叉运行.
	}

	// show database.stables;
	g.Go(func(ctx context.Context) error {
		ticker := time.NewTicker(time.Minute * 5)
		defer ticker.Stop()

		var lastTS time.Time

		for {
			startTime := ntp.Now()
			lastTS = inputs.AlignTime(startTime, lastTS, (time.Minute * 5))
			select {
			case <-datakit.Exit.Wait():
				return nil
			case <-t.stop:
				return nil
			case <-ticker.C:
				if !t.upstream {
					continue
				}
				l.Debugf("run getSTablesNum")
				msmC <- t.getSTablesNum(lastTS.UnixNano())
			}
		}
	})

	for {
		select {
		case <-t.stop:
			l.Infof("tdengine stop run")
			return
		case msm := <-msmC:
			l.Debugf("measurements receive from channel and len =%d", len(msm))
			if len(msm) > 0 && t.upstream {
				if err := t.Ipt.feeder.FeedV2(point.Metric, msm,
					dkio.WithElection(t.Ipt.Election),
					dkio.WithInputName(inputName),
				); err != nil {
					l.Errorf("FeedMeasurement: %s", err)
				}
			}
		}
	}
}

func UserToBase64(userName, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(userName + ":" + password))
}

type database struct {
	name, creatTime string
}

func (t *tdEngine) getDatabase() []*database {
	// show databases;
	body, err := query(t.adapter, t.basic, "", []byte("show databases;"))
	if err != nil {
		l.Errorf("query data: %v", err)
		return nil
	}
	var res restResult
	if err = json.Unmarshal(body, &res); err != nil {
		l.Error(fmt.Sprint("parse json error: ", err))
		return nil
	}
	var nameIndex int
	var creatIndex int
	for i := 0; i < len(res.ColumnMeta); i++ {
		if res.ColumnMeta[i][0].(string) == "name" {
			nameIndex = i
		}
		if res.ColumnMeta[i][0].(string) == "created_time" {
			creatIndex = i
		}
	}

	databases := make([]*database, 0)
	for i := 0; i < len(res.Data); i++ {
		name := res.Data[i][nameIndex].(string)
		creatTime := res.Data[i][creatIndex].(string)
		databases = append(databases, &database{
			name:      name,
			creatTime: creatTime,
		})
	}
	return databases
}

func query(url string, basicAuth, token string, reqBody []byte) ([]byte, error) {
	var reqBodyBuffer io.Reader = bytes.NewBuffer(reqBody)

	sqlUtcURL := url + "/rest/sql"
	if token != "" {
		sqlUtcURL = sqlUtcURL + "?token=" + token
	}
	req, err := http.NewRequest("POST", sqlUtcURL, reqBodyBuffer)
	if err != nil {
		l.Errorf("query "+url+"/rest/sqlutc error: %v", err)
		return []byte{}, err
	}

	req.Header.Set("Authorization", "Basic "+basicAuth)

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		l.Errorf("query "+url+"/rest/sql error: %v", err)
		return []byte{}, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		l.Errorf("when writing to [%s] received status code: %d", string(reqBody), resp.StatusCode)
		return []byte{}, fmt.Errorf("when writing to [] received status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("when writing to [] received error: %v", err)
		return []byte{}, fmt.Errorf("when writing to [] received error: %w", err)
	}
	return body, nil
}

func makeMeasurements(subMetricName string, res restResult, sql selectSQL, ipt *Input) (pts []*point.Point) {
	host := getHostTagIfNotLoopback(ipt.AdapterEndpoint)
	pts = make([]*point.Point, 0, res.Rows)
	if len(res.Data) == 0 {
		return
	}

	for i := 0; i < len(res.Data); i++ {
		msm := &Measurement{
			tags:     map[string]string{},
			fields:   make(map[string]interface{}),
			ts:       ipt.alignTS,
			election: ipt.Election,
		}
		if host != "" {
			msm.tags["host"] = host
		}
		for j := 0; j < len(res.Data[i]); j++ {
			name := res.ColumnMeta[j][0].(string)
			for _, tag := range sql.tags {
				if name == tag {
					switch t := res.Data[i][j].(type) {
					case nil:
						l.Debugf(" x type is :%T", t)
					case int, int64:
						f := res.Data[i][j].(int64)
						msm.tags[name] = strconv.FormatInt(f, 10)
					case float32, float64:
						f := res.Data[i][j].(float64)
						// 如果是以 `.00` 结尾,那么切掉结尾，取正整数.
						msm.tags[name] = strings.ReplaceAll(strconv.FormatFloat(f, 'f', 2, 64), ".00", "")
					case string:
						msm.tags[name] = res.Data[i][j].(string)
					case bool:
						b := res.Data[i][j].(bool)
						msm.tags[name] = strconv.FormatBool(b)
					default:
						l.Debugf("unknown")
					}
				}
			}
			for _, field := range sql.fields {
				if field == name {
					msm.fields[name] = res.Data[i][j] // set field
				}
			}

			if name == "ts" {
				switch t := res.Data[i][j].(type) {
				case string:
					tsLayout := res.Data[i][j].(string)
					if timeLayout, err := dateparse.ParseFormat(tsLayout); err != nil {
						l.Errorf("ts parse layout error %s", tsLayout)
						continue
					} else {
						parsedTime, _ := time.Parse(timeLayout, tsLayout)
						msm.ts = parsedTime.UnixNano()
					}
				case int64:
					msm.ts = res.Data[i][j].(int64)
				default:
					l.Errorf("unexpected type for ts: %T", t)
				}
			}
		}

		if msm.ts == 0 {
			msm.ts = ntp.Now().UnixNano()
		}

		if sql.unit != "" {
			msm.tags["unit"] = sql.unit
		}
		msm.name = metricName(subMetricName, sql.title)

		opts := point.DefaultMetricOptions()
		opts = append(opts, point.WithTimestamp(msm.ts))

		if ipt.Election {
			msm.tags = inputs.MergeTagsWrapper(msm.tags, ipt.Tagger.ElectionTags(), ipt.Tags, ipt.AdapterEndpoint)
		} else {
			msm.tags = inputs.MergeTagsWrapper(msm.tags, ipt.Tagger.HostTags(), ipt.Tags, ipt.AdapterEndpoint)
		}

		pt := point.NewPointV2(msm.name,
			append(point.NewTags(msm.tags), point.NewKVs(msm.fields)...),
			opts...)
		pts = append(pts, pt)
	}
	return pts
}

func getHostTagIfNotLoopback(u string) string {
	uu, err := url.Parse(u)
	if err != nil {
		l.Errorf("url parse: %v", err)
		return ""
	}
	host, _, err := net.SplitHostPort(uu.Host)
	if err != nil {
		l.Errorf("split host and port: %v", err)
		return ""
	}
	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		return host
	}
	return ""
}
