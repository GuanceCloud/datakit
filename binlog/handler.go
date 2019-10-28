package binlog

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/siddontang/go-mysql/schema"
	"gitlab.jiagouyun.com/cloudcare-tools/ftcollector/config"
	"gitlab.jiagouyun.com/cloudcare-tools/ftcollector/uploader"

	"github.com/siddontang/go-log/log"
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
func (c *Binloger) SetEventHandler(h EventHandler) {
	c.eventHandler = h
}

type MainEventHandler struct {
	DummyEventHandler

	binloger *Binloger
}

func tuneTagKVFieldK(n string) string {
	res := strings.Replace(n, ",", "\\,", -1)
	res = strings.Replace(res, " ", "\\ ", -1)
	res = strings.Replace(res, "=", "\\=", -1)
	return res
}

func tuneMeasureName(n string) string {
	res := strings.Replace(n, ",", "\\,", -1)
	res = strings.Replace(res, " ", "\\ ", -1)
	return res
}

func (h *MainEventHandler) OnRow(e *RowsEvent) error {

	defer func() {
		if e := recover(); e != nil {
			log.Errorf("%s", e)
		}
	}()

	switch e.Header.EventType {
	case replication.WRITE_ROWS_EVENTv0, replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2, replication.DELETE_ROWS_EVENTv0, replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2, replication.UPDATE_ROWS_EVENTv0, replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
		//log.Infof("OnRow: %v\n", e)

		target := e.Input

		if target == nil || len(e.Rows) == 0 {
			return nil
		}

		var targetTable *config.BinlogTable

		if len(target.Tables) > 0 {
			for _, t := range target.Tables {
				if t.Table == e.Table.Name {
					targetTable = t
					break
				}
			}
		}

		if len(target.ExcludeTables) > 0 {
			for _, t := range target.ExcludeTables {
				if t == e.Table.Name {
					targetTable = nil
					break
				}
			}
		}

		if targetTable == nil {
			return nil
		}

		for _, le := range targetTable.ExcludeListenEvents {
			if le == e.Action {
				return nil
			}
		}

		measureName := e.Table.Name

		if targetTable != nil && targetTable.Measurement != "" {
			measureName = targetTable.Measurement
		}

		hostname, _ := os.Hostname()

		systags := fmt.Sprintf(",_host_=%s", h.binloger.cfg.Addr)
		systags += fmt.Sprintf(",_collector_host_=%s", tuneTagKVFieldK(hostname))
		systags += fmt.Sprintf(",_db_=%s", tuneTagKVFieldK(e.Table.Schema))
		systags += fmt.Sprintf(",_table_=%s", tuneTagKVFieldK(e.Table.Name))
		systags += fmt.Sprintf(",_event_=%s", e.Action)

		type TableColumWrap struct {
			col   schema.TableColumn
			index int
		}

		var tagCols []*TableColumWrap
		var fieldCols []*TableColumWrap

		if targetTable == nil {
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

		//fmt.Printf("rowcount: %d\n", len(e.Rows))

		var rows [][]interface{}
		switch e.Header.EventType {
		case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
			for i := 0; i < len(e.Rows); i += 2 {
				rows = append(rows, e.Rows[i+1])
			}
		default:
			rows = e.Rows
		}

		for _, row := range rows {

			line := fmt.Sprintf("%s", tuneMeasureName(measureName))

			line += systags

			strTags := ""
			strFields := ""

			for _, col := range tagCols {
				if col.index >= len(row) || e.checkIgnoreColumn(&col.col) {
					continue
				}
				k := tuneTagKVFieldK(col.col.Name)
				v := ""
				if row[col.index] == nil {
					v = "NULL"
				} else {
					v = tuneTagKVFieldK(fmt.Sprintf("%v", row[col.index]))
				}
				strTags += fmt.Sprintf(",%s=%s", k, v)
			}

			for i, col := range fieldCols {
				if col.index >= len(row) || e.checkIgnoreColumn(&col.col) {
					continue
				}
				k := tuneTagKVFieldK(col.col.Name)
				v := ""
				if row[col.index] == nil {
					v = "\"NULL\""
				} else {
					switch col.col.Type {
					case schema.TYPE_FLOAT, schema.TYPE_DECIMAL:
						v = fmt.Sprintf("%v", row[col.index])
					case schema.TYPE_NUMBER, schema.TYPE_MEDIUM_INT, schema.TYPE_BIT, schema.TYPE_ENUM, schema.TYPE_SET:
						v = fmt.Sprintf("%vi", row[col.index])
					default:
						v = fmt.Sprintf("\"%v\"", row[col.index])
					}
				}

				var pair string
				if i == 0 {
					pair = fmt.Sprintf(" %s=%s", k, v)
				} else {
					pair = fmt.Sprintf(",%s=%s", k, v)
				}
				strFields += pair
			}

			line += strTags
			line += strFields

			//line += fmt.Sprintf(" %v", uint64(e.Header.Timestamp)*1000000000)
			line += fmt.Sprintf(" %v", time.Now().UnixNano())

			log.Debugf("*** %s", line)

			if h.binloger.storage != nil {
				h.binloger.storage.AddLog(&uploader.LogItem{
					Log: line,
				})

			}
		}

	}

	return nil
}

func (h *MainEventHandler) String() string {
	return "MainEventHandler"
}
