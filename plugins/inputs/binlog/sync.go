// +build !386,!arm

package binlog

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser/ast"
	uuid "github.com/satori/go.uuid"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"github.com/siddontang/go-mysql/schema"
)

var (
	UnknownTableRetryPeriod = time.Second * time.Duration(10)
	ErrExcludedTable        = errors.New("excluded table meta")
)

func (rb *RunningBinloger) startSyncer() (*replication.BinlogStreamer, error) {
	gset := rb.master.GTIDSet()
	if gset == nil {
		pos := rb.master.Position()
		s, err := rb.syncer.StartSync(pos)
		if err != nil {
			return nil, fmt.Errorf("start sync replication at binlog %v error %v", pos, err)
		}
		moduleLogger.Infof("start binlog from %v", pos)
		return s, nil
	} else {
		gsetClone := gset.Clone()
		s, err := rb.syncer.StartSyncGTID(gset)
		if err != nil {
			return nil, errors.Errorf("start sync replication at GTID set %v error %v", gset, err)
		}
		moduleLogger.Infof("start sync binlog at GTID set %v", gsetClone)
		return s, nil
	}
}

func (rb *RunningBinloger) doSync(ctx context.Context) error {

	s, err := rb.startSyncer()
	if err != nil {
		moduleLogger.Errorf("start sync failed, %s", err)
		return err
	}

	savePos := false
	force := false

	// The name of the binlog file received in the fake rotate event.
	// It must be preserved until the new position is saved.
	fakeRotateLogName := ""

	defer func() {
		rb.saveMasterStatus(rb.master)
		if rb.syncer != nil {
			rb.syncer.Close()
		}
	}()

	for {
		ev, err := s.GetEvent(ctx)
		if err != nil {
			moduleLogger.Errorf("GetEvent failed, %s", err)
			return err
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		// Update the delay between the Canal and the Master before the handler hooks are called
		rb.updateReplicationDelay(ev)

		// If log pos equals zero then the received event is a fake rotate event and
		// contains only a name of the next binlog file
		// See https://github.com/mysql/mysql-server/blob/8e797a5d6eb3a87f16498edcb7261a75897babae/sql/rpl_binlog_sender.h#L235
		// and https://github.com/mysql/mysql-server/blob/8cc757da3d87bf4a1f07dcfb2d3c96fed3806870/sql/rpl_binlog_sender.cc#L899
		if ev.Header.LogPos == 0 {
			switch e := ev.Event.(type) {
			case *replication.RotateEvent:
				fakeRotateLogName = string(e.NextLogName)
				moduleLogger.Debugf("received fake rotate event, next log name is %s", e.NextLogName)
			}

			continue
		}

		savePos = false
		force = false
		pos := rb.master.Position()

		curPos := pos.Pos
		// next binlog pos
		pos.Pos = ev.Header.LogPos

		// new file name received in the fake rotate event
		if fakeRotateLogName != "" {
			pos.Name = fakeRotateLogName
		}

		// We only save position with RotateEvent and XIDEvent.
		// For RowsEvent, we can't save the position until meeting XIDEvent
		// which tells the whole transaction is over.
		// TODO: If we meet any DDL query, we must save too.
		switch e := ev.Event.(type) {
		case *replication.RotateEvent:
			pos.Name = string(e.NextLogName)
			pos.Pos = uint32(e.Position)
			moduleLogger.Debugf("rotate binlog to %v", pos)
			savePos = true
			force = true
			if err = rb.eventHandler.OnRotate(e); err != nil {
				return err
			}
		case *replication.RowsEvent:
			// we only focus row based event
			err = rb.handleRowsEvent(ev)
			if err != nil {
				e := errors.Cause(err)
				// if error is not ErrExcludedTable or ErrTableNotExist or ErrMissingTableMeta, stop canal
				if e != ErrExcludedTable &&
					e != schema.ErrTableNotExist &&
					e != schema.ErrMissingTableMeta {
					moduleLogger.Errorf("handle rows event at (%s, %d) error %v", pos.Name, curPos, err)
					return err
				} else {
					moduleLogger.Errorf("handleRowsEvent failed, %s", err)
				}
			}
			continue
		case *replication.XIDEvent:
			savePos = true
			// try to save the position later
			if err := rb.eventHandler.OnXID(pos); err != nil {
				return err
			}
			if e.GSet != nil {
				rb.master.UpdateGTIDSet(e.GSet)
			}
		case *replication.MariadbGTIDEvent:
			// try to save the GTID later
			gtid, err := mysql.ParseMariadbGTIDSet(e.GTID.String())
			if err != nil {
				return err
			}
			if err := rb.eventHandler.OnGTID(gtid); err != nil {
				return err
			}
		case *replication.GTIDEvent:
			u, _ := uuid.FromBytes(e.SID)
			gtid, err := mysql.ParseMysqlGTIDSet(fmt.Sprintf("%s:%d", u.String(), e.GNO))
			if err != nil {
				return err
			}
			if err := rb.eventHandler.OnGTID(gtid); err != nil {
				return err
			}
		case *replication.QueryEvent:
			//log.Printf("D! [binlog] query event come: %s", string(e.Query))
			stmts, _, err := rb.parser.Parse(string(e.Query), "", "")
			if err != nil {
				moduleLogger.Errorf("parse query(%s) err %v, will skip this event", e.Query, err)
				continue
			}
			for _, stmt := range stmts {

				nodes := parseStmt(stmt)

				for _, node := range nodes {
					if node.db == "" {
						node.db = string(e.Schema)
					}
					if rb.checkTableMatch(node.db, node.table) == nil {
						continue
					}
					if err = rb.updateTable(node.db, node.table); err != nil {
						return err
					}
				}
				if len(nodes) > 0 {
					savePos = true
					force = true
					// Now we only handle Table Changed DDL, maybe we will support more later.
					if err = rb.eventHandler.OnDDL(pos, e); err != nil {
						return err
					}
				}
			}
			if savePos && e.GSet != nil {
				rb.master.UpdateGTIDSet(e.GSet)
			}
		default:
			continue
		}

		if savePos {
			rb.master.Update(pos)
			rb.master.UpdateTimestamp(ev.Header.Timestamp)
			fakeRotateLogName = ""

			rb.saveMasterStatus(rb.master)

			if err := rb.eventHandler.OnPosSynced(pos, rb.master.GTIDSet(), force); err != nil {
				return err
			}
		}
	}

}

type node struct {
	db    string
	table string
}

func parseStmt(stmt ast.StmtNode) (ns []*node) {
	switch t := stmt.(type) {
	case *ast.RenameTableStmt:
		for _, tableInfo := range t.TableToTables {
			n := &node{
				db:    tableInfo.OldTable.Schema.String(),
				table: tableInfo.OldTable.Name.String(),
			}
			ns = append(ns, n)
		}
	case *ast.AlterTableStmt:
		n := &node{
			db:    t.Table.Schema.String(),
			table: t.Table.Name.String(),
		}
		ns = []*node{n}
	case *ast.DropTableStmt:
		for _, table := range t.Tables {
			n := &node{
				db:    table.Schema.String(),
				table: table.Name.String(),
			}
			ns = append(ns, n)
		}
	case *ast.CreateTableStmt:
		n := &node{
			db:    t.Table.Schema.String(),
			table: t.Table.Name.String(),
		}
		ns = []*node{n}
	case *ast.TruncateTableStmt:
		n := &node{
			db:    t.Table.Schema.String(),
			table: t.Table.Schema.String(),
		}
		ns = []*node{n}
	}
	return
}

func (rb *RunningBinloger) updateTable(db, table string) (err error) {
	rb.clearTableCache([]byte(db), []byte(table))
	moduleLogger.Warnf("table structure changed, clear table cache: %s.%s\n", db, table)
	if err = rb.eventHandler.OnTableChanged(db, table); err != nil && errors.Cause(err) != schema.ErrTableNotExist {
		return err
	}
	return
}
func (rb *RunningBinloger) updateReplicationDelay(ev *replication.BinlogEvent) {
	var newDelay uint32
	now := uint32(time.Now().Unix())
	if now >= ev.Header.Timestamp {
		newDelay = now - ev.Header.Timestamp
	}
	atomic.StoreUint32(rb.delay, newDelay)
}

func (b *RunningBinloger) handleRowsEvent(e *replication.BinlogEvent) error {
	ev := e.Event.(*replication.RowsEvent)

	// Caveat: table may be altered at runtime.
	schema := string(ev.Table.Schema)
	table := string(ev.Table.Table)

	t, target, err := b.getTable(schema, table)
	if err != nil {
		if err == ErrExcludedTable {
			return nil
		}
		return err
	}

	moduleLogger.Debugf("get table info ok, %s.%s", schema, table)

	var action string
	switch e.Header.EventType {
	case replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
		action = InsertAction
	case replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
		action = DeleteAction
	case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
		action = UpdateAction
	default:
		return errors.Errorf("%s not supported now", e.Header.EventType)
	}
	events := newRowsEvent(t, action, ev.Rows, e.Header)
	events.DBCfg = target
	return b.eventHandler.OnRow(events)
}

func (b *RunningBinloger) FlushBinlog() error {
	_, err := b.Execute("FLUSH BINARY LOGS")
	return errors.Trace(err)
}

func (rb *RunningBinloger) WaitUntilPos(pos mysql.Position, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	for {
		select {
		case <-timer.C:
			return fmt.Errorf("wait position %v too long > %s", pos, timeout)
		default:
			err := rb.FlushBinlog()
			if err != nil {
				return err
			}
			curPos := rb.master.Position()
			if curPos.Compare(pos) >= 0 {
				return nil
			} else {
				moduleLogger.Debugf("master pos is %v, wait catching %v", curPos, pos)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (rb *RunningBinloger) GetMasterGTIDSet() (mysql.GTIDSet, error) {
	query := ""
	switch rb.cfg.Flavor {
	case mysql.MariaDBFlavor:
		query = "SELECT @@GLOBAL.gtid_current_pos"
	default:
		query = "SELECT @@GLOBAL.GTID_EXECUTED"
	}
	rr, err := rb.Execute(query)
	if err != nil {
		return nil, err
	}
	gx, err := rr.GetString(0, 0)
	if err != nil {
		return nil, err
	}
	gset, err := mysql.ParseGTIDSet(rb.cfg.Flavor, gx)
	if err != nil {
		return nil, err
	}
	return gset, nil
}

func (rb *RunningBinloger) CatchMasterPos(timeout time.Duration) error {
	pos, err := rb.GetMasterPos()
	if err != nil {
		return errors.Trace(err)
	}

	return rb.WaitUntilPos(pos, timeout)
}
