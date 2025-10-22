// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package ddtrace ddtrace apm telemetry
package ddtrace

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	telemetryMeasurementName = "tracing_service"
)

type jvmTelemetry struct {
	lock         sync.RWMutex
	host         Host
	application  Application
	dependencies map[string]Dependency
	traceTime    time.Time
	runtimeID    string

	// kvs    point.KVs
	tags   map[string]string
	fields map[string]interface{}
	change bool
	Name   string
}

func (ob *jvmTelemetry) toPoint() *point.Point {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("toPoint err:%v", err)
		}
	}()
	ob.lock.RLock()
	defer ob.lock.RUnlock()
	opts := point.DefaultObjectOptions()
	opts = append(opts, point.WithTime(ob.traceTime))

	pt := point.NewPoint(telemetryMeasurementName,
		append(point.NewTags(ob.tags), point.NewKVs(ob.fields)...), opts...)
	if pt != nil && len(pt.Warns()) > 0 {
		log.Errorf("toPoint err:%v", pt.Warns())
	}

	return pt
}

func (ob *jvmTelemetry) setField(key string, val interface{}) {
	ob.lock.Lock()
	defer ob.lock.Unlock()

	ob.fields[key] = val
}

func (ob *jvmTelemetry) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   telemetryMeasurementName,
		Desc:   "Collect service, host, process APM telemetry message.",
		DescZh: "采集 DDTrace 的 Service、Host、进程等配置信息",
		Cat:    point.CustomObject,
		Fields: map[string]interface{}{
			requestTypeMap[RequestTypeAppStarted]: &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.String,
				Desc:     "App Started config",
			},
			requestTypeMap[RequestTypeDependenciesLoaded]: &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.String,
				Desc:     "App dependencies loaded",
			},
			requestTypeMap[RequestTypeAppClientConfigurationChange]: &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.String,
				Desc:     "App client configuration change config",
			},
			requestTypeMap[RequestTypeAppIntegrationsChange]: &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.String,
				Desc:     "App Integrations change",
			},
			requestTypeMap[RequestTypeAppClosing]: &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.String,
				Desc:     "App close",
			},
			"spans_created": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Float,
				Unit:     inputs.NCount,
				Desc:     "Create span count",
			},
			"spans_finished": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Float,
				Unit:     inputs.NCount,
				Desc:     "Finish span count",
			},
		},
		Tags: map[string]interface{}{
			"hostname":         inputs.NewTagInfo("Host name"),
			"os":               inputs.NewTagInfo("OS name"),
			"os_version":       inputs.NewTagInfo("OS version"),
			"architecture":     inputs.NewTagInfo("Architecture"),
			"kernel_name":      inputs.NewTagInfo("Kernel name"),
			"kernel_release":   inputs.NewTagInfo("Kernel release"),
			"kernel_version":   inputs.NewTagInfo("Kernel version"),
			"service":          inputs.NewTagInfo("Service"),
			"name":             inputs.NewTagInfo("Same as service name"),
			"env":              inputs.NewTagInfo("Service ENV"),
			"service_version":  inputs.NewTagInfo("Service version"),
			"tracer_version":   inputs.NewTagInfo("DDTrace version"),
			"language_name":    inputs.NewTagInfo("Language name"),
			"language_version": inputs.NewTagInfo("Language version"),
			"runtime_name":     inputs.NewTagInfo("Runtime name"),
			"runtime_version":  inputs.NewTagInfo("Runtime version"),
			"runtime_patches":  inputs.NewTagInfo("Runtime patches"),
			"runtime_id":       inputs.NewTagInfo("Runtime ID"),
		},
	}
}

