// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cibmi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

const measurementName = "ibm_i"

type systemInfo struct {
	HostName  string
	OSVersion int
	OSRelease int
}

type systemInfoRow struct {
	HostName  sql.NullString `db:"HOST_NAME"`
	OSVersion sql.NullString `db:"OS_VERSION"`
	OSRelease sql.NullString `db:"OS_RELEASE"`
}

type baseDiskUsageRow struct {
	ASPNumber           sql.NullString  `db:"ASP_NUMBER"`
	UnitNumber          sql.NullString  `db:"UNIT_NUMBER"`
	UnitType            sql.NullString  `db:"UNIT_TYPE"`
	SerialNumber        sql.NullString  `db:"SERIAL_NUMBER"`
	UnitStorageCapacity sql.NullInt64   `db:"UNIT_STORAGE_CAPACITY"`
	UnitSpaceAvailable  sql.NullInt64   `db:"UNIT_SPACE_AVAILABLE"`
	PercentUsed         sql.NullFloat64 `db:"PERCENT_USED"`
}

type diskUsageRow struct {
	ASPNumber          sql.NullString  `db:"ASP_NUMBER"`
	UnitNumber         sql.NullString  `db:"UNIT_NUMBER"`
	UnitType           sql.NullString  `db:"UNIT_TYPE"`
	SerialNumber       sql.NullString  `db:"SERIAL_NUMBER"`
	ResourceName       sql.NullString  `db:"RESOURCE_NAME"`
	ElapsedPercentBusy sql.NullFloat64 `db:"ELAPSED_PERCENT_BUSY"`
	ElapsedIORequests  sql.NullFloat64 `db:"ELAPSED_IO_REQUESTS"`
}

type cpuUsageRow struct {
	AverageCPUUtilization sql.NullFloat64 `db:"AVERAGE_CPU_UTILIZATION"`
	ConfiguredCPUs        sql.NullFloat64 `db:"CONFIGURED_CPUS"`
	CurrentCPUCapacity    sql.NullFloat64 `db:"CURRENT_CPU_CAPACITY"`
	PartitionID           sql.NullString  `db:"PARTITION_ID"`
	ElapsedCPUShared      sql.NullFloat64 `db:"ELAPSED_CPU_SHARED"`
	NormalizedCPUUsage    sql.NullFloat64 `db:"NORMALIZED_CPU_USAGE"`
}

type jobQJobStatusRow struct {
	JobID           sql.NullString  `db:"JOB_ID"`
	JobUser         sql.NullString  `db:"JOB_USER"`
	JobName         sql.NullString  `db:"JOB_NAME"`
	JobSubsystem    sql.NullString  `db:"JOB_SUBSYSTEM"`
	JobStatus       sql.NullString  `db:"JOB_STATUS"`
	JobQueueLibrary sql.NullString  `db:"JOB_QUEUE_LIBRARY"`
	JobQueueName    sql.NullString  `db:"JOB_QUEUE_NAME"`
	JobQueueStatus  sql.NullString  `db:"JOB_QUEUE_STATUS"`
	Status          sql.NullInt64   `db:"STATUS"`
	JobQDuration    sql.NullFloat64 `db:"JOBQ_DURATION"`
}

type activeJobStatusRow struct {
	JobID           sql.NullString  `db:"JOB_ID"`
	JobUser         sql.NullString  `db:"JOB_USER"`
	JobName         sql.NullString  `db:"JOB_NAME"`
	Subsystem       sql.NullString  `db:"SUBSYSTEM"`
	JobStatus       sql.NullString  `db:"JOB_STATUS"`
	JobActiveStatus sql.NullString  `db:"JOB_ACTIVE_STATUS"`
	Status          sql.NullInt64   `db:"STATUS"`
	CPUUsage        sql.NullFloat64 `db:"CPU_USAGE"`
	CPUUsagePct     sql.NullFloat64 `db:"CPU_USAGE_PCT"`
	ActiveDuration  sql.NullFloat64 `db:"ACTIVE_DURATION"`
}

type jobMemoryUsageRow struct {
	JobID            sql.NullString `db:"JOB_ID"`
	JobUser          sql.NullString `db:"JOB_USER"`
	JobName          sql.NullString `db:"JOB_NAME"`
	Subsystem        sql.NullString `db:"SUBSYSTEM"`
	JobActiveStatus  sql.NullString `db:"JOB_STATUS"`
	MemoryPool       sql.NullString `db:"MEMORY_POOL"`
	TemporaryStorage sql.NullInt64  `db:"TEMPORARY_STORAGE"`
}

