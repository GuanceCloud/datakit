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

type Telemetry struct {
	lock        sync.RWMutex
	host        Host
	application Application
	traceTime   time.Time
	runtimeID   string
	tags        map[string]string
	fields      map[string]interface{}
	change      bool
}

func (ob *Telemetry) toPoint() *point.Point {
	ob.lock.RLock()
	defer ob.lock.RUnlock()
	opts := point.DefaultObjectOptions()
	opts = append(opts, point.WithTime(ob.traceTime))
	kvs := append(point.NewTags(ob.tags), point.NewKVs(ob.fields)...)
	return point.NewPointV2(telemetryMeasurementName, kvs, opts...)
}

func (ob *Telemetry) setField(key string, val interface{}) {
	ob.lock.Lock()
	defer ob.lock.Unlock()
	ob.fields[key] = val
}

func (ob *Telemetry) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: telemetryMeasurementName,
		Desc: "Collect service, host, process APM telemetry message.",
		Cat:  point.CustomObject,
		Fields: map[string]interface{}{
			string(RequestTypeAppStarted): &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "App Started config",
			},
			string(RequestTypeDependenciesLoaded): &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "App dependencies loaded",
			},
			string(RequestTypeAppClientConfigurationChange): &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "App client configuration change config",
			},
			string(RequestTypeAppProductChange): &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "App product change",
			},
			string(RequestTypeAppIntegrationsChange): &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "App Integrations change",
			},
			string(RequestTypeAppClosing): &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
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
			"name":             inputs.NewTagInfo("same as service name"),
			"env":              inputs.NewTagInfo("Service ENV"),
			"service_version":  inputs.NewTagInfo("Service version"),
			"tracer_version":   inputs.NewTagInfo("DDTrace version"),
			"language_name":    inputs.NewTagInfo("Language name"),
			"language_version": inputs.NewTagInfo("Language version"),
			"runtime_name":     inputs.NewTagInfo("Runtime name"),
			"runtime_version":  inputs.NewTagInfo("Runtime_version"),
			"runtime_patches":  inputs.NewTagInfo("Runtime patches"),
			"runtime_id":       inputs.NewTagInfo("RuntimeID"),
		},
	}
}

func (ob *Telemetry) parseEvent(requestType RequestType, payload interface{}) {
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
		if len(start.Configuration) > 0 {
			for _, conf := range start.Configuration {
				if conf.Name == "trace_tags" {
					str, ok := (conf.Value).(string)
					if ok {
						kvsStr := strings.Split(str, ",")
						for _, st := range kvsStr {
							kvs := strings.Split(st, ":")
							if len(kvs) == 2 {
								ob.tags[kvs[0]] = kvs[1]
								setCustomTags([]string{kvs[0]})
							}
						}
					}
				}
			}
		}
		ob.setField(string(requestType), string(bts))
		log.Debugf("type=%s body=%s", requestType, string(bts))
		ob.change = true
	case RequestTypeDependenciesLoaded,
		RequestTypeAppClientConfigurationChange,
		RequestTypeAppProductChange,
		RequestTypeAppIntegrationsChange:
		bts, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("err=%v", err)
			return
		}
		ob.setField(string(requestType), string(bts))
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
		ob.setField(string(requestType), "service closing")
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
		log.Warnf("unknown telemetry request type")
	}
}

type Manager struct {
	obsLock sync.Mutex
	OBS     map[string]*Telemetry
	OBChan  chan *Telemetry
}

func (ipt *Input) OMInitAndRunning() {
	ipt.om = &Manager{
		OBS:    map[string]*Telemetry{},
		OBChan: make(chan *Telemetry, 10),
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_ddtrace"})
	g.Go(func(ctx context.Context) error {
		for {
			select {
			case ob := <-ipt.om.OBChan:
				ipt.om.obsLock.Lock()
				err := ipt.feeder.FeedV2(
					point.CustomObject,
					[]*point.Point{ob.toPoint()},
					dkio.WithInputName(customObjectFeedName),
				)
				if err != nil {
					log.Errorf("feed err=%v", err)
				}
				ipt.om.obsLock.Unlock()
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
	bodyTags := make(map[string]string)
	bodyTags["hostname"] = body.Host.Hostname
	bodyTags["os"] = body.Host.OS
	bodyTags["os_version"] = body.Host.OSVersion
	bodyTags["architecture"] = body.Host.Architecture
	bodyTags["kernel_name"] = body.Host.KernelName
	bodyTags["kernel_release"] = body.Host.KernelRelease
	bodyTags["kernel_version"] = body.Host.KernelVersion

	bodyTags["service"] = body.Application.ServiceName
	bodyTags["name"] = body.Application.ServiceName + "-" + body.RuntimeID
	bodyTags["env"] = body.Application.Env
	bodyTags["service_version"] = body.Application.ServiceVersion
	bodyTags["tracer_version"] = body.Application.TracerVersion
	bodyTags["language_name"] = body.Application.LanguageName
	bodyTags["language_version"] = body.Application.LanguageVersion
	bodyTags["runtime_name"] = body.Application.RuntimeName
	bodyTags["runtime_version"] = body.Application.RuntimeVersion
	bodyTags["runtime_patches"] = body.Application.RuntimePatches

	bodyTags["runtime_id"] = body.RuntimeID

	for k, v := range bodyTags {
		if v == "" {
			delete(bodyTags, k)
		}
	}

	ob, ok := om.OBS[body.Application.ServiceName+body.RuntimeID]
	if !ok {
		ob = &Telemetry{
			host:        body.Host,
			application: body.Application,
			traceTime:   time.Unix(body.TracerTime, 0),
			runtimeID:   body.RuntimeID,
			tags:        bodyTags,
			fields:      map[string]interface{}{},
		}
	} else {
		ob.host = body.Host
		ob.application = body.Application
		ob.traceTime = time.Unix(body.TracerTime, 0)
	}

	ob.parseEvent(body.RequestType, body.Payload)
	om.OBS[body.Application.ServiceName+body.RuntimeID] = ob
	if ob.change {
		ob.change = false
		// 发生变化，准备发送到io.
		om.OBChan <- ob
	}
}
