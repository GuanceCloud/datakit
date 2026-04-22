// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggr

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/aggregate"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/endpoint"
)

const (
	maxAsyncSendWorkers = 8
	maxPooledBodySize   = 8 * dataway.DefaultMaxRawBodySize

	aggregatePayloadContentType = "application/x-protobuf"
	identityContentEncoding     = "identity"
	payloadSizeHeader           = "X-Payload-Size"
)

type pooledMarshalBody struct {
	buf []byte
}

var marshalBodyPool = sync.Pool{
	New: func() interface{} {
		return &pooledMarshalBody{
			buf: make([]byte, 0, dataway.DefaultMaxRawBodySize),
		}
	},
}

type tailSamplingSendTask struct {
	pickKey uint64
	packet  *aggregate.DataPacket
}

type metricBatchSendTask struct {
	pickKey uint64
	batch   *aggregate.Batchs
}

type sendStatusError struct {
	reason string
	msg    string
}

func (e *sendStatusError) Error() string {
	return e.msg
}

func (ag *Aggregator) SendTailSamplingPackages(packages map[uint64]*aggregate.DataPacket) error {
	if len(packages) == 0 {
		log.Debugf("skip sending tail sampling packages: no packages")
		return nil
	}

	maxRawBodySize := ag.maxRawBodySize()
	configVersion := ag.tailSamplingConfigVersion()
	tasks := make([]tailSamplingSendTask, 0, len(packages))
	for pickKey, pkg := range packages {
		if pkg == nil || pkg.PointCount <= 0 || len(pkg.PointsPayload) == 0 {
			continue
		}
		if pkg.ConfigVersion == 0 && configVersion != 0 {
			pkg.ConfigVersion = configVersion
		}

		splitPkgs := splitDataPacketBySize(pkg, maxRawBodySize)
		if len(splitPkgs) > 1 {
			log.Debugf("split tail sampling package: pick_key=%d points=%d split=%d max_raw_body_size=%d",
				pickKey, pkg.PointCount, len(splitPkgs), maxRawBodySize)
		}
		for _, splitPkg := range splitPkgs {
			tasks = append(tasks, tailSamplingSendTask{pickKey: pickKey, packet: splitPkg})
		}
	}

	if len(tasks) == 0 {
		log.Debugf("skip sending tail sampling packages: no valid packages")
		return nil
	}

	recordGeneratedBatches("tail_sampling", tasks[0].packet.DataType, len(tasks))

	if err := ag.runAsyncSend(len(tasks), func(i int) error {
		task := tasks[i]
		return ag.sendTailSamplingPackage(task.pickKey, task.packet)
	}); err != nil {
		return err
	}

	return nil
}

