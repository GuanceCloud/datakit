// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggr

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/aggregate"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/endpoint"
)

const (
	maxAsyncSendWorkers           = 8
	maxPooledBodySize             = 8 * dataway.DefaultMaxRawBodySize
	pbPointsArrayTag              = byte(1<<3 | 2)
	tailSamplingLegacyProtocolTTL = 10 * time.Minute

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
	pickKey     uint64
	batch       *aggregate.Batchs
	sinkHeaders sinkHeaders
}

type sendStatusError struct {
	reason string
	msg    string
}

func (e *sendStatusError) Error() string {
	return e.msg
}

type sinkHeaders struct {
	key   string
	value string
}

func (h sinkHeaders) apply(headers map[string]string) {
	if h.key == "" || h.value == "" {
		return
	}
	headers[h.key] = h.value
}

func (ag *Aggregator) SendTailSamplingPackages(packages map[uint64]*aggregate.DataPacket) error {
	return ag.SendTailSamplingPackagesContext(context.Background(), packages)
}

// SendTailSamplingPackagesContext sends selected tail-sampling packets until completion or cancellation.
func (ag *Aggregator) SendTailSamplingPackagesContext(ctx context.Context,
	packages map[uint64]*aggregate.DataPacket,
) error {
	if len(packages) == 0 {
		log.Debugf("skip sending tail sampling packages: no packages")
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	sendCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	maxRawBodySize := ag.maxRawBodySize()
	configVersion := ag.tailSamplingConfigVersion()
	workerCount := ag.sendWorkerCount(maxAsyncSendWorkers)
	var totalPoints int64
	for _, pkg := range packages {
		if pkg != nil && pkg.PointCount > 0 && len(pkg.PointsPayload) > 0 {
			totalPoints += int64(pkg.PointCount)
		}
	}
	// Keep the queue proportional to the worker count so a slow or failed
	// downstream cannot retain every generated split in memory.
	taskCh := make(chan tailSamplingSendTask, workerCount)

	var (
		wg           sync.WaitGroup
		firstErrOnce sync.Once
		firstErr     error
		generated    int
		dataType     string
		attempted    atomic.Int64
	)

	recordError := func(err error) {
		firstErrOnce.Do(func() {
			firstErr = err
		})
	}
	stopProduction := func(err error) {
		if ctx.Err() != nil {
			err = context.Cause(ctx)
		}
		recordError(err)
		cancel(err)
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-sendCtx.Done():
					return
				case task, ok := <-taskCh:
					if !ok {
						return
					}
					err := ag.sendTailSamplingPackageContext(sendCtx, task.pickKey, task.packet)
					attempted.Add(int64(task.packet.PointCount))
					if err != nil {
						recordError(err)
					}
				}
			}
		}()
	}

	for pickKey, pkg := range packages {
		if pkg == nil || pkg.PointCount <= 0 || len(pkg.PointsPayload) == 0 {
			continue
		}
		if dataType == "" {
			dataType = pkg.DataType
		}
		if pkg.ConfigVersion == 0 && configVersion != 0 {
			pkg.ConfigVersion = configVersion
		}

		splitCount, err := walkDataPacketPartsBySize(sendCtx, pkg, maxRawBodySize, func(splitPkg *aggregate.DataPacket) error {
			select {
			case <-sendCtx.Done():
				return context.Cause(sendCtx)
			case taskCh <- tailSamplingSendTask{pickKey: pickKey, packet: splitPkg}:
				generated++
				return nil
			}
		})
		if splitCount > 1 {
			log.Debugf("split tail sampling package: pick_key=%d points=%d split=%d max_raw_body_size=%d",
				pickKey, pkg.PointCount, splitCount, maxRawBodySize)
		}
		if err != nil {
			stopProduction(err)
			break
		}
	}
	close(taskCh)
	wg.Wait()
	if ctx.Err() != nil {
		firstErr = context.Cause(ctx)
	} else if firstErr == nil && sendCtx.Err() != nil {
		firstErr = context.Cause(sendCtx)
	}

	if generated == 0 && firstErr == nil {
		log.Debugf("skip sending tail sampling packages: no valid packages")
		return nil
	}

	if generated > 0 {
		recordGeneratedBatches("tail_sampling", dataType, generated)
	}
	if firstErr != nil {
		if unsentPoints := totalPoints - attempted.Load(); unsentPoints > 0 {
			recordLostPoints("tail_sampling", dataType, sendFailureReason(firstErr, "server"), int(unsentPoints))
		}
	}

	return firstErr
}

