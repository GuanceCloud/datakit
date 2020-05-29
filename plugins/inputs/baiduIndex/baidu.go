package baiduIndex

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/gjson"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type BaiduIndex struct {
	Baidu       []*Baidu
	ctx         context.Context
	cancelFun   context.CancelFunc
	accumulator telegraf.Accumulator
	logger      *models.Logger

	runningInstances []*runningInstance
}

type runningInstance struct {
	cfg    *Baidu
	agent  *BaiduIndex
	logger *models.Logger
	// client     *rds.Client
	metricName string
}

func (_ *BaiduIndex) Catalog() string {
	return "baidu"
}

func (_ *BaiduIndex) SampleConfig() string {
	return configSample
}

func (_ *BaiduIndex) Description() string {
	return ""
}

func (_ *BaiduIndex) Gather(telegraf.Accumulator) error {
	return nil
}

func (_ *BaiduIndex) Init() error {
	return nil
}

func (a *BaiduIndex) Start(acc telegraf.Accumulator) error {
	a.logger = &models.Logger{
		Name: `baiduIndex`,
	}

	if len(a.Baidu) == 0 {
		a.logger.Warnf("no configuration found")
		return nil
	}

	a.logger.Infof("starting...")

	a.accumulator = acc

	for _, instCfg := range a.Baidu {
		r := &runningInstance{
			cfg:    instCfg,
			agent:  a,
			logger: a.logger,
		}
		r.metricName = instCfg.MetricName
		if r.metricName == "" {
			r.metricName = "baiduIndex"
		}

		if r.cfg.Interval.Duration == 0 {
			r.cfg.Interval.Duration = time.Minute * 10
		}

		a.runningInstances = append(a.runningInstances, r)

		go r.run(a.ctx)
	}
	return nil
}

func (a *BaiduIndex) Stop() {
	a.cancelFun()
}

func (r *runningInstance) run(ctx context.Context) error {
	defer func() {
		if e := recover(); e != nil {

		}
	}()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		r.command()

		internal.SleepContext(ctx, r.cfg.Interval.Duration)
	}
}

type searchWord struct {
	Name     string `json:"name"`
	WordType int    `json:"wordType"`
}

func (r *runningInstance) command() {
	go r.getSearchIndex()
	go r.getExtendedIndex("feed")
	go r.getExtendedIndex("news")
}

