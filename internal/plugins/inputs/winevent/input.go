// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package winevent

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
	"syscall"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"golang.org/x/sys/windows"
)

var statusList = []string{"info", "critical", "error", "warning", "info"}

func (*Input) SampleConfig() string { return sample }

func (*Input) Catalog() string { return "windows" }

func (*Input) RunPipeline() { /*nil*/ }

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelWindows}
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func (ipt *Input) Init() {
	// init cache
	defaultDuration := 30 * time.Second
	ipt.handleCache = newHandleCache(defaultDuration, 10, func(s string, eh EvtHandle) {
		if eh != NilHandle {
			_EvtClose(eh) // nolint:errcheck
		}
	}, func() {
		eventCacheNumber.Set(float64(ipt.handleCache.Size()))
	})

	ipt.mergedTags = inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")

	if ipt.subscribeFlag == 0 {
		ipt.subscribeFlag = EvtSubscribeToFutureEvents
	}

	ipt.handleCache.StartCleanWorker(defaultDuration)

	if ipt.EventFetchSize <= 0 {
		ipt.EventFetchSize = defaultEventFetchSize
	}

	ipt.winMetaCache = winMetaCache{
		cache:  make(map[string]winMetaCacheEntry),
		ttl:    time.Hour,
		logger: l,
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	var err error

	ipt.Init()

	ipt.subscription, err = ipt.evtSubscribe("", ipt.Query)
	if err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
			dkio.WithLastErrorCategory(point.Logging))
		return
	}
	defer func() {
		if ipt.handleCache != nil {
			ipt.handleCache.StopCleanWorker()
		}
	}()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("win event exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("win event return")
			return

		default:
			start := time.Now()
			events, err := ipt.fetchEvents(ipt.subscription)
			if err != nil {
				if !errors.Is(err, ErrorNoMoreItems) {
					l.Errorf("fetch events failed: %s", err.Error())

					ipt.feeder.FeedLastError(err.Error(),
						dkio.WithLastErrorInput(inputName),
						dkio.WithLastErrorCategory(point.Logging))
				} // else: ignore no-more-items error
			}

			if len(events) == 0 { // no event available
				time.Sleep(10 * time.Millisecond)
			} else {
				for _, event := range events {
					ipt.handleEvent(event)
				}

				if len(ipt.collectCache) > 0 {
					if err := ipt.feeder.FeedV2(point.Logging, ipt.collectCache,
						dkio.WithCollectCost(time.Since(start)),
						dkio.WithInputName(inputName),
					); err != nil {
						l.Errorf("feed error: %s", err.Error())
						ipt.feeder.FeedLastError(err.Error(),
							dkio.WithLastErrorInput(inputName),
							dkio.WithLastErrorCategory(point.Logging))
					}
					// expose metric
					pointTime := ipt.collectCache[len(ipt.collectCache)-1].Time()
					diffTime := time.Since(pointTime).Seconds()
					if diffTime > 60 {
						eventTimeDiff.Observe(diffTime)
					}

					ipt.collectCache = ipt.collectCache[:0]
				}
			}
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) handleEvent(event Event) {
	var kvs point.KVs
	ts, err := time.Parse(time.RFC3339Nano, event.TimeCreated.SystemTime)
	if err != nil {
		l.Error(err.Error())
		ts = time.Now()
	}

	msg, err := json.Marshal(event)
	if err != nil {
		l.Error(err.Error())
		return
	}

	kvs = kvs.Add("event_source", event.Source.Name, false, true)
	kvs = kvs.Add("event_id", event.EventID.ID, false, true)
	kvs = kvs.Add("version", event.Version, false, true)
	kvs = kvs.Add("task", event.Task, false, true)
	kvs = kvs.Add("keyword", event.Keywords, false, true)
	kvs = kvs.Add("event_record_id", event.EventRecordID, false, true)
	kvs = kvs.Add("process_id", int(event.Execution.ProcessID), false, true)
	kvs = kvs.Add("channel", event.Channel, false, true)
	kvs = kvs.Add("computer", event.Computer, false, true)
	kvs = kvs.Add("message", event.Message, false, true)
	kvs = kvs.Add("level", event.Level, false, true)
	kvs = kvs.Add("total_message", string(msg), false, true)
	kvs = kvs.Add("status", ipt.getEventStatus(int(event.LevelRaw)), false, true)

	opts := point.CommonLoggingOptions()
	opts = append(opts, point.WithTime(ts))

	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	pt := point.NewPointV2("windows_event", kvs, opts...)

	ipt.collectCache = append(ipt.collectCache, pt)
}