type memoryInfoRow struct {
	PoolName      sql.NullString  `db:"POOL_NAME"`
	SubsystemName sql.NullString  `db:"SUBSYSTEM_NAME"`
	CurrentSize   sql.NullFloat64 `db:"CURRENT_SIZE"`
	ReservedSize  sql.NullFloat64 `db:"RESERVED_SIZE"`
	DefinedSize   sql.NullFloat64 `db:"DEFINED_SIZE"`
}

type subsystemRow struct {
	SubsystemDescription sql.NullString `db:"SUBSYSTEM_DESCRIPTION"`
	Active               sql.NullInt64  `db:"ACTIVE"`
	CurrentActiveJobs    sql.NullInt64  `db:"CURRENT_ACTIVE_JOBS"`
}

type jobQueueRow struct {
	JobQueueName   sql.NullString `db:"JOB_QUEUE_NAME"`
	JobQueueStatus sql.NullString `db:"JOB_QUEUE_STATUS"`
	SubsystemName  sql.NullString `db:"SUBSYSTEM_NAME"`
	NumberOfJobs   sql.NullInt64  `db:"NUMBER_OF_JOBS"`
	ReleasedJobs   sql.NullInt64  `db:"RELEASED_JOBS"`
	ScheduledJobs  sql.NullInt64  `db:"SCHEDULED_JOBS"`
	HeldJobs       sql.NullInt64  `db:"HELD_JOBS"`
}

type messageQueueRow struct {
	MessageQueueName     sql.NullString `db:"MESSAGE_QUEUE_NAME"`
	MessageQueueLibrary  sql.NullString `db:"MESSAGE_QUEUE_LIBRARY"`
	MessageQueueSize     sql.NullInt64  `db:"MESSAGE_QUEUE_SIZE"`
	CriticalMessageQueue sql.NullInt64  `db:"CRITICAL_MESSAGE_QUEUE_SIZE"`
}

func (ipt *Input) collectMetricPoints(ctx context.Context, ts time.Time) ([]*point.Point, []*point.Point, error) {
	var (
		metricPts []*point.Point
		errs      []string
		up        = int64(1)
	)

	if err := ipt.ensureSystemInfo(ctx); err != nil {
		errs = append(errs, err.Error())
		metricPts = append(metricPts, ipt.buildUpPoint(ts, 0))
		return metricPts, nil, joinErrors(errs)
	}

	for _, query := range ipt.Queries {
		pts, err := ipt.collectQueryPoints(ctx, ts, query)
		if err != nil {
			errs = append(errs, err.Error())
			if query == queryCPUUsage {
				up = 0
			}
			continue
		}
		metricPts = append(metricPts, pts...)
	}

	metricPts = append(metricPts, ipt.buildUpPoint(ts, up))

	return metricPts, nil, joinErrors(errs)
}

func (ipt *Input) collectQueryPoints(ctx context.Context, ts time.Time, query string) ([]*point.Point, error) {
	switch query {
	case queryDiskUsage:
		return ipt.collectDiskUsagePoints(ctx, ts)
	case queryCPUUsage:
		return ipt.collectCPUUsagePoints(ctx, ts)
	case queryJobQJobStatus:
		return ipt.collectJobQJobStatusPoints(ctx, ts)
	case queryActiveJobStatus:
		return ipt.collectActiveJobStatusPoints(ctx, ts)
	case queryJobMemoryUsage:
		return ipt.collectJobMemoryUsagePoints(ctx, ts)
	case queryMemoryInfo:
		return ipt.collectMemoryInfoPoints(ctx, ts)
	case querySubsystem:
		if !ipt.systemInfo.Is73OrHigher() {
			return nil, nil
		}
		return ipt.collectSubsystemPoints(ctx, ts)
	case queryJobQueue:
		return ipt.collectJobQueuePoints(ctx, ts)
	case queryMessageQueue:
		return ipt.collectMessageQueuePoints(ctx, ts)
	default:
		return nil, fmt.Errorf("unknown query %q", query)
	}
}