func (r *runningInstance) getSearchIndex() {
	var keywords = [][]*searchWord{}
	var et = time.Now()
	var st = et.Add(-time.Duration(24 * time.Hour))
	var startDate = st.Format(`2006-01-02`)
	var endDate = et.Format(`2006-01-02`)

	for _, word := range r.cfg.Keywords {
		st := &searchWord{
			Name:     word,
			WordType: 1,
		}

		keyWord := []*searchWord{st}

		keywords = append(keywords, keyWord)
	}

	data, _ := json.Marshal(keywords)
	keywordStr := string(data)

	path := fmt.Sprintf("http://index.baidu.com/api/SearchApi/index?area=0&word=%s&startDate=%v&endDate=%v", keywordStr, startDate, endDate)

	_, resp := Get(path, r.cfg.Cookie)

	uniqid := gjson.Parse(resp).Get("data").Get("uniqid").String()
	key := getKey(uniqid, r.cfg.Cookie)

	for idx, item := range gjson.Parse(resp).Get("data").Get("userIndexes").Array() {
		all := item.Get("all").Get("data").String()
		allIndex := decrypt(key, all)
		allAvgKey := fmt.Sprintf("data.generalRatio.%d.all.avg", idx)
		allYoyKey := fmt.Sprintf("data.generalRatio.%d.all.yoy", idx)
		allQoqKey := fmt.Sprintf("data.generalRatio.%d.all.qoq", idx)

		allAvg := gjson.Get(resp, allAvgKey).Int()

		allYoy := gjson.Get(resp, allYoyKey).Int()
		allQoq := gjson.Get(resp, allQoqKey).Int()

		word := item.Get("word.0").Get("name").String()

		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["keyword"] = word
		tags["type"] = "search"
		tags["device"] = "all"

		fields["index"] = ConvertToFloat(allIndex)
		fields["avg"] = allAvg
		fields["yoy"] = allYoy
		fields["qoq"] = allQoq

		r.agent.accumulator.AddFields(r.metricName, fields, tags)

		tags["device"] = "pc"

		pc := item.Get("pc").Get("data").String()
		pcAvgKey := fmt.Sprintf("data.generalRatio.%d.pc.avg", idx)
		pcYoyKey := fmt.Sprintf("data.generalRatio.%d.pc.yoy", idx)
		pcQoqKey := fmt.Sprintf("data.generalRatio.%d.pc.qoq", idx)

		pcIndex := decrypt(key, pc)

		pcAvg := gjson.Get(resp, pcAvgKey).Int()
		pcYoy := gjson.Get(resp, pcYoyKey).Int()
		pcQoq := gjson.Get(resp, pcQoqKey).Int()

		fields["index"] = ConvertToFloat(pcIndex)
		fields["avg"] = pcAvg
		fields["yoy"] = pcYoy
		fields["qoq"] = pcQoq

		r.agent.accumulator.AddFields(r.metricName, fields, tags)

		tags["device"] = "wise"

		wise := item.Get("wise").Get("data").String()
		wiseIndex := decrypt(key, wise)

		wiseAvgKey := fmt.Sprintf("data.generalRatio.%d.wise.avg", idx)
		wiseYoyKey := fmt.Sprintf("data.generalRatio.%d.wise.yoy", idx)
		wiseQoqKey := fmt.Sprintf("data.generalRatio.%d.wise.qoq", idx)

		wiseAvg := gjson.Get(resp, wiseAvgKey).Int()
		wiseYoy := gjson.Get(resp, wiseYoyKey).Int()
		wiseQoq := gjson.Get(resp, wiseQoqKey).Int()

		fields["index"] = ConvertToFloat(wiseIndex)
		fields["avg"] = wiseAvg
		fields["yoy"] = wiseYoy
		fields["qoq"] = wiseQoq
		r.agent.accumulator.AddFields(r.metricName, fields, tags)
	}
}

func (r *runningInstance) getExtendedIndex(tt string) {
	var path = ""
	var keywords = [][]*searchWord{}
	var et = time.Now()
	var st = et.Add(-time.Duration(24 * time.Hour))
	var startDate = st.Format(`2006-01-02`)
	var endDate = et.Format(`2006-01-02`)

	for _, word := range r.cfg.Keywords {
		st := &searchWord{
			Name:     word,
			WordType: 1,
		}

		keyWord := []*searchWord{st}

		keywords = append(keywords, keyWord)
	}

	data, _ := json.Marshal(keywords)
	keywordStr := string(data)

	if tt == "feed" {
		path = fmt.Sprintf("http://index.baidu.com/api/FeedSearchApi/getFeedIndex?area=0&word=%s&startDate=%v&endDate=%v", keywordStr, startDate, endDate)
	} else {
		path = fmt.Sprintf("http://index.baidu.com/api/NewsApi/getNewsIndex?area=0&word=%s&startDate=%v&endDate=%v", keywordStr, startDate, endDate)
	}

	_, resp := Get(path, r.cfg.Cookie)

	uniqid := gjson.Parse(resp).Get("data").Get("uniqid").String()
	key := getKey(uniqid, r.cfg.Cookie)

	for _, item := range gjson.Parse(resp).Get("data").Get("index").Array() {
		data := item.Get("data").String()
		avg := item.Get("data.generalRatio.avg").Int()
		yoy := item.Get("data.generalRatio.avg").Int()
		qoq := item.Get("data.generalRatio.avg").Int()

		index := decrypt(key, data)
		word := item.Get("key.0").Get("name").String()

		tags := map[string]string{}
		fields := map[string]interface{}{}

		tags["keyword"] = word
		tags["type"] = tt

		fields["index"] = ConvertToFloat(index)
		fields["avg"] = avg
		fields["yoy"] = yoy
		fields["qoq"] = qoq

		r.agent.accumulator.AddFields(r.metricName, fields, tags)
	}
}

func init() {
	inputs.Add("baiduIndex", func() inputs.Input {
		ac := &BaiduIndex{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
