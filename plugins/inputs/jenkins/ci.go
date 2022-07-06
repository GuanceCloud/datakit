// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jenkins

import (
	"encoding/json"
	"net/http"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type ciEventType byte

const (
	pipeline ciEventType = 1 << (iota + 1)
	stage
	job
	unknown
)

// CI pipeline/job statuses.
const (
	success = "success"
	failed  = "failed"
)

//nolint:lll
type ddSpan struct {
	Service  string             `protobuf:"bytes,1,opt,name=service,proto3" json:"service" msg:"service"`
	Name     string             `protobuf:"bytes,2,opt,name=name,proto3" json:"name" msg:"name"`
	Resource string             `protobuf:"bytes,3,opt,name=resource,proto3" json:"resource" msg:"resource"`
	TraceID  uint64             `protobuf:"varint,4,opt,name=traceID,proto3" json:"trace_id" msg:"trace_id"`
	SpanID   uint64             `protobuf:"varint,5,opt,name=spanID,proto3" json:"span_id" msg:"span_id"`
	ParentID uint64             `protobuf:"varint,6,opt,name=parentID,proto3" json:"parent_id" msg:"parent_id"`
	Start    int64              `protobuf:"varint,7,opt,name=start,proto3" json:"start" msg:"start"`
	Duration int64              `protobuf:"varint,8,opt,name=duration,proto3" json:"duration" msg:"duration"`
	Error    int32              `protobuf:"varint,9,opt,name=error,proto3" json:"error" msg:"error"`
	Meta     map[string]string  `protobuf:"bytes,10,rep,name=meta,proto3" json:"meta" msg:"meta" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Metrics  map[string]float64 `protobuf:"bytes,11,rep,name=metrics,proto3" json:"metrics" msg:"metrics" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"fixed64,2,opt,name=value,proto3"`
	Type     string             `protobuf:"bytes,12,opt,name=type,proto3" json:"type" msg:"type"`
}

type ddTrace []*ddSpan

type ddTraces []ddTrace

func (n *Input) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	l.Debugf("Jenkins CI event server receives request: %s", req.URL.Path)
	traces, err := decodeTraces(req)
	if err != nil {
		l.Errorf(err.Error())
		return
	}
	var pts []*io.Point
	for _, trace := range traces {
		for _, span := range trace {
			if !needStatus(span) {
				l.Debugf("skip span with ci.status = %s", span.Meta["ci.status"])
				continue
			}
			pt, err := n.getPoint(span)
			if err != nil {
				l.Errorf(err.Error())
				continue
			}
			if pt == nil {
				continue
			}
			pts = append(pts, pt)
		}
	}
	if len(pts) == 0 {
		l.Debugf("empty Jenkins CI point array")
		return
	}
	if err := io.Feed("jenkins_ci", datakit.Logging, pts, &io.Option{}); err != nil {
		io.FeedLastError("jenkins_ci", err.Error())
		resp.WriteHeader(http.StatusInternalServerError)
	}
}

func decodeTraces(req *http.Request) (ddTraces, error) {
	var traces ddTraces
	if err := json.NewDecoder(req.Body).Decode(&traces); err != nil {
		return nil, err
	}
	return traces, nil
}

func (n *Input) getPoint(span *ddSpan) (*io.Point, error) {
	switch typeOf(span) {
	case pipeline:
		return n.getPipelinePoint(span)
	case job:
		return n.getJobPoint(span)
	case stage, unknown:
		// We don't need this type of span currently.
		l.Debugf("received unneeded CI event type: %s, skipped", span.Meta["_dd.ci.level"])
		return nil, nil
	default:
		l.Debugf("received unrecognized CI event type, skipped")
		return nil, nil
	}
}

func typeOf(span *ddSpan) ciEventType {
	switch span.Meta["_dd.ci.level"] {
	case "pipeline":
		return pipeline
	case "job":
		return job
	case "stage":
		return stage
	default:
		return unknown
	}
}

func (n *Input) getPipelinePoint(span *ddSpan) (*io.Point, error) {
	measurementName := "jenkins_pipeline"
	tags := getPipelineTags(span)
	n.putExtraTags(tags)
	return io.NewPoint(measurementName, tags, getPipelineFields(span), inputs.OptElectionLogging)
}

func (n *Input) getJobPoint(span *ddSpan) (*io.Point, error) {
	measurementName := "jenkins_job"
	tags := getJobTags(span)
	n.putExtraTags(tags)
	return io.NewPoint(measurementName, tags, getJobFields(span), inputs.OptElectionLogging)
}

func getPipelineTags(span *ddSpan) map[string]string {
	tags := make(map[string]string)
	putTagIfExist(tags, span, "git.commit.author.email", "author_email")
	switch span.Meta["ci.status"] {
	case "success":
		tags["ci_status"] = success
	case "error":
		tags["ci_status"] = failed
	}
	putTagIfExist(tags, span, "git.commit_sha", "commit_sha")
	putTagIfExist(tags, span, "_dd.ci.level", "object_kind")
	putTagIfExist(tags, span, "_dd.ci.level", "operation_name")
	putTagIfExist(tags, span, "ci.pipeline.name", "pipeline_name")
	putTagIfExist(tags, span, "ci.pipeline.url", "pipeline_url")
	putTagIfExist(tags, span, "git.branch", "ref")
	putTagIfExist(tags, span, "git.repository_url", "repository_url")
	if p := extractProjectName(tags["repository_url"]); p != "" {
		tags["resource"] = p
	}
	return tags
}

func getPipelineFields(span *ddSpan) map[string]interface{} {
	fields := make(map[string]interface{})
	putFieldIfExist(fields, span, "git.commit.message", "commit_message")
	putFieldIfExist(fields, span, "ci.pipeline.id", "message")
	putFieldIfExist(fields, span, "ci.pipeline.id", "pipeline_id")
	fields["created_at"] = nano2Micro(span.Start)
	fields["duration"] = nano2Micro(span.Duration)
	fields["finished_at"] = nano2Micro(span.Start + span.Duration)
	return fields
}

func getJobTags(span *ddSpan) map[string]string {
	tags := make(map[string]string)
	putTagIfExist(tags, span, "git.commit_sha", "build_commit_sha")
	putTagIfExist(tags, span, "error.message", "build_failure_reason")
	putTagIfExist(tags, span, "ci.job.name", "build_name")
	putTagIfExist(tags, span, "git.repository_url", "build_repo_name")
	putTagIfExist(tags, span, "ci.stage.name", "build_stage")
	switch span.Meta["ci.status"] {
	case "success":
		tags["build_status"] = success
	case "error":
		tags["build_status"] = failed
	}
	putTagIfExist(tags, span, "_dd.ci.level", "object_kind")
	putTagIfExist(tags, span, "git.commit_sha", "sha")
	putTagIfExist(tags, span, "git.commit.author.email", "user_email")
	if p := extractProjectName(tags["build_repo_name"]); p != "" {
		tags["project_name"] = p
	}
	return tags
}

func getJobFields(span *ddSpan) map[string]interface{} {
	fields := make(map[string]interface{})
	putFieldIfExist(fields, span, "ci.job.number", "build_id")
	putFieldIfExist(fields, span, "ci.job.name", "message")
	putFieldIfExist(fields, span, "ci.pipeline.id", "pipeline_id")
	putFieldIfExist(fields, span, "ci.node.name", "runner_id")
	fields["build_started_at"] = nano2Micro(span.Start)
	fields["build_duration"] = nano2Micro(span.Duration)
	fields["build_finished_at"] = nano2Micro(span.Start + span.Duration)
	return fields
}

func putTagIfExist(tags map[string]string, span *ddSpan, want, tagKey string) {
	if v, has := span.Meta[want]; has {
		tags[tagKey] = v
	}
}

func putFieldIfExist(fields map[string]interface{}, span *ddSpan, want, tagKey string) {
	if v, has := span.Meta[want]; has {
		fields[tagKey] = v
	}
}

// nano2Micro converts nanosecond to microsecond.
func nano2Micro(ns int64) int64 {
	return ns / 1000
}

func extractProjectName(projectURL string) string {
	if projectURL == "" {
		return ""
	}
	if !strings.Contains(projectURL, "/") || !strings.HasSuffix(projectURL, ".git") {
		return ""
	}
	return projectURL[strings.LastIndex(projectURL, "/")+1 : len(projectURL)-4]
}

// putExtraTags puts extra tags specified in CIExtraTags into tags.
// If a tag key already exists in tags, it will not be overwritten.
func (n *Input) putExtraTags(tags map[string]string) {
	for k, v := range n.CIExtraTags {
		if _, has := tags[k]; has {
			continue
		}
		tags[k] = v
	}
}

// needStatus filters spans that are needed.
// i.e. spans with ci.status = success/error.
func needStatus(span *ddSpan) bool {
	status, has := span.Meta["ci.status"]
	if !has {
		return false
	}
	return status == "success" || status == "error"
}
