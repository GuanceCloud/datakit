package postgresql

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

var postgreFields = map[string]interface{}{
	// db
	"numbackends":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of active connections to this database."},
	"xact_commit":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of transactions that have been committed in this database."},
	"xact_rollback": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of transactions that have been rolled back in this database."},
	"blks_read":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of disk blocks read in this database."},
	"blks_hit":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of times disk blocks were found in the buffer cache, preventing the need to read from the database."},
	"tup_returned":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of rows returned by queries in this database."},
	"tup_fetched":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of rows fetched by queries in this database."},
	"tup_inserted":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of rows inserted by queries in this database."},
	"tup_updated":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of rows updated by queries in this database."},
	"tup_deleted":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of rows deleted by queries in this database."},
	"deadlocks":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of deadlocks detected in this database."},
	"temp_bytes":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The amount of data written to temporary files by queries in this database."},
	"temp_files":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of temporary files created by queries in this database."},
	"database_size": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The disk space used by this database."},
	// bg_writer
	"checkpoints_timed":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.Count, Desc: "The number of scheduled checkpoints that were performed."},
	"checkpoints_req":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.Count, Desc: "The number of requested checkpoints that were performed."},
	"buffers_checkpoint":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.Count, Desc: "The number of buffers written during checkpoints."},
	"buffers_clean":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.Count, Desc: "The number of buffers written by the background writer."},
	"maxwritten_clean":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.Count, Desc: "The number of times the background writer stopped a cleaning scan due to writing too many buffers."},
	"buffers_backend":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.Count, Desc: "The number of buffers written directly by a backend."},
	"buffers_alloc":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.Count, Desc: "The number of buffers allocated"},
	"buffers_backend_fsync": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.Count, Desc: "The of times a backend had to execute its own fsync call instead of the background writer."},
	"checkpoint_write_time": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The total amount of checkpoint processing time spent writing files to disk."},
	"checkpoint_sync_time":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "The total amount of checkpoint processing time spent synchronizing files to disk."},

	// CONNECTION_METRICS
	"max_connections":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The maximum number of client connections allowed to this database."},
	"percent_usage_connections": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Count, Desc: "The number of connections to this database as a fraction of the maximum number of allowed connections."},
}
