// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pyroscope

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile/metrics"
	"golang.org/x/exp/maps"
)

var (
	pprofLRU    = expirable.NewLRU[string, *pprofBatch](2048, onEvict, time.Second*15)
	lruLock     sync.Mutex
	fileMapPool = &sync.Pool{
		New: func() interface{} {
			return make(map[string][]byte, 6)
		},
	}
)

type pprofFile struct {
	name    string
	payload []byte
}

type pprofBatch struct {
	pprofFiles map[string][]byte
	metadata   *metrics.Metadata
	header     map[string]string
	ipt        *Input
}

func (ipt *Input) addPProfToCachePool(sessionID string, endTime time.Time, file *pprofFile, metadata metrics.Metadata, header map[string]string) {
	key := fmt.Sprintf("%s:%d", sessionID, endTime.Unix())
	lruLock.Lock()
	defer lruLock.Unlock()

	fileBatch, ok := pprofLRU.Get(key)
	if !ok {
		pprofFiles := fileMapPool.Get().(map[string][]byte)
		pprofFiles[file.name] = file.payload
		metadata.Attachments = []string{file.name}

		if header == nil {
			header = make(map[string]string)
		}
		fileBatch = &pprofBatch{
			pprofFiles: pprofFiles,
			metadata:   &metadata,
			header:     header,
			ipt:        ipt,
		}
		pprofLRU.Add(key, fileBatch)
	} else {
		fileBatch.pprofFiles[file.name] = file.payload
		fileBatch.metadata.Attachments = append(fileBatch.metadata.Attachments, file.name)
		if fileBatch.metadata.Start.After(metadata.Start) {
			fileBatch.metadata.Start = metadata.Start
			fileBatch.metadata.End = metadata.End
		}
		for k, v := range header {
			if _, exists := fileBatch.header[k]; !exists {
				fileBatch.header[k] = v
			}
		}
	}
}

func onEvict(_ string, batch *pprofBatch) {
	if err := batch.ipt.batchSend(batch); err != nil {
		log.Errorf("unable to send pyroscope go pprof files: %v", err)
	}
}

func (ipt *Input) batchSend(batch *pprofBatch) error {
	defer func() {
		if batch.pprofFiles != nil {
			maps.Clear(batch.pprofFiles)
			fileMapPool.Put(batch.pprofFiles)
		}
	}()
	if len(batch.pprofFiles) == 0 {
		return fmt.Errorf("no pprof files")
	}
	if batch.metadata == nil {
		return fmt.Errorf("need metadata")
	}

	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)
	w, err := mw.CreateFormFile(metrics.EventJSONFile, metrics.EventJSONFile)
	if err != nil {
		return fmt.Errorf("unable to create form file: %w", err)
	}
	if err = json.NewEncoder(w).Encode(batch.metadata); err != nil {
		return fmt.Errorf("unable to marshal data for profiling event file: %w", err)
	}

	for name, payload := range batch.pprofFiles {
		w, err = mw.CreateFormFile(name, name)
		if err != nil {
			return fmt.Errorf("unable to create form file: %w", err)
		}
		if _, err := w.Write(payload); err != nil {
			return fmt.Errorf("unable to write pprof file to : %w", err)
		}
	}

	if err = mw.Close(); err != nil {
		return fmt.Errorf("unable to close multipart writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, ipt.profileSendingAPI.String(), buf)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())

	for k, v := range batch.header {
		if k != "Content-Length" && k != "Host" && k != "Content-Type" && req.Header[k] == nil {
			req.Header.Set(k, v)
		}
	}

	return ipt.doSend(req, buf.Bytes(), batch.metadata.Tags)
}