func (ipt *Input) getEventStatus(level int) string {
	if level >= 0 && level < len(statusList) {
		return statusList[level]
	}

	return "info"
}

func (ipt *Input) evtSubscribe(logName, xquery string) (EvtHandle, error) {
	var logNamePtr, xqueryPtr *uint16

	sigEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(sigEvent) // nolint:errcheck

	logNamePtr, err = syscall.UTF16PtrFromString(logName)
	if err != nil {
		return 0, err
	}

	xqueryPtr, err = syscall.UTF16PtrFromString(xquery)
	if err != nil {
		return 0, err
	}

	subsHandle, err := _EvtSubscribe(0, uintptr(sigEvent), logNamePtr, xqueryPtr,
		0, 0, 0, ipt.subscribeFlag)
	if err != nil {
		return 0, err
	}

	return subsHandle, nil
}

func (ipt *Input) fetchEventHandles(subsHandle EvtHandle) ([]EvtHandle, error) {
	var eventsNumber uint32
	var evtReturned uint32

	eventsNumber = ipt.EventFetchSize

	eventHandles := make([]EvtHandle, eventsNumber)

	err := _EvtNext(subsHandle, eventsNumber, &eventHandles[0], 0, 0, &evtReturned)
	if err != nil {
		if errors.Is(err, ErrorInvalidOperation) && evtReturned == 0 {
			return nil, ErrorNoMoreItems
		}
		return nil, err
	}

	return eventHandles[:evtReturned], nil
}

func (ipt *Input) fetchEvents(subsHandle EvtHandle) ([]Event, error) {
	var events []Event

	eventHandles, err := ipt.fetchEventHandles(subsHandle)
	if err != nil {
		return nil, err
	}

	for _, eventHandle := range eventHandles {
		if eventHandle != 0 {
			event, err := ipt.renderEvent(eventHandle)
			if err == nil {
				// w.Log.Debugf("Got event: %v", event)
				events = append(events, event)
			}
		}
	}

	for i := 0; i < len(eventHandles); i++ {
		err := _EvtClose(eventHandles[i])
		if err != nil {
			return events, err
		}
	}
	return events, nil
}

// TODO: publisherMeta
func (ipt *Input) setValues(publisherMeta *WinMeta, event *Event) {
	rawKeyword := int64(event.KeywordsRaw)

	if len(event.Keywords) == 0 {
		for m, k := range keywordsMap {
			if rawKeyword&m != 0 {
				event.Keywords = append(event.Keywords, k)
				rawKeyword &^= m
			}
		}

		if publisherMeta != nil {
			for mask, keyword := range publisherMeta.Keywords {
				if rawKeyword&mask != 0 {
					event.Keywords = append(event.Keywords, keyword)
					rawKeyword &^= mask
				}
			}
		}
	}
	event.KeywordsText = strings.Join(event.Keywords, ",")

	var found bool
	if event.Opcode == "" {
		if event.OpcodeRaw != nil {
			event.Opcode, found = opcodesMap[*event.OpcodeRaw]
			if !found && publisherMeta != nil {
				event.Opcode = publisherMeta.Opcodes[*event.OpcodeRaw]
			}
		}
	}

	if event.Level == "" {
		event.Level, found = levelsMap[event.LevelRaw]
		if !found && publisherMeta != nil {
			event.Level = publisherMeta.Levels[event.LevelRaw]
		}
	}

	if event.Task == "" {
		if publisherMeta != nil {
			event.Task, found = publisherMeta.Tasks[event.TaskRaw]
			if !found {
				event.Task = tasksMap[event.TaskRaw]
			}
		} else {
			event.Task = tasksMap[event.TaskRaw]
		}
	}
}

