package baiduIndex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/gjson"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l         *logger.Logger
	inputName = "baiduIndex"
)

func (_ *BaiduIndex) Catalog() string {
	return "baidu"
}

func (_ *BaiduIndex) SampleConfig() string {
	return configSample
}

func (_ *BaiduIndex) Description() string {
	return ""
}

func (_ *BaiduIndex) Gather() error {
	return nil
}

func (_ *BaiduIndex) Init() error {
	return nil
}

func (b *BaiduIndex) Run() {
	l = logger.SLogger("baiduIndex")

	l.Info("baiduIndex input started...")

	if b.initcfg() {
		return
	}

	interval, err := time.ParseDuration(b.Interval)
	if err != nil {
		l.Error(err)
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			b.command()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (b *BaiduIndex) initcfg() bool {
	if b.MetricName == "" {
		b.MetricName = "baiduIndex"
	}

	if b.Cookie == "" {
		l.Error("cookie is required")
		return true
	}

	return false
}

type searchWord struct {
	Name     string `json:"name"`
	WordType int    `json:"wordType"`
}

func (b *BaiduIndex) command() {
	go b.getSearchIndex()
	go b.getExtendedIndex("feed")
	go b.getExtendedIndex("news")
}

func (b *BaiduIndex) getSearchIndex() {
	var lines [][]byte
	var keywords = [][]*searchWord{}
	var et = time.Now()
	var st = et.Add(-time.Duration(24 * time.Hour))
	var startDate = st.Format(`2006-01-02`)
	var endDate = et.Format(`2006-01-02`)

	for _, word := range b.Keywords {
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

	_, resp := Get(path, b.Cookie)

	uniqid := gjson.Parse(resp).Get("data").Get("uniqid").String()
	key := getKey(uniqid, b.Cookie)

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

		tagsSearch := map[string]string{}
		fieldsSearch := map[string]interface{}{}

		tagsSearch["keyword"] = word
		tagsSearch["type"] = "search"
		tagsSearch["device"] = "all"

		fieldsSearch["index"] = ConvertToFloat(allIndex)
		fieldsSearch["avg"] = allAvg
		fieldsSearch["yoy"] = allYoy
		fieldsSearch["qoq"] = allQoq

		pt, err := io.MakeMetric(b.MetricName, tagsSearch, fieldsSearch, time.Now())
		if err != nil {
			l.Errorf("make metric point error %s", err)
		}

		err = io.NamedFeed([]byte(pt), datakit.Metric, inputName)
		if err != nil {
			l.Errorf("push metric point error %s", err)
		}

		tagsPc := map[string]string{}
		fieldsPc := map[string]interface{}{}

		tagsPc["device"] = "pc"

		pc := item.Get("pc").Get("data").String()
		pcAvgKey := fmt.Sprintf("data.generalRatio.%d.pc.avg", idx)
		pcYoyKey := fmt.Sprintf("data.generalRatio.%d.pc.yoy", idx)
		pcQoqKey := fmt.Sprintf("data.generalRatio.%d.pc.qoq", idx)

		pcIndex := decrypt(key, pc)

		pcAvg := gjson.Get(resp, pcAvgKey).Int()
		pcYoy := gjson.Get(resp, pcYoyKey).Int()
		pcQoq := gjson.Get(resp, pcQoqKey).Int()

		fieldsPc["index"] = ConvertToFloat(pcIndex)
		fieldsPc["avg"] = pcAvg
		fieldsPc["yoy"] = pcYoy
		fieldsPc["qoq"] = pcQoq

		pt2, err := io.MakeMetric(b.MetricName, tagsPc, fieldsPc, time.Now())
		if err != nil {
			l.Errorf("make metric point error %s", err)
		}

		err = io.NamedFeed([]byte(pt2), datakit.Metric, inputName)
		if err != nil {
			l.Errorf("push metric point error %s", err)
		}

		tagsWise := map[string]string{}
		fieldsWise := map[string]interface{}{}

		tagsWise["device"] = "wise"

		wise := item.Get("wise").Get("data").String()
		wiseIndex := decrypt(key, wise)

		wiseAvgKey := fmt.Sprintf("data.generalRatio.%d.wise.avg", idx)
		wiseYoyKey := fmt.Sprintf("data.generalRatio.%d.wise.yoy", idx)
		wiseQoqKey := fmt.Sprintf("data.generalRatio.%d.wise.qoq", idx)

		wiseAvg := gjson.Get(resp, wiseAvgKey).Int()
		wiseYoy := gjson.Get(resp, wiseYoyKey).Int()
		wiseQoq := gjson.Get(resp, wiseQoqKey).Int()

		fieldsWise["index"] = ConvertToFloat(wiseIndex)
		fieldsWise["avg"] = wiseAvg
		fieldsWise["yoy"] = wiseYoy
		fieldsWise["qoq"] = wiseQoq

		pt3, err := io.MakeMetric(b.MetricName, tagsWise, fieldsWise, time.Now())
		if err != nil {
			l.Errorf("make metric point error %s", err)
		}

		lines = append(lines, pt)

		err = io.NamedFeed([]byte(pt3), datakit.Metric, inputName)
		if err != nil {
			l.Errorf("push metric point error %s", err)
		}
	}

	b.resData = bytes.Join(lines, []byte("\n"))
}

func (b *BaiduIndex) getExtendedIndex(tt string) {
	var path = ""
	var keywords = [][]*searchWord{}
	var et = time.Now()
	var st = et.Add(-time.Duration(24 * time.Hour))
	var startDate = st.Format(`2006-01-02`)
	var endDate = et.Format(`2006-01-02`)

	for _, word := range b.Keywords {
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

	_, resp := Get(path, b.Cookie)

	uniqid := gjson.Parse(resp).Get("data").Get("uniqid").String()
	key := getKey(uniqid, b.Cookie)

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

		pt, err := io.MakeMetric(b.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %s", err)
		}

		err = io.NamedFeed([]byte(pt), datakit.Metric, inputName)
		if err != nil {
			l.Errorf("push metric point error %s", err)
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &BaiduIndex{}
	})
}