func (ag *Aggregator) sendTailSamplingPackage(pickKey uint64, pkg *aggregate.DataPacket) error {
	startTime := time.Now()
	pointsCount := int(pkg.PointCount)
	category := pkg.DataType
	body, err := marshalDataPacketWithPool(pkg)
	if err != nil {
		log.Errorf("marshal tail sampling package failed: %v", err)
		recordSendFailed("tail_sampling", category, "marshal")
		recordLostPoints("tail_sampling", category, "marshal", pointsCount)
		return err
	}
	defer putMarshalBody(body)

	eps := ag.endpointsForPickKey(pickKey)
	if len(eps) == 0 {
		err := fmt.Errorf("tail sampling endpoint is empty")
		log.Errorf("%v", err)
		recordSendFailed("tail_sampling", category, "transport")
		recordLostPoints("tail_sampling", category, "transport", pointsCount)
		return err
	}

	var firstErr error
	success := false
	attempted := false
	for _, ep := range eps {
		if ep == nil {
			continue
		}
		attempted = true

		resp, _, err := ep.WriteAggrData(&endpoint.AggrData{
			API:             datakit.TailSampling,
			Category:        category,
			ContentType:     aggregatePayloadContentType,
			ContentEncoding: identityContentEncoding,
			Body:            body.buf,
			RawLen:          len(body.buf),
			Points:          pointsCount,
			Headers: map[string]string{
				aggregate.GuancePickKey: strconv.FormatUint(pickKey, 10),
				payloadSizeHeader:       strconv.Itoa(len(body.buf)),
			},
		})
		if resp == nil {
			if err != nil {
				log.Errorf("send tail sampling package failed: %v", err)
			}
			if firstErr == nil {
				if err != nil {
					firstErr = err
				} else {
					firstErr = endpoint.ErrRequestTerminated
				}
			}
			continue
		}

		switch resp.StatusCode {
		case http.StatusOK:
			// log.Debugf("send tail sampling package success: status=%d", resp.StatusCode)
			success = true
		case http.StatusPreconditionFailed:
			ag.sendTSConfigToDW()
			log.Infof("send tail sampling package got status=%d, resend tail sampling config", resp.StatusCode)
			success = true
		default:
			err = &sendStatusError{
				reason: "server",
				msg:    fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
			}
			log.Errorf("send tail sampling package got unexpected status=%d", resp.StatusCode)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	if !attempted && firstErr == nil {
		firstErr = endpoint.ErrRequestTerminated
	}

	recordSendLatency("tail_sampling", category, time.Since(startTime))
	if success {
		recordSendSuccess("tail_sampling", category)
		recordSendPoints("tail_sampling", category, pointsCount)
	}
	if firstErr != nil {
		recordSendFailed("tail_sampling", category, sendFailureReason(firstErr, "server"))
		if !success {
			recordLostPoints("tail_sampling", category, sendFailureReason(firstErr, "server"), pointsCount)
		}
	}

	return firstErr
}

func (ag *Aggregator) SendMetricBatches(category string, batchMap map[uint64]*aggregate.Batchs) error {
	if len(batchMap) == 0 {
		log.Debugf("skip sending metric batches: no batches")
		return nil
	}

	maxRawBodySize := ag.maxRawBodySize()
	tasks := make([]metricBatchSendTask, 0, len(batchMap))

	for pickKey, batch := range batchMap {
		if batch == nil || len(batch.Batchs) == 0 {
			continue
		}

		splitBatches := splitBatchsBySize(batch, maxRawBodySize)
		if len(splitBatches) > 1 {
			log.Debugf("split metric batches: pick_key=%d batches=%d split=%d max_raw_body_size=%d",
				pickKey, len(batch.Batchs), len(splitBatches), maxRawBodySize)
		}
		for _, splitBatch := range splitBatches {
			tasks = append(tasks, metricBatchSendTask{pickKey: pickKey, batch: splitBatch})
		}
	}

	if len(tasks) == 0 {
		log.Debugf("skip sending metric batches: no valid batches")
		return nil
	}

	recordGeneratedBatches("metric", category, len(tasks))

	if err := ag.runAsyncSend(len(tasks), func(i int) error {
		task := tasks[i]
		return ag.sendMetricBatch(category, task.pickKey, task.batch)
	}); err != nil {
		log.Errorf("send metric batches failed: %v", err)
		return err
	}

	return nil
}

func (ag *Aggregator) sendMetricBatch(category string, pickKey uint64, batch *aggregate.Batchs) error {
	startTime := time.Now()
	pointsCount := countSelectedMetricPointsInBatch(batch)

	body, err := marshalBatchsWithPool(batch)
	if err != nil {
		log.Errorf("marshal metric batches failed: %v", err)
		recordSendFailed("metric", category, "marshal")
		recordLostPoints("metric", category, "marshal", pointsCount)
		return err
	}
	defer putMarshalBody(body)

	eps := ag.endpointsForPickKey(pickKey)
	if len(eps) == 0 {
		err := fmt.Errorf("aggregate endpoint is empty")
		log.Errorf("%v", err)
		recordSendFailed("metric", category, "transport")
		recordLostPoints("metric", category, "transport", pointsCount)
		return err
	}

	var firstErr error
	success := false
	attempted := false
	for _, ep := range eps {
		if ep == nil {
			continue
		}
		attempted = true

		log.Debugf("send metric batches: url=%s pick_key=%d batches=%d body_size=%d",
			ep.CategoryURL[datakit.Aggregate], pickKey, len(batch.Batchs), len(body.buf))

		resp, respBody, err := ep.WriteAggrData(&endpoint.AggrData{
			API:             datakit.Aggregate,
			Category:        category,
			ContentType:     aggregatePayloadContentType,
			ContentEncoding: identityContentEncoding,
			Body:            body.buf,
			RawLen:          len(body.buf),
			Points:          pointsCount,
			Headers: map[string]string{
				aggregate.GuancePickKey:    strconv.FormatUint(pickKey, 10),
				aggregate.GuanceRoutingKey: strconv.FormatUint(pickKey, 10),
				payloadSizeHeader:          strconv.Itoa(len(body.buf)),
			},
		})
		if resp == nil {
			if err != nil {
				log.Errorf("send metric batches failed: %v", err)
			}
			if firstErr == nil {
				if err != nil {
					firstErr = err
				} else {
					firstErr = endpoint.ErrRequestTerminated
				}
			}
			continue
		}

		switch resp.StatusCode / 100 {
		case 2:
			log.Debugf("send metric batches success: status=%d", resp.StatusCode)
			success = true
		default:
			err = &sendStatusError{
				reason: "other",
				msg:    fmt.Sprintf("metric batches got unexpected status=%d", resp.StatusCode),
			}
			if firstErr == nil {
				firstErr = err
			}
		}

		log.Debugf("metric batch response body=%s", string(respBody))
	}
	if !attempted && firstErr == nil {
		firstErr = endpoint.ErrRequestTerminated
	}

	recordSendLatency("metric", category, time.Since(startTime))
	if success {
		recordSendSuccess("metric", category)
		recordSendPoints("metric", category, pointsCount)
	}
	if firstErr != nil {
		recordSendFailed("metric", category, sendFailureReason(firstErr, "other"))
		if !success {
			recordLostPoints("metric", category, sendFailureReason(firstErr, "other"), pointsCount)
		}
	}

	return firstErr
}

func (ag *Aggregator) runAsyncSend(taskCount int, sendFn func(taskIndex int) error) error {
	if taskCount == 0 {
		return nil
	}

	workerCount := ag.sendWorkerCount(taskCount)
	taskCh := make(chan int, workerCount)

	var (
		wg       sync.WaitGroup
		firstErr error
		mu       sync.Mutex
	)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for taskIndex := range taskCh {
				if err := sendFn(taskIndex); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
				}
			}
		}()
	}

	for i := 0; i < taskCount; i++ {
		taskCh <- i
	}
	close(taskCh)
	wg.Wait()

	return firstErr
}