func (ipt *Input) ensureSystemInfo(ctx context.Context) error {
	if ipt.systemInfo != nil {
		return nil
	}

	rows := []systemInfoRow{}
	if err := ipt.selectContext(ctx, &rows, systemInfoSQL, ipt.QueryTimeout); err != nil {
		return fmt.Errorf("collect system info: %w", err)
	}
	if len(rows) != 1 {
		return fmt.Errorf("collect system info: expected 1 row, got %d", len(rows))
	}

	info, err := parseSystemInfo(rows[0])
	if err != nil {
		return fmt.Errorf("collect system info: %w", err)
	}
	ipt.systemInfo = info
	return nil
}

func (ipt *Input) collectDiskUsagePoints(ctx context.Context, ts time.Time) ([]*point.Point, error) {
	var (
		pts  []*point.Point
		errs []string
	)

	baseRows := []baseDiskUsageRow{}
	baseSQL := baseDiskUsage72SQL
	if ipt.systemInfo.Is73OrHigher() {
		baseSQL = baseDiskUsage73SQL
	}
	if err := ipt.selectContext(ctx, &baseRows, baseSQL, ipt.QueryTimeout); err != nil {
		errs = append(errs, fmt.Sprintf("collect base disk usage: %v", err))
	} else {
		for _, row := range baseRows {
			pts = append(pts, ipt.buildBaseDiskUsagePoint(row, ts))
		}
	}

	if ipt.systemInfo.Is73OrHigher() {
		usageRows := []diskUsageRow{}
		if err := ipt.selectContext(ctx, &usageRows, diskUsageSQL, ipt.QueryTimeout); err != nil {
			errs = append(errs, fmt.Sprintf("collect disk usage: %v", err))
		} else {
			for _, row := range usageRows {
				pts = append(pts, ipt.buildDiskUsagePoint(row, ts))
			}
		}
	}

	return pts, joinErrors(errs)
}

func (ipt *Input) collectCPUUsagePoints(ctx context.Context, ts time.Time) ([]*point.Point, error) {
	rows := []cpuUsageRow{}
	if err := ipt.selectContext(ctx, &rows, cpuUsageSQL, ipt.QueryTimeout); err != nil {
		return nil, fmt.Errorf("collect CPU usage: %w", err)
	}
	pts := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		pts = append(pts, ipt.buildCPUUsagePoint(row, ts))
	}
	return pts, nil
}

func (ipt *Input) collectJobQJobStatusPoints(ctx context.Context, ts time.Time) ([]*point.Point, error) {
	rows := []jobQJobStatusRow{}
	if err := ipt.selectContext(ctx, &rows, jobQJobStatusSQL, ipt.JobQueryTimeout); err != nil {
		return nil, fmt.Errorf("collect job queue job status: %w", err)
	}
	pts := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		pts = append(pts, ipt.buildJobQJobStatusPoint(row, ts))
	}
	return pts, nil
}

func (ipt *Input) collectActiveJobStatusPoints(ctx context.Context, ts time.Time) ([]*point.Point, error) {
	rows := []activeJobStatusRow{}
	if err := ipt.selectContext(ctx, &rows, activeJobStatusSQL, ipt.JobQueryTimeout); err != nil {
		return nil, fmt.Errorf("collect active job status: %w", err)
	}
	pts := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		pts = append(pts, ipt.buildActiveJobStatusPoint(row, ts))
	}
	return pts, nil
}

func (ipt *Input) collectJobMemoryUsagePoints(ctx context.Context, ts time.Time) ([]*point.Point, error) {
	rows := []jobMemoryUsageRow{}
	if err := ipt.selectContext(ctx, &rows, jobMemoryUsageSQL, ipt.JobQueryTimeout); err != nil {
		return nil, fmt.Errorf("collect job memory usage: %w", err)
	}
	pts := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		pts = append(pts, ipt.buildJobMemoryUsagePoint(row, ts))
	}
	return pts, nil
}

func (ipt *Input) collectMemoryInfoPoints(ctx context.Context, ts time.Time) ([]*point.Point, error) {
	rows := []memoryInfoRow{}
	if err := ipt.selectContext(ctx, &rows, memoryInfoSQL, ipt.QueryTimeout); err != nil {
		return nil, fmt.Errorf("collect memory info: %w", err)
	}
	pts := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		pts = append(pts, ipt.buildMemoryInfoPoint(row, ts))
	}
	return pts, nil
}