func (ag *Aggregator) sendTailSamplingPackageContext(ctx context.Context,
	pickKey uint64, pkg *aggregate.DataPacket,
) error {
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

	var rawBody *pooledMarshalBody
	defer func() {
		if rawBody != nil {
			putMarshalBody(rawBody)
		}
	}()
	getRawBody := func() (*pooledMarshalBody, error) {
		if pkg.PayloadCompression == aggregate.PayloadCompressionNone {
			return body, nil
		}
		if rawBody != nil {
			return rawBody, nil
		}

		rawPacket := *pkg
		if err := aggregate.SetDataPacketPayloadCompression(&rawPacket, aggregate.PayloadCompressionNone); err != nil {
			return nil, err
		}
		marshaled, marshalErr := marshalDataPacketWithPool(&rawPacket)
		if marshalErr != nil {
			return nil, marshalErr
		}
		rawBody = marshaled
		return rawBody, nil
	}

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

		var resp *http.Response
		if pkg.PayloadCompression == aggregate.PayloadCompressionZstd && !ag.tailSamplingUsesLegacyProtocol(ep, time.Now()) {
			resp, err = writeTailSamplingPacket(ctx, ep, datakit.TailSamplingV2, category,
				pickKey, pointsCount, body, aggregate.TailSamplingPayloadCompressionZstd)
			if resp != nil && (resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnsupportedMediaType) {
				fallbackBody, fallbackErr := getRawBody()
				if fallbackErr != nil {
					err = fallbackErr
					resp = nil
				} else {
					ag.markTailSamplingLegacyProtocol(ep, time.Now())
					log.Infof("tail sampling endpoint does not support zstd payloads, retry with legacy protocol: endpoint=%s status=%d",
						ep.Host, resp.StatusCode)
					resp, err = writeTailSamplingPacket(ctx, ep, datakit.TailSampling, category,
						pickKey, pointsCount, fallbackBody, "")
				}
			} else if resp != nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPreconditionFailed) {
				ag.markTailSamplingZstdProtocol(ep)
			}
		} else {
			legacyBody, legacyErr := getRawBody()
			if legacyErr != nil {
				err = legacyErr
			} else {
				resp, err = writeTailSamplingPacket(ctx, ep, datakit.TailSampling, category,
					pickKey, pointsCount, legacyBody, "")
			}
		}
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

func writeTailSamplingPacket(ctx context.Context, ep *endpoint.EndPoint, api, category string,
	pickKey uint64, pointsCount int, body *pooledMarshalBody, payloadCompression string,
) (*http.Response, error) {
	if body == nil {
		return nil, fmt.Errorf("tail sampling request body is nil")
	}

	headers := map[string]string{
		aggregate.GuancePickKey: strconv.FormatUint(pickKey, 10),
		payloadSizeHeader:       strconv.Itoa(len(body.buf)),
	}
	if payloadCompression != "" {
		headers[aggregate.TailSamplingPayloadCompressionHeader] = payloadCompression
	}

	resp, _, err := ep.WriteAggrData(&endpoint.AggrData{
		Context:         ctx,
		API:             api,
		Category:        category,
		ContentType:     aggregatePayloadContentType,
		ContentEncoding: identityContentEncoding,
		Body:            body.buf,
		RawLen:          len(body.buf),
		Points:          pointsCount,
		Headers:         headers,
	})
	return resp, err
}

func (ag *Aggregator) tailSamplingUsesLegacyProtocol(ep *endpoint.EndPoint, now time.Time) bool {
	ag.tsProtocolMu.Lock()
	defer ag.tsProtocolMu.Unlock()

	until, ok := ag.tsLegacyUntil[ep]
	if !ok {
		return false
	}
	if now.Before(until) {
		return true
	}
	delete(ag.tsLegacyUntil, ep)
	return false
}