func (ag *Aggregator) sendWorkerCount(taskCount int) int {
	workerCount := datakit.AvailableCPUs * 2
	if workerCount <= 0 {
		workerCount = 1
	}
	if workerCount > maxAsyncSendWorkers {
		workerCount = maxAsyncSendWorkers
	}
	if workerCount > taskCount {
		return taskCount
	}

	return workerCount
}

func sendFailureReason(err error, statusReason string) string {
	if err == nil {
		return statusReason
	}

	if errors.Is(err, endpoint.ErrRequestTerminated) {
		return "transport"
	}

	var statusErr *sendStatusError
	if ok := errors.As(err, &statusErr); ok && statusErr != nil && statusErr.reason != "" {
		return statusErr.reason
	}

	return "network"
}

func (ag *Aggregator) maxRawBodySize() int {
	switch {
	case ag.MaxRawBodySize > 0:
		return ag.MaxRawBodySize
	case ag.DW != nil && ag.DW.MaxRawBodySize > 0:
		return ag.DW.MaxRawBodySize
	default:
		return dataway.DefaultMaxRawBodySize
	}
}

func splitDataPacketBySize(pkg *aggregate.DataPacket, maxRawBodySize int) []*aggregate.DataPacket {
	if pkg == nil {
		return nil
	}
	if maxRawBodySize <= 0 || pkg.Size() <= maxRawBodySize || pkg.PointCount <= 1 {
		return []*aggregate.DataPacket{pkg}
	}

	base := cloneDataPacketMeta(pkg)
	base.PointsPayload = nil
	base.PointCount = 0
	base.MaxPointTimeUnixNano = 0
	baseSize := base.Size()
	if baseSize >= maxRawBodySize {
		log.Warnf("tail sampling packet meta exceeds max body size: meta=%d limit=%d group_id=%s",
			baseSize, maxRawBodySize, pkg.RawGroupId)
		return []*aggregate.DataPacket{pkg}
	}

	parts := make([]*aggregate.DataPacket, 0, pkg.PointCount)
	part := cloneDataPacketMeta(pkg)
	part.PointsPayload = make([]byte, 0, len(pkg.PointsPayload))
	part.PointCount = 0
	part.MaxPointTimeUnixNano = 0
	partSize := baseSize
	var splitErr error

	walkErr := point.WalkPBPointsPayload(pkg.PointsPayload, func(raw []byte) bool {
		if len(raw) == 0 {
			return true
		}

		pointSize := protoListElemSize(len(raw))
		if part.PointCount > 0 && partSize+pointSize > maxRawBodySize {
			parts = append(parts, part)

			part = cloneDataPacketMeta(pkg)
			part.PointsPayload = make([]byte, 0, len(pkg.PointsPayload))
			part.PointCount = 0
			part.MaxPointTimeUnixNano = 0
			partSize = baseSize
		}

		pb := &point.PBPoint{}
		if err := pb.Unmarshal(raw); err != nil {
			splitErr = err
			return false
		}

		part.PointsPayload = point.AppendPBPointToPBPointsPayload(part.PointsPayload, pb)
		part.PointCount++
		if pb.Time > part.MaxPointTimeUnixNano {
			part.MaxPointTimeUnixNano = pb.Time
		}
		partSize += pointSize

		if part.PointCount == 1 && partSize > maxRawBodySize {
			parts = append(parts, part)

			log.Warnf("single tail sampling point exceeds max body size: size=%d limit=%d group_id=%s",
				partSize, maxRawBodySize, pkg.RawGroupId)

			part = cloneDataPacketMeta(pkg)
			part.PointsPayload = make([]byte, 0, len(pkg.PointsPayload))
			part.PointCount = 0
			part.MaxPointTimeUnixNano = 0
			partSize = baseSize
		}
		return true
	})
	if walkErr != nil {
		log.Warnf("split tail sampling packet failed to walk payload: group_id=%s err=%v", pkg.RawGroupId, walkErr)
		return []*aggregate.DataPacket{pkg}
	}
	if splitErr != nil {
		log.Warnf("split tail sampling packet failed to decode point payload: group_id=%s err=%v", pkg.RawGroupId, splitErr)
		return []*aggregate.DataPacket{pkg}
	}

	if part.PointCount > 0 {
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return []*aggregate.DataPacket{pkg}
	}

	return parts
}

