// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cat testing
package cat

import (
	"encoding/xml"
	"testing"
	"time"
)

func TestUnmarshal(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?>
<status timestamp="2023-05-31 14:38:23.706">
   <runtime start-time="1685514987499" up-time="116231" java-version="1.8.0_371" user-name="songlq">
      <user-dir>/home/songlq/gitee/cat-demo/cat-demo-stock</user-dir>
      <java-classpath>cat-demo-stock.jar,jaccess.jar,localedata.jar,dnsns.jar,nashorn.jar,zipfs.jar,cldrdata.jar,sunpkcs11.jar,sunec.jar,jfxrt.jar,sunjce_provider.jar</java-classpath>
   </runtime>
   <os name="Linux" arch="amd64" version="5.15.77-amd64-desktop" available-processors="16" system-load-average="0.46" process-time="9330000000" total-physical-memory="32759095296" free-physical-memory="5753192448" committed-virtual-memory="14933397504" total-swap-space="17179865088" free-swap-space="17179865088"/>
   <disk>
      <disk-volume id="/" total="52521566208" free="42877284352" usable="40176152576"/>
      <disk-volume id="/data" total="347594051584" free="283666124800" usable="265934340096"/>
   </disk>
   <memory max="7281311744" total="648019968" free="519999696" heap-usage="128020272" non-heap-usage="63735912">
      <gc name="PS Scavenge" count="6" time="51"/>
      <gc name="PS MarkSweep" count="2" time="39"/>
   </memory>
   <thread count="44" daemon-count="39" peek-count="44" total-started-count="68" cat-thread-count="0" pigeon-thread-count="0" http-thread-count="0">
   </thread>
   <message produced="0" overflowed="0" bytes="0"/>
   <extension id="System">
      <extensionDetail id="LoadAverage" value="0.46"/>
      <extensionDetail id="FreePhysicalMemory" value="5.753192448E9"/>
      <extensionDetail id="FreeSwapSpaceSize" value="1.7179865088E10"/>
   </extension>
   <extension id="Disk">
      <extensionDetail id="/ Free" value="4.2877284352E10"/>
      <extensionDetail id="/data Free" value="2.836661248E11"/>
   </extension>
   <extension id="GC">
      <extensionDetail id="PS ScavengeCount" value="6.0"/>
      <extensionDetail id="PS ScavengeTime" value="51.0"/>
      <extensionDetail id="PS MarkSweepCount" value="2.0"/>
      <extensionDetail id="PS MarkSweepTime" value="39.0"/>
   </extension>
   <extension id="JVMHeap">
      <extensionDetail id="Code Cache" value="1.3887104E7"/>
      <extensionDetail id="Metaspace" value="4.4273488E7"/>
      <extensionDetail id="Compressed Class Space" value="5581400.0"/>
      <extensionDetail id="PS Eden Space" value="1.02754112E8"/>
      <extensionDetail id="PS Survivor Space" value="7768368.0"/>
      <extensionDetail id="PS Old Gen" value="1.7497792E7"/>
   </extension>
   <extension id="FrameworkThread">
      <extensionDetail id="HttpThread" value="13.0"/>
      <extensionDetail id="CatThread" value="0.0"/>
      <extensionDetail id="PigeonThread" value="0.0"/>
      <extensionDetail id="ActiveThread" value="44.0"/>
      <extensionDetail id="StartedThread" value="68.0"/>
   </extension>
   <extension id="CatUsage">
      <extensionDetail id="Produced" value="3.0"/>
      <extensionDetail id="Overflowed" value="0.0"/>
      <extensionDetail id="Bytes" value="26580.0"/>
   </extension>
   <extension id="client-send-queue">
      <description><![CDATA[client-send-queue]]></description>
      <extensionDetail id="msg-queue" value="0.0"/>
      <extensionDetail id="atomic-queue" value="0.0"/>
   </extension>
</status>`

	status := &Status{}
	err := xml.Unmarshal([]byte(data), status)
	if err != nil {
		t.Error(err)
		return
	}

	pts := status.toPoint("domain", "hostname")
	if len(pts) == 0 {
		t.Error(err)
		return
	}
	for _, p := range pts {
		t.Logf("keys=%+v", p.Keys())
		t.Logf("fields=%+v", p.Fields())
		j, err := p.MarshalJSON()
		if err == nil {
			t.Logf("point json=%s", string(j))
			t.Log("--")
		}
	}
}

func TestTimeParse(t *testing.T) {
	timeStr := "2023-06-05 14:47:54.409"
	// time.Parse  会丢失时区
	timeNow, err := time.Parse("2006-01-02 15:04:05.000", timeStr)
	if err != nil {
		t.Error(err)
		return
	}

	timeLocal, err := time.ParseInLocation("2006-01-02 15:04:05.000", timeStr, time.Local)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("time string %s", timeNow.String())
	t.Logf("time string local %s", timeLocal.String())
}