func (ipt *Input) renderEvent(eventHandle EvtHandle) (Event, error) {
	var bufferUsed, propertyCount uint32

	event := Event{}
	err := _EvtRender(0, eventHandle, EvtRenderEventXML, uint32(len(ipt.buf)), &ipt.buf[0], &bufferUsed, &propertyCount)
	if err != nil {
		return event, err
	}

	eventXML, err := DecodeUTF16(ipt.buf[:bufferUsed])
	if err != nil {
		return event, err
	}
	err = xml.Unmarshal(eventXML, &event)
	if err != nil {
		// We can return event without most text values,
		// that way we will not loose information
		// This can happen when processing Forwarded Events
		return event, nil //nolint:nilerr
	}

	ipt.setValues(ipt.winMeta(event.Source.Name), &event)

	var publisherHandle EvtHandle

	if v := ipt.handleCache.Get(event.Source.Name); v == NilHandle {
		handle, err := openPublisherMetadata(0, event.Source.Name, 0)
		if err != nil {
			return event, nil //nolint:nilerr
		}

		publisherHandle = handle
		ipt.handleCache.Put(event.Source.Name, handle)
	} else {
		publisherHandle = v
	}

	if event.KeywordsText == "" {
		keywords, err := formatEventString(EvtFormatMessageKeyword, eventHandle, publisherHandle)
		if err == nil {
			event.KeywordsText = keywords
		}
	}

	if event.Message == "" {
		message, err := formatEventString(EvtFormatMessageEvent, eventHandle, publisherHandle)
		if err == nil {
			if len(message) > 1024*1024 { // max 1 MB
				message = message[0 : 1024*1024]
			}
			event.Message = message
		}
	}

	if event.Level == "" {
		level, err := formatEventString(EvtFormatMessageLevel, eventHandle, publisherHandle)
		if err == nil {
			event.Level = level
		}
	}

	if event.Task == "" {
		task, err := formatEventString(EvtFormatMessageTask, eventHandle, publisherHandle)
		if err == nil {
			event.Task = task
		}
	}

	if event.Opcode == "" {
		opcode, err := formatEventString(EvtFormatMessageOpcode, eventHandle, publisherHandle)
		if err == nil {
			event.Opcode = opcode
		}
	}
	return event, nil
}

func formatEventString(
	messageFlag EvtFormatMessageFlag,
	eventHandle EvtHandle,
	publisherHandle EvtHandle,
) (string, error) {
	var bufferUsed uint32
	err := _EvtFormatMessage(publisherHandle, eventHandle, 0, 0, 0, messageFlag,
		0, nil, &bufferUsed)
	if err != nil && !errors.Is(err, ErrorInsufficientBuffer) {
		return "", err
	}

	if bufferUsed < 1 {
		return "", nil
	}

	bb := NewPooledByteBuffer()
	defer bb.Free()
	bb.Reserve(int(bufferUsed * 2))

	err = _EvtFormatMessage(publisherHandle, eventHandle, 0, 0, 0, messageFlag,
		bufferUsed, bb.PtrAt(0), &bufferUsed)
	bufferUsed *= 2
	if err != nil {
		return "", err
	}

	result, err := DecodeUTF16(bb.Bytes())
	if err != nil {
		return "", err
	}

	var out string
	if messageFlag == EvtFormatMessageKeyword {
		// Keywords are returned as array of a zero-terminated strings
		splitZero := func(c rune) bool { return c == '\x00' }
		eventKeywords := strings.FieldsFunc(string(result), splitZero)
		// So convert them to comma-separated string
		out = strings.Join(eventKeywords, ",")
	} else {
		result := bytes.Trim(result, "\x00")
		out = string(result)
	}
	return out, nil
}

// openPublisherMetadata opens a handle to the publisher's metadata. Close must
// be called on returned EvtHandle when finished with the handle.
func openPublisherMetadata(
	session EvtHandle,
	publisherName string,
	lang uint32,
) (EvtHandle, error) {
	p, err := syscall.UTF16PtrFromString(publisherName)
	if err != nil {
		return 0, err
	}

	h, err := _EvtOpenPublisherMetadata(session, p, nil, lang, 0)
	if err != nil {
		return 0, err
	}

	return h, nil
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			buf:   make([]byte, 1<<14),
			Query: query,

			semStop: cliutils.NewSem(),
			feeder:  dkio.DefaultFeeder(),
			Tagger:  datakit.DefaultGlobalTagger(),
		}
	})
}