func (ag *Aggregator) markTailSamplingLegacyProtocol(ep *endpoint.EndPoint, now time.Time) {
	ag.tsProtocolMu.Lock()
	defer ag.tsProtocolMu.Unlock()
	if ag.tsLegacyUntil == nil {
		ag.tsLegacyUntil = make(map[*endpoint.EndPoint]time.Time)
	}
	ag.tsLegacyUntil[ep] = now.Add(tailSamplingLegacyProtocolTTL)
}

func (ag *Aggregator) markTailSamplingZstdProtocol(ep *endpoint.EndPoint) {
	ag.tsProtocolMu.Lock()
	delete(ag.tsLegacyUntil, ep)
	ag.tsProtocolMu.Unlock()
}

func (ag *Aggregator) SendMetricBatches(category string, batchMap map[uint64]*aggregate.Batchs) error {
	return ag.sendMetricBatches(category, batchMap, sinkHeaders{})
}

func (ag *Aggregator) sendMetricBatches(category string, batchMap map[uint64]*aggregate.Batchs, headers sinkHeaders) error {
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
			tasks = append(tasks, metricBatchSendTask{pickKey: pickKey, batch: splitBatch, sinkHeaders: headers})
		}
	}

	if len(tasks) == 0 {
		log.Debugf("skip sending metric batches: no valid batches")
		return nil
	}

	recordGeneratedBatches("metric", category, len(tasks))

	if err := ag.runAsyncSend(len(tasks), func(i int) error {
		task := tasks[i]
		return ag.sendMetricBatchWithHeaders(category, task.pickKey, task.batch, task.sinkHeaders)
	}); err != nil {
		log.Errorf("send metric batches failed: %v", err)
		return err
	}

	return nil
}

func (ag *Aggregator) sendMetricBatch(category string, pickKey uint64, batch *aggregate.Batchs) error {
	return ag.sendMetricBatchWithHeaders(category, pickKey, batch, sinkHeaders{})
}