func (ipt *Input) collectSubsystemPoints(ctx context.Context, ts time.Time) ([]*point.Point, error) {
	rows := []subsystemRow{}
	if err := ipt.selectContext(ctx, &rows, subsystemSQL, ipt.QueryTimeout); err != nil {
		return nil, fmt.Errorf("collect subsystem: %w", err)
	}
	pts := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		pts = append(pts, ipt.buildSubsystemPoint(row, ts))
	}
	return pts, nil
}

func (ipt *Input) collectJobQueuePoints(ctx context.Context, ts time.Time) ([]*point.Point, error) {
	rows := []jobQueueRow{}
	if err := ipt.selectContext(ctx, &rows, jobQueueSQL, ipt.QueryTimeout); err != nil {
		return nil, fmt.Errorf("collect job queue: %w", err)
	}
	pts := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		pts = append(pts, ipt.buildJobQueuePoint(row, ts))
	}
	return pts, nil
}

func (ipt *Input) collectMessageQueuePoints(ctx context.Context, ts time.Time) ([]*point.Point, error) {
	query, args := messageQueueSQL(ipt.SeverityThreshold, ipt.MessageQueues)
	rows := []messageQueueRow{}
	if err := ipt.selectContext(ctx, &rows, query, ipt.SystemMQQueryTimeout, args...); err != nil {
		return nil, fmt.Errorf("collect message queue: %w", err)
	}
	pts := make([]*point.Point, 0, len(rows))
	for _, row := range rows {
		pts = append(pts, ipt.buildMessageQueuePoint(row, ts))
	}
	return pts, nil
}

func (ipt *Input) buildBaseDiskUsagePoint(row baseDiskUsageRow, ts time.Time) *point.Point {
	tags := ipt.aspTags(row.ASPNumber, row.UnitNumber, row.UnitType, row.SerialNumber, sql.NullString{})
	var kvs point.KVs
	addTags(&kvs, tags)
	addIntField(&kvs, "asp_unit_storage_capacity", row.UnitStorageCapacity)
	addIntField(&kvs, "asp_unit_space_available", row.UnitSpaceAvailable)
	addFloatField(&kvs, "asp_percent_used", row.PercentUsed)
	return point.NewPoint(measurementName, kvs, metricOptions(ts)...)
}

func (ipt *Input) buildDiskUsagePoint(row diskUsageRow, ts time.Time) *point.Point {
	tags := ipt.aspTags(row.ASPNumber, row.UnitNumber, row.UnitType, row.SerialNumber, row.ResourceName)
	var kvs point.KVs
	addTags(&kvs, tags)
	addFloatField(&kvs, "asp_percent_busy", row.ElapsedPercentBusy)
	addFloatField(&kvs, "asp_io_requests_per_s", row.ElapsedIORequests)
	return point.NewPoint(measurementName, kvs, metricOptions(ts)...)
}

func (ipt *Input) buildCPUUsagePoint(row cpuUsageRow, ts time.Time) *point.Point {
	tags := ipt.commonTags()
	addIfNotEmpty(tags, "partition_id", nullString(row.PartitionID))
	var kvs point.KVs
	addTags(&kvs, tags)
	addFloatField(&kvs, "system_cpu_usage", row.AverageCPUUtilization)
	addFloatField(&kvs, "system_configured_cpus", row.ConfiguredCPUs)
	addFloatField(&kvs, "system_current_cpu_capacity", row.CurrentCPUCapacity)
	addFloatField(&kvs, "system_shared_cpu_usage", row.ElapsedCPUShared)
	addFloatField(&kvs, "system_normalized_cpu_usage", row.NormalizedCPUUsage)
	return point.NewPoint(measurementName, kvs, metricOptions(ts)...)
}

func (ipt *Input) buildJobQJobStatusPoint(row jobQJobStatusRow, ts time.Time) *point.Point {
	tags := ipt.jobTags(row.JobID, row.JobUser, row.JobName)
	addIfNotEmpty(tags, "subsystem_name", nullString(row.JobSubsystem))
	addIfNotEmpty(tags, "job_status", nullString(row.JobStatus))
	addIfNotEmpty(tags, "job_queue_library", nullString(row.JobQueueLibrary))
	addIfNotEmpty(tags, "job_queue_name", nullString(row.JobQueueName))
	addIfNotEmpty(tags, "job_queue_status", nullString(row.JobQueueStatus))
	var kvs point.KVs
	addTags(&kvs, tags)
	addIntField(&kvs, "job_status_value", row.Status)
	addFloatField(&kvs, "job_queue_duration", row.JobQDuration)
	return point.NewPoint(measurementName, kvs, metricOptions(ts)...)
}

