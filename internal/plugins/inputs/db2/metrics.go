// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package db2

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

////////////////////////////////////////////////////////////////////////////////

const (
	metricNameInstance       = "db2_instance"
	metricNameDatabase       = "db2_database"
	metricNameBufferPool     = "db2_buffer_pool"
	metricNameTableSpace     = "db2_table_space"
	metricNameTransactionLog = "db2_transaction_log"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

var (
	_ inputs.Measurement = (*instanceMeasurement)(nil)
	_ inputs.Measurement = (*databaseMeasurement)(nil)
	_ inputs.Measurement = (*bufferPoolMeasurement)(nil)
	_ inputs.Measurement = (*tableSpaceMeasurement)(nil)
	_ inputs.Measurement = (*transactionLogMeasurement)(nil)
)

////////////////////////////////////////////////////////////////////////////////

type instanceMeasurement struct {
	measurement
}

func (m *instanceMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *instanceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameInstance,
		Fields: map[string]interface{}{
			"connection_active": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The current number of connections."},
		},
		Tags: map[string]interface{}{
			"db2_server":  &inputs.TagInfo{Desc: "Server addr."},
			"db2_service": &inputs.TagInfo{Desc: "Server service."},
			"host":        &inputs.TagInfo{Desc: "Host name."},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type databaseMeasurement struct {
	measurement
}

func (m *databaseMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *databaseMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameDatabase,
		Fields: map[string]interface{}{
			"application_active":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of applications that are currently connected to the database."},
			"application_executing": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of applications for which the database manager is currently processing a request."},
			"backup_latest":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.TimestampSec, Desc: "The time elapsed since the latest database backup was completed."},
			"connection_max":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The highest number of simultaneous connections to the database since the database was activated."},
			"connection_total":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of connections to the database since the first connect, activate, or last reset (coordinator agents)."},
			"lock_active":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of locks currently held."},
			"lock_dead":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of deadlocks that have occurred."},
			"lock_pages":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The memory pages (4 KiB each) currently in use by the lock list."},
			"lock_timeouts":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of times that a request to lock an object timed out instead of being granted."},
			"lock_wait":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.TimestampSec, Desc: "The average wait time for a lock."},
			"lock_waiting":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of agents waiting on a lock."},
			"row_modified_total":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of rows inserted, updated, or deleted."},
			"row_reads_total":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of rows that had to be read in order to return result sets."},
			"row_returned_total":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of rows that have been selected by and returned to applications."},
			"status":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.UnknownUnit, Desc: "Database status. <br/>0: OK (ACTIVE, ACTIVE_STANDBY, STANDBY) <br/>1: WARNING (QUIESCE_PEND, ROLLFWD) <br/>2: CRITICAL (QUIESCED) <br/>3: UNKNOWN"},
		},
		Tags: map[string]interface{}{
			"db2_server":  &inputs.TagInfo{Desc: "Server addr."},
			"db2_service": &inputs.TagInfo{Desc: "Server service."},
			"host":        &inputs.TagInfo{Desc: "Host name."},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type bufferPoolMeasurement struct {
	measurement
}

func (m *bufferPoolMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *bufferPoolMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameBufferPool,
		Fields: map[string]interface{}{
			"bufferpool_column_hit_percent":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage of time that the database manager did not need to load a page from disk to service a column-organized table data page request."},
			"bufferpool_column_reads_logical":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of column-organized table data pages read from the logical table space containers for temporary, regular, and large table spaces."},
			"bufferpool_column_reads_physical": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of column-organized table data pages read from the physical table space containers for temporary, regular, and large table spaces."},
			"bufferpool_column_reads_total":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of column-organized table data pages read from the table space containers for temporary, regular, and large table spaces."},
			"bufferpool_data_hit_percent":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage of time that the database manager did not need to load a page from disk to service a data page request."},
			"bufferpool_data_reads_logical":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of data pages read from the logical table space containers for temporary, regular and large table spaces."},
			"bufferpool_data_reads_physical":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of data pages read from the physical table space containers for temporary, regular and large table spaces."},
			"bufferpool_data_reads_total":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of data pages read from the table space containers for temporary, regular and large table spaces."},
			"bufferpool_hit_percent":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage of time that the database manager did not need to load a page from disk to service a page request."},
			"bufferpool_index_hit_percent":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage of time that the database manager did not need to load a page from disk to service an index page request."},
			"bufferpool_index_reads_logical":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of index pages read from the logical table space containers for temporary, regular and large table spaces."},
			"bufferpool_index_reads_physical":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of index pages read from the physical table space containers for temporary, regular and large table spaces."},
			"bufferpool_index_reads_total":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of index pages read from the table space containers for temporary, regular and large table spaces."},
			"bufferpool_reads_logical":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of pages read from the logical table space containers for temporary, regular and large table spaces."},
			"bufferpool_reads_physical":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of pages read from the physical table space containers for temporary, regular and large table spaces."},
			"bufferpool_reads_total":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of pages read from the table space containers for temporary, regular and large table spaces."},
			"bufferpool_xda_hit_percent":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The percentage of time that the database manager did not need to load a page from disk to service an index page request."},
			"bufferpool_xda_reads_logical":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of data pages for XML storage objects (XDAs) read from the logical table space containers for temporary, regular and large table spaces."},
			"bufferpool_xda_reads_physical":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of data pages for XML storage objects (XDAs) read from the physical table space containers for temporary, regular and large table spaces."},
			"bufferpool_xda_reads_total":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of data pages for XML storage objects (XDAs) read from the table space containers for temporary, regular and large table spaces."},
		},
		Tags: map[string]interface{}{
			"bp_name":     &inputs.TagInfo{Desc: "Buffer pool name."},
			"db2_server":  &inputs.TagInfo{Desc: "Server addr."},
			"db2_service": &inputs.TagInfo{Desc: "Server service."},
			"host":        &inputs.TagInfo{Desc: "Host name."},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type tableSpaceMeasurement struct {
	measurement
}

func (m *tableSpaceMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *tableSpaceMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameTableSpace,
		Fields: map[string]interface{}{
			"tablespace_size":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total size of the table space in bytes."},
			"tablespace_usable":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total usable size of the table space in bytes."},
			"tablespace_used":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total used size of the table space in bytes."},
			"tablespace_utilized": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The utilization of the table space as a percentage."},
		},
		Tags: map[string]interface{}{
			"db2_server":      &inputs.TagInfo{Desc: "Server addr."},
			"db2_service":     &inputs.TagInfo{Desc: "Server service."},
			"host":            &inputs.TagInfo{Desc: "Host name."},
			"tablespace_name": &inputs.TagInfo{Desc: "Tablespace name."},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type transactionLogMeasurement struct {
	measurement
}

func (m *transactionLogMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *transactionLogMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameTransactionLog,
		Fields: map[string]interface{}{
			"log_available": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The disk blocks (4 KiB each) of active log space in the database that is not being used by uncommitted transactions."},
			"log_reads":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of log pages read from disk by the logger."},
			"log_used":      &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The disk blocks (4 KiB each) of active log space currently used in the database."},
			"log_utilized":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The utilization of active log space as a percentage."},
			"log_writes":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The number of log pages written to disk by the logger."},
		},
		Tags: map[string]interface{}{
			"db2_server":  &inputs.TagInfo{Desc: "Server addr."},
			"db2_service": &inputs.TagInfo{Desc: "Server service."},
			"host":        &inputs.TagInfo{Desc: "Host name."},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