func splitBatchsBySize(batch *aggregate.Batchs, maxRawBodySize int) []*aggregate.Batchs {
	if batch == nil {
		return nil
	}
	if maxRawBodySize <= 0 || batch.Size() <= maxRawBodySize || len(batch.Batchs) <= 1 {
		return []*aggregate.Batchs{batch}
	}

	base := &aggregate.Batchs{PickKey: batch.PickKey}
	baseSize := base.Size()
	if baseSize >= maxRawBodySize {
		log.Warnf("metric batch meta exceeds max body size: meta=%d limit=%d pick_key=%d",
			baseSize, maxRawBodySize, batch.PickKey)
		return []*aggregate.Batchs{batch}
	}

	parts := make([]*aggregate.Batchs, 0, len(batch.Batchs))
	part := &aggregate.Batchs{
		PickKey: batch.PickKey,
		Batchs:  make([]*aggregate.AggregationBatch, 0, minInt(len(batch.Batchs), 64)),
	}
	partSize := baseSize

	for _, one := range batch.Batchs {
		if one == nil {
			continue
		}

		oneSize := protoListElemSize(one.Size())
		if len(part.Batchs) > 0 && partSize+oneSize > maxRawBodySize {
			parts = append(parts, part)
			part = &aggregate.Batchs{
				PickKey: batch.PickKey,
				Batchs:  make([]*aggregate.AggregationBatch, 0, minInt(len(batch.Batchs), 64)),
			}
			partSize = baseSize
		}

		part.Batchs = append(part.Batchs, one)
		partSize += oneSize

		if len(part.Batchs) == 1 && partSize > maxRawBodySize {
			parts = append(parts, part)
			log.Warnf("single metric batch exceeds max body size: size=%d limit=%d pick_key=%d",
				partSize, maxRawBodySize, batch.PickKey)

			part = &aggregate.Batchs{
				PickKey: batch.PickKey,
				Batchs:  make([]*aggregate.AggregationBatch, 0, minInt(len(batch.Batchs), 64)),
			}
			partSize = baseSize
		}
	}

	if len(part.Batchs) > 0 {
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return []*aggregate.Batchs{batch}
	}

	return parts
}