func (ipt *Input) buildActiveJobStatusPoint(row activeJobStatusRow, ts time.Time) *point.Point {
	tags := ipt.jobTags(row.JobID, row.JobUser, row.JobName)
	addIfNotEmpty(tags, "subsystem_name", nullString(row.Subsystem))
	addIfNotEmpty(tags, "job_status", nullString(row.JobStatus))
	addIfNotEmpty(tags, "job_active_status", nullString(row.JobActiveStatus))
	var kvs point.KVs
	addTags(&kvs, tags)
	addIntField(&kvs, "job_status_value", row.Status)
	addFloatField(&kvs, "job_cpu_usage", row.CPUUsage)
	addFloatField(&kvs, "job_cpu_usage_pct", row.CPUUsagePct)
	addFloatField(&kvs, "job_active_duration", row.ActiveDuration)
	return point.NewPoint(measurementName, kvs, metricOptions(ts)...)
}

func (ipt *Input) buildJobMemoryUsagePoint(row jobMemoryUsageRow, ts time.Time) *point.Point {
	tags := ipt.jobTags(row.JobID, row.JobUser, row.JobName)
	addIfNotEmpty(tags, "subsystem_name", nullString(row.Subsystem))
	addIfNotEmpty(tags, "job_active_status", nullString(row.JobActiveStatus))
	addIfNotEmpty(tags, "memory_pool_name", nullString(row.MemoryPool))
	var kvs point.KVs
	addTags(&kvs, tags)
	addIntField(&kvs, "job_temp_storage", row.TemporaryStorage)
	return point.NewPoint(measurementName, kvs, metricOptions(ts)...)
}

func (ipt *Input) buildMemoryInfoPoint(row memoryInfoRow, ts time.Time) *point.Point {
	tags := ipt.commonTags()
	addIfNotEmpty(tags, "pool_name", nullString(row.PoolName))
	addIfNotEmpty(tags, "subsystem_name", nullString(row.SubsystemName))
	var kvs point.KVs
	addTags(&kvs, tags)
	addFloatField(&kvs, "pool_size", row.CurrentSize)
	addFloatField(&kvs, "pool_reserved_size", row.ReservedSize)
	addFloatField(&kvs, "pool_defined_size", row.DefinedSize)
	return point.NewPoint(measurementName, kvs, metricOptions(ts)...)
}

func (ipt *Input) buildSubsystemPoint(row subsystemRow, ts time.Time) *point.Point {
	tags := ipt.commonTags()
	addIfNotEmpty(tags, "subsystem_name", nullString(row.SubsystemDescription))
	var kvs point.KVs
	addTags(&kvs, tags)
	addIntField(&kvs, "subsystem_active", row.Active)
	addIntField(&kvs, "subsystem_active_jobs", row.CurrentActiveJobs)
	return point.NewPoint(measurementName, kvs, metricOptions(ts)...)
}

func (ipt *Input) buildJobQueuePoint(row jobQueueRow, ts time.Time) *point.Point {
	tags := ipt.commonTags()
	addIfNotEmpty(tags, "job_queue_name", nullString(row.JobQueueName))
	addIfNotEmpty(tags, "job_queue_status", nullString(row.JobQueueStatus))
	addIfNotEmpty(tags, "subsystem_name", nullString(row.SubsystemName))
	var kvs point.KVs
	addTags(&kvs, tags)
	addIntField(&kvs, "job_queue_size", row.NumberOfJobs)
	addIntField(&kvs, "job_queue_released_size", row.ReleasedJobs)
	addIntField(&kvs, "job_queue_scheduled_size", row.ScheduledJobs)
	addIntField(&kvs, "job_queue_held_size", row.HeldJobs)
	return point.NewPoint(measurementName, kvs, metricOptions(ts)...)
}

