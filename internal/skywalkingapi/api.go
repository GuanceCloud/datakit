// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package skywalkingapi handle SkyWalking tracing metrics.
package skywalkingapi

import (
	"time"

	cache "gitlab.jiagouyun.com/cloudcare-tools/cliutils/diskcache"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
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
	storage        *itrace.Storage
	log            *logger.Logger
}

// InitApiPluginAges 可以初始化多次, 用 name 区分.
//nolint:gofumpt,stylecheck
func InitApiPluginAges(pls []string, iStorage *itrace.Storage, closeResource map[string][]string,
	keepRareResource bool, sampler *itrace.Sampler, customerTags []string, itags map[string]string, name string) *SkyAPI {
	api := &SkyAPI{inputName: name, plugins: pls, tags: itags}
	api.log = logger.SLogger(name)
	if iStorage != nil {
		if cache, err := cache.Open(iStorage.Path, &cache.Option{Capacity: int64(iStorage.Capacity) << 20}); err != nil {
			api.log.Errorf("### open cache %s with cap %dMB failed, cache.Open: %s", iStorage.Path, iStorage.Capacity, err)
		} else {
			iStorage.SetCache(cache)
			iStorage.RunStorageConsumer(api.log, api.parseSegmentObjectWrapper)
			api.storage = iStorage
			api.log.Infof("### open cache %s with cap %dMB OK", iStorage.Path, iStorage.Capacity)
		}
	}

	var afterGather *itrace.AfterGather
	if api.storage == nil {
		afterGather = itrace.NewAfterGather()
	} else {
		afterGather = itrace.NewAfterGather(itrace.WithRetry(100 * time.Millisecond))
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
	if api.storage != nil {
		_ = api.storage.Close()
	}
}
