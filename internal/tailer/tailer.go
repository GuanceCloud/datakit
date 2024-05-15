// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tailer wraps logging file collection
package tailer

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/fileprovider"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/openfile"
)

const (
	// 定期寻找符合条件的新文件.
	scanNewFileInterval = time.Second * 10
)

var tailerGoroutine = datakit.G("tailer")

type Tailer struct {
	opt     *option
	options []Option

	fileList map[string]interface{}

	filePatterns   []string
	ignorePatterns []string

	stop chan interface{}
	mu   sync.Mutex

	log *logger.Logger
}

func NewTailer(filePatterns []string, opts ...Option) (*Tailer, error) {
	if len(filePatterns) == 0 {
		return nil, fmt.Errorf("filePatterns cannot be empty")
	}

	c := defaultOption()
	for _, opt := range opts {
		opt(c)
	}
	stop := make(chan interface{})
	if !c.setDone {
		c.done = stop
	}

	return &Tailer{
		opt:            c,
		options:        opts,
		filePatterns:   filePatterns,
		ignorePatterns: c.ignorePatterns,
		stop:           stop,
		fileList:       make(map[string]interface{}),
		log:            logger.SLogger("tailer/" + c.source),
	}, nil
}

func (t *Tailer) Start() {
	ticker := time.NewTicker(scanNewFileInterval)
	defer ticker.Stop()

	for {
		if t.scan() {
			t.log.Infof("all tailers end...")
			_ = tailerGoroutine.Wait()
			t.log.Info("all exit")
			return
		}

		t.log.Debugf("list of recivering: %v", t.getFileList())

		select {
		case <-t.stop:
			t.log.Infof("waiting for all tailers to exit")
			_ = tailerGoroutine.Wait()
			t.log.Info("all exit")
			return

		case <-ticker.C:
		}
	}
}

func (t *Tailer) scan() (ended bool) {
	filelist, err := fileprovider.NewProvider().
		SearchFiles(t.filePatterns).
		IgnoreFiles(t.ignorePatterns).
		Result()
	if err != nil {
		t.log.Warn(err)
	}

	for _, fn := range filelist {
		filename := filepath.Clean(fn)

		if !t.opt.fromBeginning {
			if t.opt.ignoreDeadLog > 0 && !openfile.FileIsActive(filename, t.opt.ignoreDeadLog) {
				continue
			}
		}

		if t.inFileList(filename) {
			continue
		}

		t.log.Infof("new logging file %s with source %s", filename, t.opt.source)

		func(filename string) {
			tailerGoroutine.Go(func(ctx context.Context) error {
				defer t.removeFromFileList(filename)

				tl, err := NewTailerSingle(filename, t.options...)
				if err != nil {
					t.log.Errorf("new tailer file %s error: %s", filename, err)
					return nil
				}

				t.addToFileList(filename)

				tl.Run()
				return nil
			})
		}(filename)
	}

	return false
}

func (t *Tailer) Close() {
	select {
	case <-t.stop:
		return
	default:
		close(t.stop)
	}
}

func (t *Tailer) addToFileList(filename string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fileList[filename] = nil
}

func (t *Tailer) removeFromFileList(filename string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.fileList, filename)
}

func (t *Tailer) inFileList(filename string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.fileList[filename]
	return ok
}

func (t *Tailer) getFileList() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	var list []string
	for filename := range t.fileList {
		list = append(list, filename)
	}
	return list
}
