// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package exporter collect RealTime data.
package exporter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

type ExporterV5 struct {
	ExDir      string `toml:"export_dir"`
	items      map[string]*FileReader
	metricChan chan []*point.Point
	stopChan   chan struct{}

	feeder dkio.Feeder
}

func (ex *ExporterV5) InitExporter(feeder dkio.Feeder, tags map[string]string) error {
	ex.items = make(map[string]*FileReader)
	ex.metricChan = make(chan []*point.Point, 10)
	ex.stopChan = make(chan struct{})
	ex.feeder = feeder

	stat, err := os.Stat(ex.ExDir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("ExporterDir:%s is not dir", ex.ExDir)
	}

	go ex.checkFileList(tags)
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
						log.Infof("add items file to export, full name =%s", fullName)
						ex.items[fullName] = NewFileReader(fullName, Items, tags)
						go ex.items[fullName].Read(ex.metricChan)
						time.Sleep(time.Second) // 文件多，错峰读取 items 文件。
					case Trends:
						log.Infof("add trends file to export, full name =%s", fullName)
						ex.items[fullName] = NewFileReader(fullName, Trends, tags)
						go ex.items[fullName].Read(ex.metricChan)
					case Unknown:
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
	offset     int64
	tags       map[string]string
	log        *logger.Logger

	stop      chan struct{}
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
		stop:       make(chan struct{}),
		firstOpen:  true,
	}
}

func (fr *FileReader) Read(feeder chan []*point.Point) {
	fr.log.Infof("start to read file:%s", fr.fileName)
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-fr.stop:
			return
		case <-ticker.C:
			lines, err := fr.readFromFile()
			if err != nil {
				fr.log.Errorf("read file:%s , err=%v or lines==0", fr.fileName, err)
				continue
			}
			if len(lines) == 0 {
				continue
			}

			switch fr.exportType {
			case Items:
				pts := itemsToPoints(lines, fr.tags, fr.log)
				if len(pts) > 0 {
					fr.log.Debugf("read from file:%s to point.len=%d", fr.fileName, len(pts))
					feeder <- pts
				}
			case Trends:
				pts := trendsToPoints(lines, fr.tags, fr.log)
				if len(pts) > 0 {
					fr.log.Debugf("read from file:%s to point.len=%d", fr.fileName, len(pts))
					feeder <- pts
				}
			case Unknown:
			default:
			}
		}
	}
}

func (fr *FileReader) readFromFile() ([]string, error) {
	lines := make([]string, 0)

	file, err := os.Open(fr.fileName)
	if err != nil {
		return lines, err
	}
	defer file.Close() //nolint

	if fr.firstOpen {
		firstSeek, err := file.Seek(0, io.SeekEnd)
		if err != nil {
			return lines, err
		}
		fr.log.Infof("offset =0 set to %d", firstSeek)
		fr.offset = firstSeek
		fr.firstOpen = false
		return lines, err
	}

	// 如果有seek offset，就从末尾开始读取
	if fr.offset > 0 {
		_, err = file.Seek(fr.offset, io.SeekStart)
		if err != nil {
			fr.log.Warnf("-- file offset change!! file=%s ,offset set to 0(start offset)", fr.fileName)
			fr.offset = 0
			return lines, err
		}
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		bts := scanner.Text()
		lines = append(lines, bts)
	}

	if err = scanner.Err(); err != nil {
		fr.log.Warnf("scanner err=%v", err)
	}

	// 记录当前文件结尾的位置
	newOffset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return lines, err
	}
	fr.offset = newOffset
	fr.log.Debugf("lines len=%d", len(lines))
	return lines, nil
}
