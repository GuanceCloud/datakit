// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package exporter collect RealTime data.
package exporter

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

var moduleVersion = "v5"

// ExporterV5 : V4 和 V5 的采集方式是一样的，只不过结构体不同，可共用一个结构体.
type ExporterV5 struct {
	ExDir         string `toml:"export_dir"`
	moduleVersion string `toml:"module_version"` // v4|v5 或者更高。
	items         map[string]*FileReader
	metricChan    chan []*point.Point
	stopChan      chan struct{}
	cacheData     *CacheData

	CollectItem,
	CollectTrigger,
	CollectTrends bool // 三种数据类型开关。

	feeder dkio.Feeder
}

func (ex *ExporterV5) InitExporter(feeder dkio.Feeder, tags map[string]string, cd *CacheData, objects string) error {
	ex.items = make(map[string]*FileReader)
	ex.metricChan = make(chan []*point.Point, 10)
	ex.stopChan = make(chan struct{})
	ex.feeder = feeder
	ex.cacheData = cd
	if ex.moduleVersion == "" {
		ex.moduleVersion = "v5"
	}
	moduleVersion = ex.moduleVersion
	stat, err := os.Stat(ex.ExDir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("ExporterDir:%s is not dir", ex.ExDir)
	}

	collectModules := strings.Split(objects, ",")
	if len(collectModules) == 3 || objects == "" {
		ex.CollectTrigger = true
		ex.CollectTrends = true
		ex.CollectItem = true
	} else {
		for _, module := range collectModules {
			switch strings.ToLower(module) {
			case modules[Items]:
				ex.CollectItem = true
			case modules[Trigger]:
				ex.CollectTrigger = true
			case modules[Trends]:
				ex.CollectTrends = true
			default:
				log.Errorf("unknown object =%s", module)
			}
		}
	}
	g := goroutine.NewGroup(goroutine.Option{Name: inputName})
	g.Go(func(ctx context.Context) error {
		ex.checkFileList(tags)
		return nil
	})

	return nil
}

func (ex *ExporterV5) checkFileList(tags map[string]string) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ex.stopChan:
			return
		case <-ticker.C:
			list, err := os.ReadDir(ex.ExDir)
			if err != nil {
				log.Errorf("read exportDir err=%v", err)
				return
			}
			for _, entry := range list {
				if entry.IsDir() {
					continue
				}
				fullName := filepath.Join(ex.ExDir, entry.Name())
				_, exist := ex.items[fullName]

				if !exist {
					model := CheckModel(entry.Name())
					switch model {
					case Items:
						if ex.CollectItem {
							log.Infof("add items file to export, full name =%s", fullName)
							ZabbixCollectFiles.WithLabelValues("item").Add(1)
							ex.items[fullName] = NewFileReader(fullName, Items, tags)

							g := goroutine.NewGroup(goroutine.Option{Name: inputName})
							g.Go(func(ctx context.Context) error {
								ex.items[fullName].Read(ex.metricChan, ex.cacheData)
								return nil
							})
							time.Sleep(time.Second) // 文件多，错峰读取 items 文件。
						}
					case Trends:
						if ex.CollectTrends {
							log.Infof("add trends file to export, full name =%s", fullName)
							ZabbixCollectFiles.WithLabelValues("trends").Add(1)
							ex.items[fullName] = NewFileReader(fullName, Trends, tags)

							g := goroutine.NewGroup(goroutine.Option{Name: inputName})
							g.Go(func(ctx context.Context) error {
								ex.items[fullName].Read(ex.metricChan, ex.cacheData)
								return nil
							})
						}

					case Trigger, Unknown:
					default:
						log.Infof("unknown mode, filename=%s", fullName)
					}
				}
			}
		}
	}
}

