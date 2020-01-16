package binlog

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"

	"github.com/siddontang/go-mysql/schema"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/uploader"

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

func (h *MainEventHandler) OnRow(e *RowsEvent) error {

	h.binloger.logger.Debugf("binlog action: %s, eventType: %v", e.Action, e.Header.EventType)

	defer func() {
		if e := recover(); e != nil {
			log.Errorf("%s", e)
		}
	}()

	switch e.Header.EventType {
	case replication.WRITE_ROWS_EVENTv0, replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2, replication.DELETE_ROWS_EVENTv0, replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2, replication.UPDATE_ROWS_EVENTv0, replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:

		target := e.Input

		if target == nil || len(e.Rows) == 0 {
			return nil
		}

		var targetTable *BinlogTable

		if len(target.Tables) > 0 {
			for _, t := range target.Tables {
				if t.Name == e.Table.Name {
					targetTable = t
					break
				}
			}
		}

		if targetTable == nil {
			h.binloger.logger.Debugf("no match table of %s", e.Table.Name)
			return nil
		}

		for _, le := range targetTable.ExcludeListenEvents {
			if le == e.Action {
				h.binloger.logger.Debugf("action [%s] is excluded", e.Action)
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
			"_host":            h.binloger.cfg.Addr,
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

		var rows [][]interface{}
		switch e.Header.EventType {
		case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
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

			if config.Cfg.GlobalTags != nil {
				for k, v := range config.Cfg.GlobalTags {
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
						v = Cfg.NullFloat
					case schema.TYPE_NUMBER, schema.TYPE_MEDIUM_INT, schema.TYPE_BIT:
						v = Cfg.NullInt
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

			mt, _ := metric.New(measureName, tags, fields, time.Unix(0, time.Now().UnixNano()))

			serializer := influx.NewSerializer()
			line, err := serializer.Serialize(mt)

			if err != nil {
				h.binloger.logger.Errorf("fail to serialize as influx format: %s", err)
				return nil
			}

			h.binloger.logger.Debugf("binlog event[%s]: %s", e.Action, string(line))

			if h.binloger.storage != nil {
				h.binloger.storage.AddLog(&uploader.LogItem{
					Log: string(line),
				})
			}
		}

	}

	return nil
}

func (h *MainEventHandler) String() string {
	return "MainEventHandler"
}