func (ipt *Input) buildMessageQueuePoint(row messageQueueRow, ts time.Time) *point.Point {
	tags := ipt.commonTags()
	addIfNotEmpty(tags, "message_queue_name", nullString(row.MessageQueueName))
	addIfNotEmpty(tags, "message_queue_library", nullString(row.MessageQueueLibrary))
	var kvs point.KVs
	addTags(&kvs, tags)
	addIntField(&kvs, "message_queue_size", row.MessageQueueSize)
	addIntField(&kvs, "message_queue_critical_size", row.CriticalMessageQueue)
	return point.NewPoint(measurementName, kvs, metricOptions(ts)...)
}

func (ipt *Input) aspTags(aspNumber, unitNumber, unitType, serialNumber, resourceName sql.NullString) map[string]string {
	tags := ipt.commonTags()
	addIfNotEmpty(tags, "asp_number", nullString(aspNumber))
	addIfNotEmpty(tags, "unit_number", nullString(unitNumber))
	addIfNotEmpty(tags, "unit_type", nullString(unitType))
	addIfNotEmpty(tags, "serial_number", nullString(serialNumber))
	addIfNotEmpty(tags, "resource_name", nullString(resourceName))
	return tags
}

func (ipt *Input) commonTags() map[string]string {
	tags := inputsMergeTags(nil, ipt.Tags, ipt.Host)
	if tags["host"] == "" {
		tags["host"] = firstNonEmpty(ipt.Host, ipt.systemHostName())
	}
	return tags
}

func (ipt *Input) systemHostName() string {
	if ipt.systemInfo == nil {
		return ""
	}
	return ipt.systemInfo.HostName
}

func (ipt *Input) jobTags(jobID, jobUser, jobName sql.NullString) map[string]string {
	tags := ipt.commonTags()
	addIfNotEmpty(tags, "job_id", nullString(jobID))
	addIfNotEmpty(tags, "job_user", nullString(jobUser))
	addIfNotEmpty(tags, "job_name", nullString(jobName))
	return tags
}

func (s *systemInfo) Is73OrHigher() bool {
	if s == nil {
		return true
	}
	return s.OSVersion > 7 || (s.OSVersion == 7 && s.OSRelease >= 3)
}

func parseSystemInfo(row systemInfoRow) (*systemInfo, error) {
	version, err := parseSystemInfoInt("OS_VERSION", row.OSVersion)
	if err != nil {
		return nil, err
	}
	release, err := parseSystemInfoInt("OS_RELEASE", row.OSRelease)
	if err != nil {
		return nil, err
	}
	return &systemInfo{
		HostName:  nullString(row.HostName),
		OSVersion: version,
		OSRelease: release,
	}, nil
}

func parseSystemInfoInt(name string, value sql.NullString) (int, error) {
	raw := nullString(value)
	if raw == "" {
		return 0, fmt.Errorf("%s is empty", name)
	}
	out, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	return out, nil
}

func addTags(kvs *point.KVs, tags map[string]string) {
	for k, v := range tags {
		if k == "" || v == "" {
			continue
		}
		*kvs = kvs.AddTag(k, v)
	}
}

func addFloatField(kvs *point.KVs, name string, v sql.NullFloat64) {
	if v.Valid {
		*kvs = kvs.Set(name, v.Float64)
	}
}

func addIntField(kvs *point.KVs, name string, v sql.NullInt64) {
	if v.Valid {
		*kvs = kvs.Set(name, v.Int64)
	}
}

func metricOptions(ts time.Time) []point.Option {
	return append(point.DefaultMetricOptions(), point.WithTime(ts))
}

func nullString(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return strings.TrimSpace(v.String)
}

func joinJobID(jobNumber, jobUser, jobName string) string {
	parts := []string{jobNumber, jobUser, jobName}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if parts[0] == "" && parts[1] == "" && parts[2] == "" {
		return ""
	}
	return strings.Join(parts, "/")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func addIfNotEmpty(tags map[string]string, key, value string) {
	if strings.TrimSpace(value) != "" {
		tags[key] = strings.TrimSpace(value)
	}
}

func connValue(dsn, key string) string {
	for _, part := range strings.Split(dsn, ";") {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(k), key) {
			return strings.Trim(strings.TrimSpace(v), "{}")
		}
	}
	return ""
}

func joinErrors(errs []string) error {
	filtered := make([]string, 0, len(errs))
	for _, err := range errs {
		if strings.TrimSpace(err) != "" {
			filtered = append(filtered, err)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return errors.New(strings.Join(filtered, "; "))
}