func (ob *jvmTelemetry) parseEvent(requestType RequestType, payload interface{}) {
	log.Debugf("parse RequestType=%s", string(requestType))

	switch requestType {
	case RequestTypeAppStarted:
		bts, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("err=%v", err)
			return
		}
		start := &AppStarted{}
		err = json.Unmarshal(bts, start)
		if err != nil {
			log.Errorf("can unmarshal to AppStarted err=%v", err)
			return
		}
		tags := getConfigTags(start.Configuration)
		for k, v := range tags {
			k = strings.ReplaceAll(k, ".", "_")
			ob.tags[k] = v
		}

		ob.setField(requestTypeMap[requestType], string(bts))
		log.Debugf("type=%s body=%s", requestType, string(bts))
		ob.change = true
	case RequestTypeDependenciesLoaded:
		bts, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("err=%v", err)
			return
		}

		ds := &Dependencies{}
		err = json.Unmarshal(bts, ds)
		if err != nil {
			log.Errorf("can unmarshal to Dependencies err=%v", err)
			return
		}
		ob.addDependenciesLoaded(ds)

		ob.setField(requestTypeMap[requestType], string(ob.getDependenciesLoaded()))
		log.Debugf("Dependencies type=%s body=%s", requestType, string(bts))
		ob.change = true
	case RequestTypeAppClientConfigurationChange:
		bts, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("err=%v", err)
			return
		}
		configs := &ConfigurationChange{}
		err = json.Unmarshal(bts, configs)
		if err != nil {
			log.Errorf("can unmarshal to AppStarted err=%v", err)
			return
		}
		tags := getConfigTags(configs.Configuration)
		for k, v := range tags {
			k = strings.ReplaceAll(k, ".", "_")
			ob.tags[k] = v
		}
		ob.setField(requestTypeMap[requestType], string(bts))
		ob.change = true
	case RequestTypeAppProductChange,
		RequestTypeAppIntegrationsChange:
		bts, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("err=%v", err)
			return
		}
		ob.setField(requestTypeMap[requestType], string(bts))
		log.Debugf("Dependencies type=%s body=%s", requestType, string(bts))
		ob.change = true
	case RequestTypeAppHeartbeat,
		RequestTypeDistributions:
		// nothing to do.
	case RequestTypeGenerateMetrics:
		bts, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("err=%v", err)
			return
		}
		metrics := &Metrics{}
		err = json.Unmarshal(bts, metrics)
		if err != nil {
			log.Errorf("can unmarshal to AppStarted err=%v", err)
			return
		}
		for _, series := range metrics.Series {
			seriesTag := "_"
			for _, tag := range series.Tags {
				seriesTag += strings.ReplaceAll(tag, ":", "_")
			}

			if len(series.Points) > 0 {
				var m float64
				for _, points := range series.Points {
					m += points[1]
				}
				ob.setField(series.Metric+seriesTag, m)
			}
		}
		ob.change = true
	case RequestTypeAppClosing:
		ob.setField(requestTypeMap[requestType], "service closing")
		ob.change = true
	case RequestTypeMessageBatch:
		bts, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("err=%v", err)
			return
		}
		as := make([]BatchBody, 0)
		err = json.Unmarshal(bts, &as)
		if err != nil {
			log.Errorf("can unmarshal to AppStarted err=%v", err)
			return
		}
		if len(as) > 0 {
			for _, batchBody := range as {
				ob.parseEvent(batchBody.RequestType, batchBody.Payload)
			}
		}
	default:
		log.Warnf("unknown telemetry request type %s", string(requestType))
	}
}

func getConfigTags(configs []Configuration) map[string]string {
	tags := make(map[string]string)
	if len(configs) > 0 {
		for _, conf := range configs {
			if conf.Name == "trace_tags" {
				str, ok := (conf.Value).(string)
				if ok {
					kvsStr := strings.Split(str, ",")
					for _, st := range kvsStr {
						kvs := strings.Split(st, ":")
						if len(kvs) == 2 {
							tags[kvs[0]] = kvs[1]
							setCustomTags([]string{kvs[0]})
						}
					}
				}
			}
		}
	}

	return tags
}

func (ob *jvmTelemetry) addDependenciesLoaded(ds *Dependencies) {
	ob.lock.Lock()
	defer ob.lock.Unlock()
	if ob.dependencies == nil {
		ob.dependencies = map[string]Dependency{}
	}
	if len(ds.Dependencies) == 0 {
		return
	}
	for _, dependency := range ds.Dependencies {
		ob.dependencies[dependency.Name] = dependency
	}
}

func (ob *jvmTelemetry) getDependenciesLoaded() []byte {
	ob.lock.RLock()
	defer ob.lock.RUnlock()
	ds := &Dependencies{Dependencies: make([]Dependency, 0)}
	for _, dependency := range ob.dependencies {
		ds.Dependencies = append(ds.Dependencies, dependency)
	}
	bts, err := json.Marshal(ds)
	if err != nil {
		return nil
	}
	return bts
}

