// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Beats (https://github.com/elastic/beats).

//go:build windows
// +build windows

package winevent

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/andrewkroh/sys/windows/svc/eventlog"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

const (
	providerName    = "WinlogbeatTestGo"
	sourceName      = "Integration Test"
	gigabyte        = 1 << 30
	maxEventsNumber = 100000

	eventCreateMsgFile = "%SystemRoot%\\System32\\EventCreate.exe"
)

var (
	testQuery = `<QueryList>
    <Query Id="0" Path="WinlogbeatTestGo">
        <Select Path="WinlogbeatTestGo">*</Select>
    </Query>
</QueryList>`
	benchTest = flag.Bool("benchtest", false, "Run benchmarks for the eventlog package.")
)

// mockFeeder implements Feeder interface
type mockFeeder struct {
	ptsNumber int
	semStop   *cliutils.Sem
	maxNumber int
}

func (m *mockFeeder) Feed(name string, category point.Category, pts []*point.Point, opt ...*io.Option) error {
	m.ptsNumber += len(pts)
	if m.maxNumber <= m.ptsNumber {
		m.semStop.Close()
	}

	return nil
}

func (m *mockFeeder) FeedV2(category point.Category, pts []*point.Point, opts ...io.FeedOption) error {
	m.ptsNumber += len(pts)
	if m.maxNumber <= m.ptsNumber {
		m.semStop.Close()
	}

	return nil
}

func (m *mockFeeder) FeedLastError(err string, opts ...io.LastErrorOption) {}

func TestEventlog(t *testing.T) {
	maxEventsNumber := 10000
	writer, teardown := createLog(t)
	defer teardown()

	setLogSize(t, providerName, gigabyte)

	// Publish large test messages.
	const messageSize = 256 // Originally 31800, such a large value resulted in an empty eventlog under Win10.
	for i := 0; i < maxEventsNumber; i++ {
		safeWriteEvent(t, writer, eventlog.Info, uint32(i%1000)+1, []string{strconv.Itoa(i) + " " + randomSentence(messageSize)})
	}

	semStop := cliutils.NewSem()
	feeder := &mockFeeder{
		semStop:   semStop,
		maxNumber: maxEventsNumber,
	}

	input := &Input{
		buf:            make([]byte, 1<<14),
		Query:          testQuery,
		semStop:        semStop,
		feeder:         feeder,
		EventFetchSize: 100,
		subscribeFlag:  EvtSubscribeStartAtOldestRecord,
		Tagger:         datakit.DefaultGlobalTagger(),
	}

	start := time.Now()
	input.Run()

	assert.Equal(t, maxEventsNumber, feeder.ptsNumber)
	t.Logf("Avg: %f", float64(maxEventsNumber)/time.Since(start).Seconds())
}

func TestBenchmarkEventlog(t *testing.T) {
	if !*benchTest {
		t.Skip("-benchtest not enabled")
	}
	writer, teardown := createLog(t)
	defer teardown()

	setLogSize(t, providerName, gigabyte)

	// Publish large test messages.
	const messageSize = 256 // Originally 31800, such a large value resulted in an empty eventlog under Win10.
	for i := 0; i < maxEventsNumber; i++ {
		safeWriteEvent(t, writer, eventlog.Info, uint32(i%1000)+1, []string{strconv.Itoa(i) + " " + randomSentence(messageSize)})
	}

	for _, fetchSize := range []int{10, 100, 500, 1000} {
		t.Run(fmt.Sprintf("fetch_size=%d", fetchSize), func(t *testing.T) {
			result := testing.Benchmark(benchmarkEventlog(fetchSize))
			outputBenchmarkResults(t, result)
		})
	}
}

func benchmarkEventlog(fetchSize int) func(b *testing.B) {
	return func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			semStop := cliutils.NewSem()
			feeder := &mockFeeder{
				semStop:   semStop,
				maxNumber: maxEventsNumber,
			}
			input := &Input{
				buf:            make([]byte, 1<<14),
				Query:          testQuery,
				semStop:        semStop,
				feeder:         feeder,
				EventFetchSize: uint32(fetchSize),
				subscribeFlag:  EvtSubscribeStartAtOldestRecord,
				Tagger:         datakit.DefaultGlobalTagger(),
			}
			input.Run()
		}
		b.StopTimer()

		b.ReportMetric(float64(maxEventsNumber), "events")
	}
}

