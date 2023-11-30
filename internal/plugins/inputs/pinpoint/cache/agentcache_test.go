// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cache pinpoint cache testing.
package cache

import (
	"testing"
	"time"

	ppv1 "github.com/GuanceCloud/tracing-protos/pinpoint-gen-go/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func MockProcessSpan(span *ppv1.PSpan, md metadata.MD) {}

func TestAgentCache_GetSpan(t *testing.T) {
	agentCache := &AgentCache{
		SpanData:    map[int64]*SpanItem{},
		ProcessFunc: MockProcessSpan,
		timeout:     time.Second * 5,
	}
	spanItem := &SpanItem{
		Span:      &ppv1.PSpan{SpanId: 123},
		Meta:      metadata.MD{"agentid": {"tmall"}},
		expiredAt: time.Now().Add(time.Second * 5),
	}
	go agentCache.deleteExpiredSpan()
	agentCache.SetSpan(123, spanItem.Span, spanItem.Meta, time.Second*5)
	agentCache.SetSpan(456, spanItem.Span, spanItem.Meta, time.Second*5)

	item, ok := agentCache.GetSpan(123)
	assert.NotNil(t, item, "item mast not nil")
	if !ok {
		t.Errorf("this ok must be true in 5 second")
	}
	time.Sleep(time.Second * 5)

	exItem, ok := agentCache.GetSpan(456)
	assert.Nil(t, exItem, "exItem must be nill")
	if ok {
		t.Errorf("this ok mast be false out 5 second")
	}
}

func TestAgentCache_GetEvent(t *testing.T) {
	agentCache := &AgentCache{
		EventData: map[string]*EventItem{},
		timeout:   time.Second * 5,
	}

	value := []*ppv1.PSpanEvent{
		{ApiId: 456},
		{ApiId: 123},
	}

	go agentCache.deleteExpiredEvent()
	agentCache.SetEvent("testTraceID_case1", value, time.Second*5)
	agentCache.SetEvent("testTraceID_case2", value, time.Second*5)

	item, ok := agentCache.GetEvent("testTraceID_case1", 1)
	assert.NotNil(t, item, "event item must not nil")
	if !ok {
		t.Errorf("this ok must be true in 5 second")
	}
	time.Sleep(time.Second * 5)

	exItem, ok := agentCache.GetEvent("testTraceID_case2", 1)
	assert.Nil(t, exItem, "exItem must be nill")
	if ok {
		t.Errorf("this ok must be false out 5 second")
	}
}

func TestAgentCache_StoreMeta(t *testing.T) {
	agentCache := &AgentCache{
		Metas:   map[string]*MetaData{},
		timeout: time.Second * 5,
	}
	apiMeta := &ppv1.PApiMetaData{ApiId: 12, ApiInfo: "org.apache.demo"}
	sqlMeta := &ppv1.PSqlMetaData{SqlId: 13, Sql: "select * from data"}
	stringMeta := &ppv1.PStringMetaData{StringId: 14, StringValue: "string_func"}

	agentCache.StoreMeta("testAgentID", apiMeta)
	agentCache.StoreMeta("testAgentID", sqlMeta)
	agentCache.StoreMeta("testAgentID", stringMeta)

	if !agentCache.reWriteFile {
		t.Errorf("reWriteFile must be not true")
		return
	}
	meta := agentCache.Metas["testAgentID"]
	if len(meta.PSqlDatas) != 1 {
		t.Errorf("the length of PSqlDatas must be 1")
		return
	}
	if len(meta.PApiDatas) != 1 {
		t.Errorf("the length of PApiDatas must be 1")
		return
	}
	if len(meta.PStringDatas) != 1 {
		t.Errorf("the length of PStringDatas must be 1")
		return
	}

	res, opt, find := agentCache.FindAPIInfo("testAgentID", 12)
	if !find {
		t.Errorf("can not find api id=12 from PApiDatas")
		return
	}
	if res == "" || opt == "" {
		t.Errorf("res or opt is nil,res=%s opt=%s", res, opt)
	}
}