func (ag *Aggregator) sendMetricBatchWithHeaders(category string, pickKey uint64, batch *aggregate.Batchs, headers sinkHeaders) error {
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

		httpHeaders := map[string]string{
			aggregate.GuancePickKey:    strconv.FormatUint(pickKey, 10),
			aggregate.GuanceRoutingKey: strconv.FormatUint(pickKey, 10),
			payloadSizeHeader:          strconv.Itoa(len(body.buf)),
		}
		headers.apply(httpHeaders)

		resp, respBody, err := ep.WriteAggrData(&endpoint.AggrData{
			API:             datakit.Aggregate,
			Category:        category,
			ContentType:     aggregatePayloadContentType,
			ContentEncoding: identityContentEncoding,
			Body:            body.buf,
			RawLen:          len(body.buf),
			Points:          pointsCount,
			Headers:         httpHeaders,
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

	if errors.Is(err, endpoint.ErrRequestTerminated) ||
		errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
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
	var parts []*aggregate.DataPacket
	_, err := walkDataPacketPartsBySize(context.Background(), pkg, maxRawBodySize, func(part *aggregate.DataPacket) error {
		parts = append(parts, part)
		return nil
	})
	if err != nil {
		log.Warnf("split tail sampling packet failed: group_id=%s err=%v", pkg.RawGroupId, err)
		return []*aggregate.DataPacket{pkg}
	}

	return parts
}

func walkDataPacketPartsBySize(ctx context.Context, pkg *aggregate.DataPacket, maxRawBodySize int,
	emit func(*aggregate.DataPacket) error,
) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return 0, context.Cause(ctx)
	}
	if pkg == nil {
		return 0, nil
	}
	if emit == nil {
		return 0, fmt.Errorf("tail sampling packet emitter is nil")
	}

	emitOriginal := func() (int, error) {
		if err := ctx.Err(); err != nil {
			return 0, context.Cause(ctx)
		}
		if err := emit(pkg); err != nil {
			return 0, err
		}
		return 1, nil
	}
	emitPart := func(part *aggregate.DataPacket) error {
		if pkg.PayloadCompression == aggregate.PayloadCompressionZstd {
			compressed, compression, err := aggregate.CompressPointsPayload(part.PointsPayload)
			if err != nil {
				return fmt.Errorf("compress split tail sampling payload: %w", err)
			}
			part.PointsPayload = compressed
			part.PayloadCompression = compression
		}
		return emit(part)
	}

	if maxRawBodySize <= 0 || pkg.PointCount <= 1 {
		return emitOriginal()
	}
	exceedsRawBodySize, sizeErr := dataPacketExceedsRawBodySize(pkg, maxRawBodySize)
	if sizeErr != nil {
		log.Warnf("inspect tail sampling decoded payload size failed: group_id=%s err=%v",
			pkg.RawGroupId, sizeErr)
		return emitOriginal()
	}
	if !exceedsRawBodySize {
		return emitOriginal()
	}

	// PickTrace 组包后 payload 可能已 zstd 压缩：拆分前先解压，
	// 拆分段在发送前重新压缩（zstd 帧不能从中间切开）。
	rawPayload, decompressErr := aggregate.DecompressPointsPayload(pkg.PointsPayload, pkg.PayloadCompression)
	if decompressErr != nil {
		log.Warnf("split tail sampling packet failed to decompress payload: group_id=%s err=%v",
			pkg.RawGroupId, decompressErr)
		return emitOriginal()
	}

	base := cloneDataPacketMeta(pkg)
	base.PointsPayload = nil
	base.PointCount = 0
	base.MaxPointTimeUnixNano = 0
	baseSize := base.Size()
	if baseSize >= maxRawBodySize {
		log.Warnf("tail sampling packet meta exceeds max body size: meta=%d limit=%d group_id=%s",
			baseSize, maxRawBodySize, pkg.RawGroupId)
		return emitOriginal()
	}

	// Validate before emitting because emit may send immediately. Discovering a
	// malformed point after an earlier split was sent would make fallback send
	// duplicate points.
	pointTimes, walkErr, decodeErr := validatePBPointsPayload(ctx, rawPayload)
	if err := ctx.Err(); err != nil {
		return 0, context.Cause(ctx)
	}
	if walkErr != nil {
		log.Warnf("split tail sampling packet failed to walk payload: group_id=%s err=%v", pkg.RawGroupId, walkErr)
		return emitOriginal()
	}
	if decodeErr != nil {
		log.Warnf("split tail sampling packet failed to decode point payload: group_id=%s err=%v", pkg.RawGroupId, decodeErr)
		return emitOriginal()
	}

	partPayloadCapacity := minInt(len(rawPayload), maxRawBodySize-baseSize)
	newPart := func() *aggregate.DataPacket {
		part := cloneDataPacketMeta(pkg)
		part.PointsPayload = make([]byte, 0, partPayloadCapacity)
		part.PointCount = 0
		part.MaxPointTimeUnixNano = 0
		return part
	}

	part := newPart()
	emitted := 0
	pointIndex := 0
	var emitErr error

	walkErr = point.WalkPBPointsPayload(rawPayload, func(raw []byte) bool {
		if err := ctx.Err(); err != nil {
			emitErr = context.Cause(ctx)
			return false
		}
		if len(raw) == 0 {
			return true
		}
		if pointIndex >= len(pointTimes) {
			emitErr = fmt.Errorf("tail sampling point count changed while splitting")
			return false
		}

		pointSize := protoListElemSize(len(raw))
		pointTime := pointTimes[pointIndex]
		pointIndex++
		maxPointTime := part.MaxPointTimeUnixNano
		if pointTime > maxPointTime {
			maxPointTime = pointTime
		}
		candidateSize := dataPacketPartSize(baseSize, len(part.PointsPayload)+pointSize,
			int(part.PointCount)+1, maxPointTime)
		if part.PointCount > 0 && candidateSize > maxRawBodySize {
			if err := emitPart(part); err != nil {
				emitErr = err
				return false
			}
			emitted++
			part = newPart()
			maxPointTime = pointTime
			candidateSize = dataPacketPartSize(baseSize, pointSize, 1, maxPointTime)
		}

		// The frame was validated above, so preserve its bytes instead of building
		// and marshaling another PBPoint object.
		part.PointsPayload = appendPBPointFrame(part.PointsPayload, raw)
		part.PointCount++
		part.MaxPointTimeUnixNano = maxPointTime

		if part.PointCount == 1 && candidateSize > maxRawBodySize {
			log.Warnf("single tail sampling point exceeds max body size: size=%d limit=%d group_id=%s",
				candidateSize, maxRawBodySize, pkg.RawGroupId)
			if err := emitPart(part); err != nil {
				emitErr = err
				return false
			}
			emitted++
			part = newPart()
		}
		return true
	})
	if emitErr != nil {
		return emitted, emitErr
	}
	if walkErr != nil {
		return emitted, fmt.Errorf("walk validated tail sampling payload: %w", walkErr)
	}

	if part.PointCount > 0 {
		if err := emitPart(part); err != nil {
			return emitted, err
		}
		emitted++
	}
	if emitted == 0 {
		return emitOriginal()
	}

	return emitted, nil
}