func (ex *ExporterV5) collect() {
	var err error
	log.Infof("start collect point from read file.")

	for {
		select {
		case pts := <-ex.metricChan:
			if len(pts) > 0 {
				err = ex.feeder.FeedV2(point.Metric, pts, dkio.WithInputName(inputName))
				if err != nil {
					log.Errorf("feed pts err=%v", err)
				}
			}
		case <-ex.stopChan:
			for _, item := range ex.items {
				close(item.stop)
			}
			log.Infof("exporter stop")
			return
		}
	}
}

type FileReader struct {
	exportType ExportType
	fileName   string
	tags       map[string]string
	log        *logger.Logger
	lines      chan string

	stop      chan interface{}
	firstOpen bool
}

func NewFileReader(fileName string, eType ExportType, tags map[string]string) *FileReader {
	logName := "exporter_file"
	_, name := filepath.Split(fileName)
	strs := strings.Split(name, ".")
	if len(strs) == 2 {
		logName = strs[0]
	}
	return &FileReader{
		fileName:   fileName,
		exportType: eType,
		tags:       tags,
		log:        logger.SLogger(logName),
		stop:       make(chan interface{}),
		firstOpen:  true,
		lines:      make(chan string, 100),
	}
}

func (fr *FileReader) Read(feeder chan []*point.Point, cd *CacheData) {
	fr.log.Infof("start to read file:%s", fr.fileName)
	ticker := time.NewTicker(time.Second * 10)
	itemsC := make([]string, 0, 20)
	trendsC := make([]string, 0, 20)

	g := goroutine.NewGroup(goroutine.Option{Name: inputName})
	g.Go(func(ctx context.Context) error {
		fr.readFromFileV2()
		return nil
	})

	defer ticker.Stop()

	for {
		select {
		case <-fr.stop:
			return
		case text := <-fr.lines:
			switch fr.exportType {
			case Items:
				log.Debugf("add item data to cache")
				itemsC = append(itemsC, text)
			case Trends:
				trendsC = append(trendsC, text)
			case Trigger, Unknown:
			default:
			}

		case <-ticker.C:
			// 将数据组装 points 发送
			if len(itemsC) > 0 {
				var pts []*point.Point
				if moduleVersion == "v5" || moduleVersion == "" {
					pts = itemsToPoints(itemsC, fr.tags, fr.log, cd)
				}
				if moduleVersion == "v4" {
					pts = ItemValuesToPoint(itemsC, fr.tags, fr.log, cd)
				}
				if len(pts) > 0 {
					fr.log.Debugf("read from file:%s to point.len=%d", fr.fileName, len(pts))
					ZabbixCollectMetrics.WithLabelValues("item").Add(1)
					feeder <- pts
				}
				itemsC = make([]string, 0, 20)
			}
			if len(trendsC) > 0 {
				var pts []*point.Point
				if moduleVersion == "v5" || moduleVersion == "" {
					pts = trendsToPoints(trendsC, fr.tags, fr.log, cd)
				}
				if moduleVersion == "v4" {
					pts = trendsValueToPoints(itemsC, fr.tags, fr.log, cd)
				}

				if len(pts) > 0 {
					ZabbixCollectMetrics.WithLabelValues("trends").Add(float64(len(pts)))
					fr.log.Debugf("read from file:%s to point.len=%d", fr.fileName, len(pts))
					feeder <- pts
				}
				trendsC = make([]string, 0, 20)
			}
		}
	}
}

func (fr *FileReader) readFromFileV2() {
	var err error
	fn := func(filename, text string, fields map[string]interface{}) error {
		log.Debugf("read from tailer")
		fr.lines <- text
		return nil
	}

	tailOpts := []tailer.Option{
		tailer.WithForwardFunc(fn),
		tailer.WithFromBeginning(false),
		tailer.WithFileFromBeginningThresholdSize(math.MaxInt64), // 设置最大值，从不开头读。
	}

	tail, err := tailer.NewTailer([]string{fr.fileName}, tailOpts...)
	if err != nil {
		log.Errorf("new single err=%v", err)
		return
	}
	tail.Start()
}
