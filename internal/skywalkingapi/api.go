// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalkingapi handle SkyWalking tracing metrics.
package skywalkingapi

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	agentv3 "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/skywalking/compiled/language/agent/v3"
	"google.golang.org/protobuf/proto"
)

/*
该包主要解决 SkyWalking 在处理 trace, metric, logging, profiling 等数据时
重复引用的问题。
所以提取出公共函数
*/

type SkyAPI struct {
	inputName      string
	plugins        []string
	afterGatherRun itrace.AfterGatherHandler
	customerKeys   []string
	tags           map[string]string
	localCache     *storage.Storage
	log            *logger.Logger
}

// InitApiPluginAges 可以初始化多次, 用 name 区分.
//nolint:gofumpt,stylecheck
func InitApiPluginAges(pls []string, localCacheConfig *storage.StorageConfig, closeResource map[string][]string,
	keepRareResource bool, sampler *itrace.Sampler, customerTags []string, itags map[string]string, name string) *SkyAPI {
	api := &SkyAPI{inputName: name, plugins: pls, tags: itags}
	api.log = logger.SLogger(name)
	if localCacheConfig != nil {
		if localCache, err := storage.NewStorage(localCacheConfig, api.log); err != nil {
			api.log.Errorf("### new local-cache failed: %s", err.Error())
		} else {
			api.localCache = localCache
			api.localCache.RegisterConsumer(storage.SKY_WALKING_GRPC_KEY, func(buf []byte) error {
				start := time.Now()
				segobj := &agentv3.SegmentObject{}
				if err := proto.Unmarshal(buf, segobj); err != nil {
					return err
				}
				api.parseSegmentObject(segobj)

				api.log.Debugf("### process status: buffer-size: %dkb, cost: %dms, err: %v", len(buf)>>10, time.Since(start)/time.Millisecond, err)

				return nil
			})
			if err = localCache.RunConsumeWorker(); err != nil {
				api.log.Errorf("### run local-cache consumer failed: %s", err.Error())
			}
		}
	}

	var afterGather *itrace.AfterGather
	if api.localCache != nil && api.localCache.Enabled() {
		afterGather = itrace.NewAfterGather(itrace.WithLogger(api.log), itrace.WithRetry(100*time.Millisecond), itrace.WithBlockIOModel(true))
	} else {
		afterGather = itrace.NewAfterGather(itrace.WithLogger(api.log))
	}
	api.afterGatherRun = afterGather

	if len(closeResource) != 0 {
		iCloseResource := &itrace.CloseResource{}
		iCloseResource.UpdateIgnResList(closeResource)
		afterGather.AppendFilter(iCloseResource.Close)
	}

	// add error status penetration
	afterGather.AppendFilter(itrace.PenetrateErrorTracing)
	// add rare resource keeper
	if keepRareResource {
		krs := &itrace.KeepRareResource{}
		krs.UpdateStatus(keepRareResource, time.Hour)
		afterGather.AppendFilter(krs.Keep)
	}

	// add sampler
	var isampler *itrace.Sampler
	if sampler != nil && (sampler.SamplingRateGlobal >= 0 && sampler.SamplingRateGlobal <= 1) {
		isampler = sampler
	} else {
		isampler = &itrace.Sampler{SamplingRateGlobal: 1}
	}
	afterGather.AppendFilter(isampler.Sample)

	api.customerKeys = customerTags

	return api
}

func (api *SkyAPI) StopStorage() {
	if api.localCache != nil {
		_ = api.localCache.Close()
	}
}

func (api *SkyAPI) CloseLocalCache() {
	if api.localCache != nil {
		if err := api.localCache.Close(); err != nil {
			api.log.Error(err.Error())
		}
	}
}