func validatePBPointsPayload(ctx context.Context, payload []byte) ([]int64, error, error) {
	var (
		pointTimes []int64
		decodeErr  error
	)

	walkErr := point.WalkPBPointsPayload(payload, func(raw []byte) bool {
		if err := ctx.Err(); err != nil {
			decodeErr = context.Cause(ctx)
			return false
		}
		if len(raw) == 0 {
			return true
		}

		pb := &point.PBPoint{}
		if err := pb.Unmarshal(raw); err != nil {
			decodeErr = err
			return false
		}
		pointTimes = append(pointTimes, pb.Time)
		return true
	})

	return pointTimes, walkErr, decodeErr
}

func appendPBPointFrame(dst, raw []byte) []byte {
	dst = append(dst, pbPointsArrayTag)
	dst = binary.AppendUvarint(dst, uint64(len(raw)))
	return append(dst, raw...)
}

func dataPacketPartSize(baseSize, pointsPayloadSize, pointCount int, maxPointTimeUnixNano int64) int {
	size := baseSize
	if pointsPayloadSize > 0 {
		size += 1 + uvarintSize(uint64(pointsPayloadSize)) + pointsPayloadSize
	}
	if pointCount > 0 {
		size += 1 + uvarintSize(uint64(pointCount))
	}
	if maxPointTimeUnixNano != 0 {
		size += 1 + uvarintSize(uint64(maxPointTimeUnixNano))
	}
	return size
}

func dataPacketExceedsRawBodySize(pkg *aggregate.DataPacket, maxRawBodySize int) (bool, error) {
	if pkg == nil || maxRawBodySize <= 0 {
		return false, nil
	}
	if pkg.Size() > maxRawBodySize {
		return true, nil
	}
	if pkg.PayloadCompression != aggregate.PayloadCompressionZstd {
		return false, nil
	}

	decodedBytes, err := aggregate.PointsPayloadDecodedSize(pkg.PointsPayload, pkg.PayloadCompression)
	if err != nil {
		return false, err
	}
	if decodedBytes > int64(maxRawBodySize) {
		return true, nil
	}

	base := cloneDataPacketMeta(pkg)
	base.PointsPayload = nil
	base.PointCount = 0
	base.MaxPointTimeUnixNano = 0
	rawSize := dataPacketPartSize(base.Size(), int(decodedBytes), int(pkg.PointCount), pkg.MaxPointTimeUnixNano)
	return rawSize > maxRawBodySize, nil
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
		GroupIdHash:             pkg.GroupIdHash,
		RawGroupId:              pkg.RawGroupId,
		Token:                   pkg.Token,
		Source:                  pkg.Source,
		DataType:                pkg.DataType,
		ConfigVersion:           pkg.ConfigVersion,
		HasError:                pkg.HasError,
		GroupKey:                pkg.GroupKey,
		PointCount:              pkg.PointCount,
		TraceStartTimeUnixNano:  pkg.TraceStartTimeUnixNano,
		TraceEndTimeUnixNano:    pkg.TraceEndTimeUnixNano,
		PointsPayload:           pkg.PointsPayload,
		MaxPointTimeUnixNano:    pkg.MaxPointTimeUnixNano,
		PredError:               pkg.PredError,
		PredHttpError:           pkg.PredHttpError,
		PredBizError:            pkg.PredBizError,
		PredTraceKeep:           pkg.PredTraceKeep,
		MaxSpanDurationUs:       pkg.MaxSpanDurationUs,
		RootDurationUs:          pkg.RootDurationUs,
		MaxNonrootDurationUs:    pkg.MaxNonrootDurationUs,
		PredicateSummaryVersion: pkg.PredicateSummaryVersion,
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
