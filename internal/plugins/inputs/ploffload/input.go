// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ploffload used to receive pipeline offload data from DataKit
package ploffload

import (
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"google.golang.org/protobuf/proto"
)

const (
	sampleCfg = `
[inputs.ploffload]

  ## Storage config a local storage space in hard dirver to cache data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.ploffload.storage]
    # path = "./ploffload_storage"
    # capacity = 5120
`
	inputName = "ploffload"

	ginHandleURI = "/v1/write/ploffload/:category"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

var (
	log        = logger.DefaultSLogger(inputName)
	localCache *storage.Storage
)

type Input struct {
	LocalCacheConfig *storage.StorageConfig `toml:"storage"`

	wkr    *workerPool
	feeder dkio.Feeder

	semStop *cliutils.Sem // start stop signal
}

func (ipt *Input) Catalog() string { return "ploffload" }

func (ipt *Input) SampleConfig() string { return sampleCfg }

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *Input) RegHTTPHandler() {
	log = logger.SLogger(inputName)

	var err error
	if ipt.LocalCacheConfig != nil {
		if localCache, err = storage.NewStorage(ipt.LocalCacheConfig, log); err != nil {
			log.Errorf("### new local-cache failed: %s", err.Error())
		} else {
			localCache.RegisterConsumer(storage.HTTP_KEY, func(buf []byte) error {
				start := time.Now()
				reqpb := &storage.Request{}
				if err := proto.Unmarshal(buf, reqpb); err != nil {
					return err
				} else {
					req := convStorageReq2HTTPReq(reqpb)
					ipt.handlePlOffload(&httpapi.NopResponseWriter{}, req)

					log.Debugf("### process status: buffer-size: %dkb, cost: %dms, err: %v", len(reqpb.Body)>>10, time.Since(start)/time.Millisecond, err)

					return nil
				}
			})
			if err = localCache.RunConsumeWorker(); err != nil {
				log.Errorf("### run local-cache consumer failed: %s", err.Error())
			}
		}
	}

	// 创建工作线程池
	if ipt.wkr == nil {
		ipt.wkr = NewWorkerPool(16, ipt.handlePlOffload)
		ipt.wkr.Start()
	}

	httpapi.RegHTTPHandler("POST", ginHandleURI,
		HTTPStorageWrapperWithWkrPool(storage.HTTP_KEY,
			httpStatusRespFunc,
			localCache,
			ipt.wkr,
			httpapi.ProtectedHandlerFunc(ipt.handlePlOffload, log)))
}

// ReadEnv read envs and config input
//
// ENV_INPUT_PLOFFLOAD_STORAGE_PATH : string
//
// ENV_INPUT_PLOFFLOAD_STORAGE_CAPACITY : int.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if ipt.LocalCacheConfig == nil {
		ipt.LocalCacheConfig = &storage.StorageConfig{}
	}

	if storagePath, ok := envs["ENV_INPUT_PLOFFLOAD_STORAGE_PATH"]; ok {
		if storagePath != "" {
			ipt.LocalCacheConfig.Path = storagePath
		}
	}

	if storageCapacity, ok := envs["ENV_INPUT_PLOFFLOAD_STORAGE_CAPACITY"]; ok {
		if storageCapacity != "" {
			capacity, err := strconv.ParseInt(storageCapacity, 10, 64)
			if err != nil {
				log.Warn("invalid ENV_INPUT_PLOFFLOAD_STORAGE_CAPACITY: ", storageCapacity)
				capacity = 5120
			}
			ipt.LocalCacheConfig.Capacity = int(capacity)
		}
	}
}

func (ipt *Input) Run() {
	log.Info("### register ploffload router")
	select {
	case <-datakit.Exit.Wait():
		if ipt.wkr != nil {
			ipt.wkr.Stop()
		}
		ipt.exit()
		log.Info(inputName + " exit")
		return
	case <-ipt.semStop.Wait():
		if ipt.wkr != nil {
			ipt.wkr.Stop()
		}
		ipt.exit()
		log.Info(inputName + " return")
		return
	}
}

func (*Input) exit() {
	if localCache != nil {
		if err := localCache.Close(); err != nil {
			log.Error(err.Error())
		}
		log.Debug("### storage closed")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		feeder:  dkio.DefaultFeeder(),
		semStop: cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