type Manager struct {
	obsLock sync.Mutex
	OBS     map[string]*jvmTelemetry
	OBChan  chan *jvmTelemetry
}

func (ipt *Input) OMInitAndRunning() {
	ipt.om = &Manager{
		OBS:    map[string]*jvmTelemetry{},
		OBChan: make(chan *jvmTelemetry, 10),
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_ddtrace"})
	g.Go(func(ctx context.Context) error {
		for {
			select {
			case ob := <-ipt.om.OBChan:
				pt := ob.toPoint()
				if pt != nil {
					err := ipt.feeder.Feed(point.CustomObject, []*point.Point{pt},
						dkio.WithSource(customObjectFeedName))
					if err != nil {
						log.Errorf("feed err=%v", err)
					}
				}

			case <-ipt.semStop.Wait():
				return nil
			}
		}
	})
}

func (om *Manager) parseTelemetryRequest(header http.Header, bts []byte) {
	om.obsLock.Lock()
	defer om.obsLock.Unlock()
	if requestT := header.Get("Dd-Telemetry-Request-Type"); requestT == "" {
		log.Errorf("request type is null")
		return
	}

	if version := header.Get("Dd-Telemetry-Api-Version"); version != "v2" {
		log.Errorf("request Dd-Telemetry-Api-Version is not v2 :%s", version)
		return
	}
	body := &Body{}
	err := json.Unmarshal(bts, body)
	if err != nil {
		log.Errorf("can not unmarshal to Body err=%v", err)
		return
	}
	// 仅支持 java 的资源目录上报。
	// dd-java-agent 中 LANGUAGE_TAG_VALUE=jvm.
	if body.Application.LanguageName != "jvm" {
		return
	}

	ob, ok := om.OBS[body.Application.ServiceName+body.RuntimeID]
	if !ok {
		tags := make(map[string]string)
		tags["hostname"] = body.Host.Hostname
		tags["os"] = body.Host.OS
		tags["os_version"] = body.Host.OSVersion
		tags["architecture"] = body.Host.Architecture
		tags["kernel_name"] = body.Host.KernelName
		tags["kernel_release"] = body.Host.KernelRelease
		tags["kernel_version"] = body.Host.KernelVersion
		tags["service"] = body.Application.ServiceName
		tags["name"] = body.Application.ServiceName + "-" + body.RuntimeID
		tags["env"] = body.Application.Env
		tags["service_version"] = body.Application.ServiceVersion
		tags["tracer_version"] = body.Application.TracerVersion
		tags["language_name"] = body.Application.LanguageName
		tags["language_version"] = body.Application.LanguageVersion
		tags["runtime_name"] = body.Application.RuntimeName
		tags["runtime_version"] = body.Application.RuntimeVersion
		tags["runtime_patches"] = body.Application.RuntimePatches
		tags["runtime_id"] = body.RuntimeID
		ob = &jvmTelemetry{
			host:        body.Host,
			application: body.Application,
			traceTime:   time.Unix(body.TracerTime, 0),
			runtimeID:   body.RuntimeID,
			tags:        tags,
			fields:      make(map[string]interface{}),
		}
	} else {
		ob.host = body.Host
		ob.application = body.Application
		ob.traceTime = time.Unix(body.TracerTime, 0)
	}

	ob.parseEvent(body.RequestType, body.Payload)
	om.OBS[body.Application.ServiceName+body.RuntimeID] = ob
	// add metric for proxy telemetry body length.
	proxyTelemetryBody.WithLabelValues(body.Application.ServiceName).Observe(float64(len(bts)))
	if ob.change {
		ob.change = false
		// 发生变化，准备发送到io.
		select {
		case om.OBChan <- ob:
		default:
		}
		// 及时清除已经关闭连接的 Agent.
		if _, ok = ob.fields[requestTypeMap[RequestTypeAppClosing]]; ok {
			log.Infof("server:%s stop,and rumtime id is:%s", body.Application.ServiceName, body.RuntimeID)
			delete(om.OBS, body.Application.ServiceName+body.RuntimeID)
		}
	}
}
