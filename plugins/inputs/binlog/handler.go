// +build !386,!arm

package binlog

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"github.com/siddontang/go-mysql/schema"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type EventHandler interface {
	OnRotate(roateEvent *replication.RotateEvent) error
	// OnTableChanged is called when the table is created, altered, renamed or dropped.
	// You need to clear the associated data like cache with the table.
	// It will be called before OnDDL.
	OnTableChanged(schema string, table string) error
	OnDDL(nextPos mysql.Position, queryEvent *replication.QueryEvent) error
	OnRow(e *RowsEvent) error
	OnXID(nextPos mysql.Position) error
	OnGTID(gtid mysql.GTIDSet) error
	// OnPosSynced Use your own way to sync position. When force is true, sync position immediately.
	OnPosSynced(pos mysql.Position, set mysql.GTIDSet, force bool) error
	String() string
}

type DummyEventHandler struct {
}

func (h *DummyEventHandler) OnRotate(*replication.RotateEvent) error          { return nil }
func (h *DummyEventHandler) OnTableChanged(schema string, table string) error { return nil }
func (h *DummyEventHandler) OnDDL(nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	return nil
}
func (h *DummyEventHandler) OnRow(*RowsEvent) error                                { return nil }
func (h *DummyEventHandler) OnXID(mysql.Position) error                            { return nil }
func (h *DummyEventHandler) OnGTID(mysql.GTIDSet) error                            { return nil }
func (h *DummyEventHandler) OnPosSynced(mysql.Position, mysql.GTIDSet, bool) error { return nil }

func (h *DummyEventHandler) String() string { return "DummyEventHandler" }

// `SetEventHandler` registers the sync handler, you must register your
// own handler before starting Canal.
func (rb *RunningBinloger) SetEventHandler(h EventHandler) {
	rb.eventHandler = h
}

type MainEventHandler struct {
	DummyEventHandler

	rb *RunningBinloger
}

func (rb *RunningBinloger) convertValue(v interface{}, col *schema.TableColumn, forTag bool) interface{} {
	switch col.Type {
	case schema.TYPE_STRING:
		if v == nil {
			return ""
		} else {
			return fmt.Sprintf("%v", v)
		}
	case schema.TYPE_FLOAT, schema.TYPE_DECIMAL:
		if v == nil {
			moduleLogger.Warnf("null value of field %s, use default", col.Name)
			if forTag {
				return fmt.Sprintf("%v", rb.binlog.NullFloat)
			}
			return rb.binlog.NullFloat
		}
		vs := fmt.Sprintf("%v", v)
		if forTag {
			return vs
		}
		fv, err := strconv.ParseFloat(vs, 64)
		if err == nil {
			return fv
		} else {
			moduleLogger.Errorf("convert to float64 failed, %s", err)
		}
	case schema.TYPE_NUMBER, schema.TYPE_MEDIUM_INT, schema.TYPE_BIT:
		if v == nil {
			moduleLogger.Warnf("null value of field %s, use default", col.Name)
			if forTag {
				return fmt.Sprintf("%v", rb.binlog.NullInt)
			}
			return int64(rb.binlog.NullInt)
		}
		vs := fmt.Sprintf("%v", v)
		if forTag {
			return vs
		}
		nv, err := strconv.ParseInt(vs, 10, 64)
		if err == nil {
			return nv
		} else {
			moduleLogger.Errorf("convert to integer failed, %s", err)
		}
		// if strings.HasPrefix(strings.ToLower(col.col.RawType), "bool") {
		// 	if bv, err := strconv.Atoi(fmt.Sprintf("%v", row[col.index])); err == nil {
		// 		if bv > 0 {
		// 			v = "true"
		// 		} else {
		// 			v = "false"
		// 		}
		// 	}
		// }
	case schema.TYPE_ENUM:
		if v == nil {
			return ""
		}
		vs := fmt.Sprintf("%v", v)
		idx, err := strconv.ParseInt(vs, 10, 32)
		if err == nil && idx > 0 && (idx-1) < int64(len(col.EnumValues)) {
			return col.EnumValues[idx-1]
		} else {
			return vs
		}
	case schema.TYPE_SET:
		if v == nil {
			return ""
		}
		vs := fmt.Sprintf("%v", v)
		idx, err := strconv.ParseInt(vs, 10, 32)
		if err == nil && idx > 0 && (idx-1) < int64(len(col.SetValues)) {
			return col.SetValues[idx-1]
		} else {
			return vs
		}
	case schema.TYPE_DATE, schema.TYPE_DATETIME, schema.TYPE_TIMESTAMP:
		if v == nil {
			return ""
		}
		return fmt.Sprintf("%v", v)
	case schema.TYPE_JSON:
		if v == nil {
			return ""
		}
		if jv, ok := v.([]byte); ok {
			return string(jv)
		} else {
			if forTag {
				return ""
			}
		}
	default:
	}

	if forTag {
		return fmt.Sprintf("%v", v)
	}
	return v
}