func outputBenchmarkResults(t testing.TB, result testing.BenchmarkResult) {
	totalBatches := result.N
	totalEvents := int(result.Extra["events"])
	totalBytes := result.MemBytes
	totalAllocs := result.MemAllocs

	eventsPerSec := float64(totalEvents) / result.T.Seconds()
	bytesPerEvent := float64(totalBytes) / float64(totalEvents)
	bytesPerBatch := float64(totalBytes) / float64(totalBatches)
	allocsPerEvent := float64(totalAllocs) / float64(totalEvents)
	allocsPerBatch := float64(totalAllocs) / float64(totalBatches)

	t.Logf("%.2f events/sec\t %d B/event\t %d B/batch\t %d allocs/event\t %d allocs/batch",
		eventsPerSec, int(bytesPerEvent), int(bytesPerBatch), int(allocsPerEvent), int(allocsPerBatch))
}

// ---- Utility Functions -----
// refer to https://github.com/elastic/beats/blob/main/winlogbeat/eventlog/wineventlog_test.go#L333

// createLog creates a new event log and returns a handle for writing events
// to the log.
func createLog(t testing.TB, messageFiles ...string) (log *eventlog.Log, tearDown func()) {
	const name = providerName
	const source = sourceName

	messageFile := eventCreateMsgFile
	if len(messageFiles) > 0 {
		messageFile = strings.Join(messageFiles, ";")
	}

	existed, err := eventlog.Install(name, source, messageFile, true, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		t.Fatal(err)
	}

	if existed {
		EvtClearLog(NilHandle, name, "") //nolint:errcheck // This is just a resource release.
	}

	log, err = eventlog.Open(source)
	//nolint:errcheck // This is just a resource release.
	if err != nil {
		eventlog.RemoveSource(name, source)
		eventlog.RemoveProvider(name)
		t.Fatal(err)
	}

	//nolint:errcheck // This is just a resource release.
	tearDown = func() {
		log.Close()
		EvtClearLog(NilHandle, name, "")
		eventlog.RemoveSource(name, source)
		eventlog.RemoveProvider(name)
	}

	return log, tearDown
}

func safeWriteEvent(t testing.TB, log *eventlog.Log, etype uint16, eid uint32, msgs []string) {
	deadline := time.Now().Add(time.Second * 10)
	for {
		err := log.Report(etype, eid, msgs)
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("Failed to write event to event log", err)
			return
		}
	}
}

// setLogSize set the maximum number of bytes that an event log can hold.
func setLogSize(t testing.TB, provider string, sizeBytes int) {
	output, err := exec.Command("wevtutil.exe", "sl", "/ms:"+strconv.Itoa(sizeBytes), provider).CombinedOutput() //nolint:gosec // No possibility of command injection.
	if err != nil {
		t.Fatal("Failed to set log size", err, string(output))
	}
}

var randomWords = []string{
	"recover",
	"article",
	"highway",
	"bargain",
	"trolley",
	"college",
	"attract",
	"wriggle",
	"feather",
	"neutral",
	"percent",
	"quality",
	"manager",
	"hunting",
	"arrange",
}

func randomSentence(n uint) string {
	buf := bytes.NewBuffer(make([]byte, n))
	buf.Reset()

	for {
		idx := rand.Uint32() % uint32(len(randomWords))
		word := randomWords[idx]

		if buf.Len()+len(word) <= buf.Cap() {
			buf.WriteString(randomWords[idx])
		} else {
			break
		}

		if buf.Len()+1 <= buf.Cap() {
			buf.WriteByte(' ')
		} else {
			break
		}
	}

	return buf.String()
}

func TestDemo(t *testing.T) {
	t1 := time.Now()
	time.Sleep(10 * time.Second)
	t2 := time.Now()

	delta := t1.Sub(t2)

	t.Log(delta.Seconds())
}