func cloneDataPacketMeta(pkg *aggregate.DataPacket) *aggregate.DataPacket {
	if pkg == nil {
		return nil
	}

	return &aggregate.DataPacket{
		GroupIdHash:            pkg.GroupIdHash,
		RawGroupId:             pkg.RawGroupId,
		Token:                  pkg.Token,
		Source:                 pkg.Source,
		DataType:               pkg.DataType,
		ConfigVersion:          pkg.ConfigVersion,
		HasError:               pkg.HasError,
		GroupKey:               pkg.GroupKey,
		PointCount:             pkg.PointCount,
		TraceStartTimeUnixNano: pkg.TraceStartTimeUnixNano,
		TraceEndTimeUnixNano:   pkg.TraceEndTimeUnixNano,
		PointsPayload:          pkg.PointsPayload,
		MaxPointTimeUnixNano:   pkg.MaxPointTimeUnixNano,
	}
}

func marshalDataPacketWithPool(pkg *aggregate.DataPacket) (*pooledMarshalBody, error) {
	if pkg == nil {
		return nil, fmt.Errorf("tail sampling package is nil")
	}

	body := getMarshalBody(pkg.Size())
	n, err := pkg.MarshalToSizedBuffer(body.buf)
	if err != nil {
		putMarshalBody(body)
		return nil, err
	}
	body.buf = body.buf[:n]

	return body, nil
}

func marshalBatchsWithPool(batch *aggregate.Batchs) (*pooledMarshalBody, error) {
	if batch == nil {
		return nil, fmt.Errorf("metric batch is nil")
	}

	body := getMarshalBody(batch.Size())
	n, err := batch.MarshalToSizedBuffer(body.buf)
	if err != nil {
		putMarshalBody(body)
		return nil, err
	}
	body.buf = body.buf[:n]

	return body, nil
}

func getMarshalBody(size int) *pooledMarshalBody {
	body, ok := marshalBodyPool.Get().(*pooledMarshalBody)
	if !ok || body == nil {
		body = &pooledMarshalBody{}
	}

	if size <= 0 {
		body.buf = body.buf[:0]
		return body
	}

	if cap(body.buf) < size {
		body.buf = make([]byte, size)
	} else {
		body.buf = body.buf[:size]
	}

	return body
}

func putMarshalBody(body *pooledMarshalBody) {
	if body == nil {
		return
	}

	if cap(body.buf) == 0 {
		return
	}

	if cap(body.buf) > maxPooledBodySize {
		return
	}

	body.buf = body.buf[:0]
	marshalBodyPool.Put(body)
}

func protoListElemSize(payloadSize int) int {
	return 1 + uvarintSize(uint64(payloadSize)) + payloadSize
}

func uvarintSize(n uint64) int {
	size := 1
	for n >= 0x80 {
		n >>= 7
		size++
	}
	return size
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