func (h *MainEventHandler) OnRow(e *RowsEvent) error {

	moduleLogger.Debugf("eventAction: %s, eventType: %v", e.Action, e.Header.EventType)

	defer func() {
		if err := recover(); err != nil {
			moduleLogger.Errorf("panic %v", err)
		}
	}()

	switch e.Header.EventType {
	case replication.WRITE_ROWS_EVENTv0, replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2, replication.DELETE_ROWS_EVENTv0, replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2, replication.UPDATE_ROWS_EVENTv0, replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:

		targetDB := e.DBCfg

		if targetDB == nil || len(e.Rows) == 0 {
			return nil
		}

		var targetTable *TableConfig
		//检查table是否被过滤掉
		if len(targetDB.Tables) > 0 {
			for _, t := range targetDB.Tables {
				if t.Name == e.Table.Name {
					targetTable = t
					break
				}
			}
		}

		if targetTable == nil {
			moduleLogger.Debugf("no match table of %s", e.Table.Name)
			return nil
		}

		for _, le := range targetTable.ExcludeListenEvents {
			if le == e.Action {
				moduleLogger.Debugf("action [%s] is excluded", e.Action)
				return nil
			}
		}

		measureName := e.Table.Name

		if targetTable.Measurement != "" {
			measureName = targetTable.Measurement
		}

		fields := map[string]interface{}{}

		tags := map[string]string{
			"_host":            h.rb.cfg.Addr,
			"_collector_host_": config.Cfg.Hostname,
			"_db":              e.Table.Schema,
			"_table":           e.Table.Name,
			"_event":           e.Action,
		}

		type TableColumWrap struct {
			col   *schema.TableColumn
			index int
		}

		var tagCols []*TableColumWrap   //表中被配置为tag的列
		var fieldCols []*TableColumWrap //表中被配置为field的列

		if len(targetTable.Fields) == 0 {
			//没有指定field, 则默认所有字段为fields
			for i, c := range e.Table.Columns {
				fieldCols = append(fieldCols, &TableColumWrap{
					col:   &c,
					index: i,
				})
			}
		} else {
			for _, f := range targetTable.Fields {
				bhave := false
				for i, c := range e.Table.Columns {
					if f == c.Name {
						fieldCols = append(fieldCols, &TableColumWrap{
							col:   &c,
							index: i,
						})
						bhave = true
						break
					}
				}
				if !bhave {
					moduleLogger.Warnf("field %s not exist in table %s.%s", f, e.Table.Schema, e.Table.Name)
				}
			}

			for _, t := range targetTable.Tags {
				bhave := false
				for i, c := range e.Table.Columns {
					if t == c.Name {
						tagCols = append(tagCols, &TableColumWrap{
							col:   &c,
							index: i,
						})
						bhave = true
						break
					}
				}
				if !bhave {
					moduleLogger.Warnf("tag %s not exist in table %s.%s", t, e.Table.Schema, e.Table.Name)
				}
			}
		}

		var rows [][]interface{}
		switch e.Header.EventType {
		case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
			//update时会有两条新旧数据
			for i := 0; i < len(e.Rows); i += 2 {
				rows = append(rows, e.Rows[i+1])
			}
		default:
			rows = e.Rows
		}

		for _, row := range rows {

			for _, col := range tagCols {

				if col.index >= len(row) {
					moduleLogger.Warnf("index of %s is %d, but row length=%d", col.col.Name, col.index, len(row))
					continue
				}
				if e.checkIgnoreColumn(col.col) {
					continue
				}
				k := col.col.Name
				v := h.rb.convertValue(row[col.index], col.col, true)
				if vs, ok := v.(string); ok {
					tags[k] = vs
				}
			}

			if h.rb.binlog.tags != nil {
				for k, v := range h.rb.binlog.tags {
					tags[k] = v
				}
			}

			for _, col := range fieldCols {
				if e.checkIgnoreColumn(col.col) {
					continue
				}

				if col.index >= len(row) {
					moduleLogger.Warnf("index of field:%s is %d, but row length=%d", col.col.Name, col.index, len(row))
					continue
				}
				k := col.col.Name
				v := h.rb.convertValue(row[col.index], col.col, false)
				if v != nil {
					fields[k] = v
				}
			}

			finalFields := []string{}
			for fn, fv := range fields {
				finalFields = append(finalFields, fmt.Sprintf("%s=%v(%s)", fn, fv, reflect.TypeOf(fv)))
			}
			moduleLogger.Debugf("finalFields: %s", strings.Join(finalFields, ","))

			var evtime time.Time
			if e.Header.Timestamp == 0 {
				moduleLogger.Warnf("event time is zero")
				evtime = time.Now()
			} else {
				evtime = time.Unix(int64(e.Header.Timestamp), 0)
			}

			io.NamedFeedEx(inputName, datakit.Metric, measureName, tags, fields, evtime)
		}

	}

	return nil
}

func (h *MainEventHandler) String() string {
	return "MainEventHandler"
}
