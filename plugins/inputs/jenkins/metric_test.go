package jenkins

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestGetMetric(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, s)
	}))

	time.Sleep(time.Second)

	n := Input{
		URL:      ts.URL,
		Key:      "ccc",
		Interval: datakit.Duration{Duration: time.Second * 1},
	}

	n.setup()

	n.getPluginMetric()
}

var s = `{
  "version" : "4.0.0",
  "gauges" : {
    "jenkins.executor.count.value" : {
      "value" : 2
    },
    "jenkins.executor.free.value" : {
      "value" : 2
    },
    "jenkins.executor.in-use.value" : {
      "value" : 0
    },
    "jenkins.health-check.count" : {
      "value" : 4
    },
    "jenkins.health-check.inverse-score" : {
      "value" : 0.25
    },
    "jenkins.health-check.score" : {
      "value" : 0.75
    },
    "jenkins.job.averageDepth" : {
      "value" : 1.0
    },
    "jenkins.job.count.value" : {
      "value" : 3
    },
    "jenkins.node.count.value" : {
      "value" : 1
    },
    "jenkins.node.offline.value" : {
      "value" : 0
    },
    "jenkins.node.online.value" : {
      "value" : 1
    },
    "jenkins.plugins.active" : {
      "value" : 81
    },
    "jenkins.plugins.failed" : {
      "value" : 3
    },
    "jenkins.plugins.inactive" : {
      "value" : 0
    },
    "jenkins.plugins.withUpdate" : {
      "value" : 0
    },
    "jenkins.project.count.value" : {
      "value" : 2
    },
    "jenkins.project.disabled.count.value" : {
      "value" : 0
    },
    "jenkins.project.enabled.count.value" : {
      "value" : 2
    },
    "jenkins.queue.blocked.value" : {
      "value" : 0
    },
    "jenkins.queue.buildable.value" : {
      "value" : 0
    },
    "jenkins.queue.pending.value" : {
      "value" : 0
    },
    "jenkins.queue.size.value" : {
      "value" : 0
    },
    "jenkins.queue.stuck.value" : {
      "value" : 0
    },
    "jenkins.versions.core" : {
      "value" : "2.277.4"
    },
    "system.cpu.load" : {
      "value" : 0.12
    },
    "vm.blocked.count" : {
      "value" : 0
    },
    "vm.class.loaded" : {
      "value" : 15117
    },
    "vm.class.unloaded" : {
      "value" : 9
    },
    "vm.count" : {
      "value" : 30
    },
    "vm.cpu.load" : {
      "value" : 0.0066711140760507
    },
    "vm.daemon.count" : {
      "value" : 15
    },
    "vm.deadlock.count" : {
      "value" : 0
    },
    "vm.deadlocks" : {
      "value" : [ ]
    },
    "vm.file.descriptor.ratio" : {
      "value" : 3.1280517578125E-4
    },
    "vm.gc.G1-Old-Generation.count" : {
      "value" : 0
    },
    "vm.gc.G1-Old-Generation.time" : {
      "value" : 0
    },
    "vm.gc.G1-Young-Generation.count" : {
      "value" : 77
    },
    "vm.gc.G1-Young-Generation.time" : {
      "value" : 923
    },
    "vm.memory.heap.committed" : {
      "value" : 2069889024
    },
    "vm.memory.heap.init" : {
      "value" : 130023424
    },
    "vm.memory.heap.max" : {
      "value" : 2069889024
    },
    "vm.memory.heap.usage" : {
      "value" : 0.4412974132472138
    },
    "vm.memory.heap.used" : {
      "value" : 913436672
    },
    "vm.memory.non-heap.committed" : {
      "value" : 168816640
    },
    "vm.memory.non-heap.init" : {
      "value" : 7667712
    },
    "vm.memory.non-heap.max" : {
      "value" : -1
    },
    "vm.memory.non-heap.usage" : {
      "value" : -1.55603512E8
    },
    "vm.memory.non-heap.used" : {
      "value" : 155603512
    },
    "vm.memory.pools.CodeHeap-'non-nmethods'.committed" : {
      "value" : 2555904
    },
    "vm.memory.pools.CodeHeap-'non-nmethods'.init" : {
      "value" : 2555904
    },
    "vm.memory.pools.CodeHeap-'non-nmethods'.max" : {
      "value" : 5832704
    },
    "vm.memory.pools.CodeHeap-'non-nmethods'.usage" : {
      "value" : 0.4020365168539326
    },
    "vm.memory.pools.CodeHeap-'non-nmethods'.used" : {
      "value" : 2344960
    },
    "vm.memory.pools.CodeHeap-'non-profiled-nmethods'.committed" : {
      "value" : 12845056
    },
    "vm.memory.pools.CodeHeap-'non-profiled-nmethods'.init" : {
      "value" : 2555904
    },
    "vm.memory.pools.CodeHeap-'non-profiled-nmethods'.max" : {
      "value" : 122912768
    },
    "vm.memory.pools.CodeHeap-'non-profiled-nmethods'.usage" : {
      "value" : 0.10308188649693414
    },
    "vm.memory.pools.CodeHeap-'non-profiled-nmethods'.used" : {
      "value" : 12670080
    },
    "vm.memory.pools.CodeHeap-'profiled-nmethods'.committed" : {
      "value" : 39780352
    },
    "vm.memory.pools.CodeHeap-'profiled-nmethods'.init" : {
      "value" : 2555904
    },
    "vm.memory.pools.CodeHeap-'profiled-nmethods'.max" : {
      "value" : 122912768
    },
    "vm.memory.pools.CodeHeap-'profiled-nmethods'.usage" : {
      "value" : 0.3167967708611037
    },
    "vm.memory.pools.CodeHeap-'profiled-nmethods'.used" : {
      "value" : 38938368
    },
    "vm.memory.pools.Compressed-Class-Space.committed" : {
      "value" : 12845056
    },
    "vm.memory.pools.Compressed-Class-Space.init" : {
      "value" : 0
    },
    "vm.memory.pools.Compressed-Class-Space.max" : {
      "value" : 1073741824
    },
    "vm.memory.pools.Compressed-Class-Space.usage" : {
      "value" : 0.009499222040176392
    },
    "vm.memory.pools.Compressed-Class-Space.used" : {
      "value" : 10199712
    },
    "vm.memory.pools.G1-Eden-Space.committed" : {
      "value" : 1261436928
    },
    "vm.memory.pools.G1-Eden-Space.init" : {
      "value" : 15728640
    },
    "vm.memory.pools.G1-Eden-Space.max" : {
      "value" : -1
    },
    "vm.memory.pools.G1-Eden-Space.usage" : {
      "value" : 0.6176226101413134
    },
    "vm.memory.pools.G1-Eden-Space.used" : {
      "value" : 779091968
    },
    "vm.memory.pools.G1-Eden-Space.used-after-gc" : {
      "value" : 0
    },
    "vm.memory.pools.G1-Old-Gen.committed" : {
      "value" : 765460480
    },
    "vm.memory.pools.G1-Old-Gen.init" : {
      "value" : 114294784
    },
    "vm.memory.pools.G1-Old-Gen.max" : {
      "value" : 2069889024
    },
    "vm.memory.pools.G1-Old-Gen.usage" : {
      "value" : 0.04413429267983789
    },
    "vm.memory.pools.G1-Old-Gen.used" : {
      "value" : 91353088
    },
    "vm.memory.pools.G1-Old-Gen.used-after-gc" : {
      "value" : 0
    },
    "vm.memory.pools.G1-Survivor-Space.committed" : {
      "value" : 42991616
    },
    "vm.memory.pools.G1-Survivor-Space.init" : {
      "value" : 0
    },
    "vm.memory.pools.G1-Survivor-Space.max" : {
      "value" : -1
    },
    "vm.memory.pools.G1-Survivor-Space.usage" : {
      "value" : 1.0
    },
    "vm.memory.pools.G1-Survivor-Space.used" : {
      "value" : 42991616
    },
    "vm.memory.pools.G1-Survivor-Space.used-after-gc" : {
      "value" : 42991616
    },
    "vm.memory.pools.Metaspace.committed" : {
      "value" : 100790272
    },
    "vm.memory.pools.Metaspace.init" : {
      "value" : 0
    },
    "vm.memory.pools.Metaspace.max" : {
      "value" : -1
    },
    "vm.memory.pools.Metaspace.usage" : {
      "value" : 0.9073335172664283
    },
    "vm.memory.pools.Metaspace.used" : {
      "value" : 91450392
    },
    "vm.memory.total.committed" : {
      "value" : 2238705664
    },
    "vm.memory.total.init" : {
      "value" : 137691136
    },
    "vm.memory.total.max" : {
      "value" : 2069889023
    },
    "vm.memory.total.used" : {
      "value" : 1069040184
    },
    "vm.new.count" : {
      "value" : 0
    },
    "vm.runnable.count" : {
      "value" : 7
    },
    "vm.terminated.count" : {
      "value" : 0
    },
    "vm.timed_waiting.count" : {
      "value" : 11
    },
    "vm.uptime.milliseconds" : {
      "value" : 13503953
    },
    "vm.waiting.count" : {
      "value" : 12
    }
  }
}`
