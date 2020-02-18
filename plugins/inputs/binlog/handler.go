package binlog

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/siddontang/go-mysql/schema"

	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
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

func (h *MainEventHandler) OnRow(e *RowsEvent) error {

	log.Printf("D! [binlog] eventAction: %s, eventType: %v", e.Action, e.Header.EventType)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic %s", err)
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
			log.Printf("D! [binlog] no match table of %s", e.Table.Name)
			return nil
		}

		for _, le := range targetTable.ExcludeListenEvents {
			if le == e.Action {
				log.Printf("D! [binlog] action [%s] is excluded", e.Action)
				return nil
			}
		}

		measureName := e.Table.Name

		if targetTable.Measurement != "" {
			measureName = targetTable.Measurement
		}

		tags := map[string]string{}
		fields := map[string]interface{}{}

		hostname, _ := os.Hostname()

		tags = map[string]string{
			"_host":            h.rb.cfg.Addr,
			"_collector_host_": hostname,
			"_db":              e.Table.Schema,
			"_table":           e.Table.Name,
			"_event":           e.Action,
		}

		type TableColumWrap struct {
			col   schema.TableColumn
			index int
		}

		var tagCols []*TableColumWrap
		var fieldCols []*TableColumWrap

		if len(targetTable.Fields) == 0 {
			//没有指定field, 则默认所有字段为fields
			for i, c := range e.Table.Columns {
				fieldCols = append(fieldCols, &TableColumWrap{
					col:   c,
					index: i,
				})
			}
		} else {
			for _, f := range targetTable.Fields {
				for i, c := range e.Table.Columns {
					if f == c.Name {
						fieldCols = append(fieldCols, &TableColumWrap{
							col:   c,
							index: i,
						})
						break
					}
				}
			}

			for _, t := range targetTable.Tags {
				for i, c := range e.Table.Columns {
					if t == c.Name {
						tagCols = append(tagCols, &TableColumWrap{
							col:   c,
							index: i,
						})
						break
					}
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

		defvalForEmptyTag := ""

		for _, row := range rows {

			for _, col := range tagCols {
				if col.index >= len(row) || e.checkIgnoreColumn(&col.col) {
					continue
				}
				k := col.col.Name
				v := ""
				if row[col.index] == nil {
					v = defvalForEmptyTag
				} else {

					switch col.col.Type {
					case schema.TYPE_FLOAT, schema.TYPE_DECIMAL, schema.TYPE_NUMBER, schema.TYPE_MEDIUM_INT, schema.TYPE_BIT, schema.TYPE_SET:
						v = fmt.Sprintf("%v", row[col.index])
					case schema.TYPE_ENUM:
						enumindex := 0
						if iv, ok := (row[col.index]).(int64); ok {
							enumindex = int(iv)
						} else {
							if iv, ok := (row[col.index]).(int32); ok {
								enumindex = int(iv)
							} else {
								if iv, ok := (row[col.index]).(int); ok {
									enumindex = iv
								}
							}
						}

						if enumindex > 0 && (enumindex-1) < len(col.col.EnumValues) {
							v = col.col.EnumValues[enumindex-1]
						} else {
							v = defvalForEmptyTag
						}

					default:
						v = fmt.Sprintf("%s", row[col.index])
					}
				}

				tags[k] = v
			}

			if h.rb.binlog.tags != nil {
				for k, v := range h.rb.binlog.tags {
					tags[k] = v
				}
			}

			for _, col := range fieldCols {
				if col.index >= len(row) || e.checkIgnoreColumn(&col.col) {
					continue
				}
				k := col.col.Name
				var v interface{}
				if row[col.index] == nil {
					//v = ""
					switch col.col.Type {
					case schema.TYPE_FLOAT, schema.TYPE_DECIMAL:
						v = h.rb.binlog.NullFloat
					case schema.TYPE_NUMBER, schema.TYPE_MEDIUM_INT, schema.TYPE_BIT:
						v = h.rb.binlog.NullInt
					}
				} else {
					switch col.col.Type {
					case schema.TYPE_FLOAT, schema.TYPE_DECIMAL:
						v = row[col.index] // fmt.Sprintf("%v", row[col.index])
					case schema.TYPE_NUMBER, schema.TYPE_MEDIUM_INT, schema.TYPE_BIT, schema.TYPE_SET:
						v = row[col.index] // fmt.Sprintf("%vi", row[col.index])
						if strings.HasPrefix(strings.ToLower(col.col.RawType), "bool") {
							if bv, err := strconv.Atoi(fmt.Sprintf("%v", row[col.index])); err == nil {
								if bv > 0 {
									v = "true"
								} else {
									v = "false"
								}
							}
						}
					case schema.TYPE_ENUM:
						enumindex := 0
						if iv, ok := (row[col.index]).(int64); ok {
							enumindex = int(iv)
						} else {
							if iv, ok := (row[col.index]).(int32); ok {
								enumindex = int(iv)
							} else {
								if iv, ok := (row[col.index]).(int); ok {
									enumindex = iv
								}
							}
						}

						if enumindex > 0 && (enumindex-1) < len(col.col.EnumValues) {
							v = col.col.EnumValues[enumindex-1]
						} else {
							//v = `""`
						}

					default:
						v = row[col.index]
					}
				}
				if v != nil {
					fields[k] = v
				}
			}

			var evtime time.Time
			if e.Header.Timestamp == 0 {
				log.Printf("W! event time is zero")
				evtime = time.Now()
			} else {
				evtime = time.Unix(int64(e.Header.Timestamp), 0)
			}
			if h.rb.binlog.accumulator != nil {
				if len(fields) > 0 {
					h.rb.binlog.accumulator.AddFields(measureName, fields, tags, evtime)
				}
			}
		}

	}

	return nil
}

func (h *MainEventHandler) String() string {
	return "MainEventHandler"
}
