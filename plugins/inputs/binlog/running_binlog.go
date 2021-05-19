// +build !386,!arm

package binlog

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/pingcap/tidb/types/parser_driver"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser"

	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"github.com/siddontang/go-mysql/schema"
)

var (
	cachePosDir = `/etc/datakit/binlog`
)

func NewRunningBinloger(cfg *InstanceConfig) *RunningBinloger {
	b := &RunningBinloger{
		cfg:    cfg,
		parser: parser.New(),
		delay:  new(uint32),
		master: &masterInfo{},
	}

	b.eventHandler = &MainEventHandler{
		rb: b,
	}

	return b
}

type RunningBinloger struct {
	//m sync.Mutex

	binlog *Binlog

	cfg *InstanceConfig

	parser *parser.Parser

	master *masterInfo

	syncer *replication.BinlogSyncer

	eventHandler EventHandler

	connLock sync.Mutex
	conn     *client.Conn

	tableLock          sync.RWMutex
	tables             map[string]*schema.Table
	errorTablesGetTime map[string]time.Time

	delay *uint32

	//ctx context.Context
}

func (rb *RunningBinloger) run(ctx context.Context) error {

	if err := recover(); err != nil {
		moduleLogger.Errorf("panic, %s", err)
	}

	rb.tables = make(map[string]*schema.Table)
	if rb.cfg.DiscardNoMetaRowEvent {
		rb.errorTablesGetTime = make(map[string]time.Time)
	}

	var err error

	if err = rb.prepareSyncer(); err != nil {
		moduleLogger.Errorf("prepareSyncer error, %s", err)
		return err
	}

	if err = rb.checkMysqlVersion(); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	if err := rb.checkBinlogRowFormat(); err != nil {
		return err
	}

	if err := rb.checkBinlogRowImage("FULL"); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	rb.master.UpdateTimestamp(uint32(time.Now().Unix()))
	if err = rb.getMasterStatus(rb.master); err != nil {
		moduleLogger.Errorf("check if mysql binlog enabled?")
		return err
	}

	moduleLogger.Infof("check requirments ok")

	if err := rb.doSync(ctx); err != nil {
		if err != context.Canceled && errors.Cause(err) != context.Canceled {
			moduleLogger.Errorf("sync err: %s", err)
			return err
		}
	}

	return nil
}

func (rb *RunningBinloger) stop() {
	if rb.syncer != nil {
		rb.syncer.Close()
	}
}

func (rb *RunningBinloger) getMasterStatus(m *masterInfo) error {
	res, err := rb.Execute(`show master status;`)
	if err != nil {
		return err
	}

	filename, err := res.GetString(0, 0)
	if err != nil {
		return err
	}
	pos, err := res.GetUint(0, 1)
	if err != nil {
		return err
	}

	m.Update(mysql.Position{
		Name: filename,
		Pos:  uint32(pos),
	})

	if savedpos, err := rb.loadMasterStatus(rb.cfg.Addr); err == nil && savedpos != nil {
		m.Update(mysql.Position{
			Name: savedpos.Name,
			Pos:  savedpos.Pos,
		})
	}

	dodb, err := res.GetString(0, 2)
	if err == nil {
		m.binlogDoDB = dodb
	}

	ignoredb, err := res.GetString(0, 3)
	if err == nil {
		m.binlogIngoreDB = ignoredb
	}

	return nil
}

func (rb *RunningBinloger) Execute(cmd string, args ...interface{}) (rr *mysql.Result, err error) {
	rb.connLock.Lock()
	defer rb.connLock.Unlock()

	retryNum := 3
	for i := 0; i < retryNum; i++ {
		if rb.conn == nil {
			rb.conn, err = client.Connect(rb.cfg.Addr, rb.cfg.User, rb.cfg.Password, "")
			if err != nil {
				return nil, fmt.Errorf("fail to connect mysql(%s/%s), %s", rb.cfg.Addr, rb.cfg.User, err)
			}
		}

		rr, err = rb.conn.Execute(cmd, args...)
		if err != nil && !mysql.ErrorEqual(err, mysql.ErrBadConn) {
			return
		} else if mysql.ErrorEqual(err, mysql.ErrBadConn) {
			rb.conn.Close()
			rb.conn = nil
			continue
		} else {
			return
		}
	}
	return
}

func (rb *RunningBinloger) checkBinlogRowImage(image string) error {
	// need to check MySQL binlog row image? full, minimal or noblob?
	// now only log
	if rb.cfg.Flavor == mysql.MySQLFlavor {
		if res, err := rb.Execute(`SHOW GLOBAL VARIABLES LIKE "binlog_row_image"`); err != nil {
			moduleLogger.Errorf("fail to get binlog_row_image, %s", err)
			return err
		} else {
			// MySQL has binlog row image from 5.6, so older will return empty
			rowImage, _ := res.GetString(0, 1)
			if rowImage != "" && !strings.EqualFold(rowImage, image) {
				return fmt.Errorf("MySQL uses %s binlog row image, but we want %s", rowImage, image)
			}
		}
	}

	return nil
}

func (rb *RunningBinloger) checkBinlogRowFormat() error {
	res, err := rb.Execute(`SHOW GLOBAL VARIABLES LIKE "binlog_format";`)
	if err != nil {
		return err
	}

	rowFormat, err := res.GetString(0, 1)
	if err != nil {
		return err
	}

	moduleLogger.Debugf("row format: %s", rowFormat)

	if rowFormat != "ROW" {
		moduleLogger.Errorf("binlog_row_format should be ROW")
		return fmt.Errorf("binlog_row_format should be ROW")
	}

	return nil
}

func (rb *RunningBinloger) GetMasterPos() (mysql.Position, error) {
	rr, err := rb.Execute("SHOW MASTER STATUS")
	if err != nil {
		return mysql.Position{}, err
	}

	name, _ := rr.GetString(0, 0)
	pos, _ := rr.GetInt(0, 1)

	return mysql.Position{Name: name, Pos: uint32(pos)}, nil
}

func (rb *RunningBinloger) checkMysqlVersion() error {
	es, err := rb.Execute(`SELECT version();`)
	if err != nil {
		moduleLogger.Errorf("fail to execute 'SELECT version();', %s", err)
		return err
	}

	ver, _ := es.GetString(0, 0)
	moduleLogger.Debugf("mysql server version: %s", ver)
	if strings.Contains(strings.ToLower(ver), "maria") {
		rb.cfg.Flavor = "mariadb"
	}

	return nil
}

func (rb *RunningBinloger) saveMasterStatus(m *masterInfo) error {

	if m == nil {
		return nil
	}

	if err := os.MkdirAll(cachePosDir, 0755); err != nil {
		return err
	}

	data, err := json.Marshal(&m.pos)
	if err != nil {
		return err
	}

	h := md5.New()
	h.Write([]byte(rb.cfg.Addr))
	k := hex.EncodeToString(h.Sum(nil))

	path := filepath.Join(cachePosDir, k)
	ioutil.WriteFile(path, data, 0644)

	return nil
}

func (rb *RunningBinloger) loadMasterStatus(addr string) (*mysql.Position, error) {

	h := md5.New()
	h.Write([]byte(addr))
	k := hex.EncodeToString(h.Sum(nil))

	path := filepath.Join(cachePosDir, k)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p mysql.Position
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (b *RunningBinloger) prepareSyncer() error {
	syncCfg := replication.BinlogSyncerConfig{
		ServerID:                b.cfg.ServerID,
		Flavor:                  "mysql",
		User:                    b.cfg.User,
		Password:                b.cfg.Password,
		Charset:                 b.cfg.Charset,
		HeartbeatPeriod:         b.cfg.HeartbeatPeriod,
		ReadTimeout:             b.cfg.ReadTimeout,
		UseDecimal:              b.cfg.UseDecimal,
		ParseTime:               b.cfg.ParseTime,
		SemiSyncEnabled:         b.cfg.SemiSyncEnabled,
		MaxReconnectAttempts:    b.cfg.MaxReconnectAttempts,
		TimestampStringLocation: b.cfg.TimestampStringLocation,
	}

	if strings.Contains(b.cfg.Addr, "/") {
		syncCfg.Host = b.cfg.Addr
	} else {
		seps := strings.Split(b.cfg.Addr, ":")
		if len(seps) != 2 {
			return fmt.Errorf("invalid mysql addr format %s, must host:port", b.cfg.Addr)
		}

		port, err := strconv.ParseUint(seps[1], 10, 16)
		if err != nil {
			return err
		}

		syncCfg.Host = seps[0]
		syncCfg.Port = uint16(port)
	}

	b.syncer = replication.NewBinlogSyncer(syncCfg)

	return nil
}

func (rb *RunningBinloger) getTable(db string, table string) (*schema.Table, *DatabaseConfig, error) {
	key := fmt.Sprintf("%s.%s", db, table)
	// if table is excluded, return error and skip parsing event or dump
	//fmt.Println(key)
	target := rb.checkTableMatch(db, table)
	if target == nil {
		//rb.binlog.logger.Warnf("table %s.%s not in configuration, ignored", db, table)
		return nil, nil, ErrExcludedTable
	}

	rb.tableLock.RLock()
	t, ok := rb.tables[key]
	rb.tableLock.RUnlock()

	if ok {
		return t, target, nil
	}

	if rb.cfg.DiscardNoMetaRowEvent {
		rb.tableLock.RLock()
		lastTime, ok := rb.errorTablesGetTime[key]
		rb.tableLock.RUnlock()
		if ok && time.Now().Sub(lastTime) < UnknownTableRetryPeriod {
			return nil, nil, schema.ErrMissingTableMeta
		}
	}

	t, err := schema.NewTable(rb, db, table)
	if err != nil {
		// check table not exists
		if ok, err1 := schema.IsTableExist(rb, db, table); err1 == nil && !ok {
			moduleLogger.Errorf("table %s.%s not exists, %s", db, table, err1)
			return nil, nil, schema.ErrTableNotExist
		}
		// work around : RDS HAHeartBeat
		// ref : https://github.com/alibaba/canal/blob/master/parse/src/main/java/com/alibaba/otter/canal/parse/inbound/mysql/dbsync/LogEventConvert.java#L385
		// issue : https://github.com/alibaba/canal/issues/222
		// This is a common error in RDS that canal can't get HAHealthCheckSchema's meta, so we mock a table meta.
		// If canal just skip and log error, as RDS HA heartbeat interval is very short, so too many HAHeartBeat errors will be logged.
		if key == schema.HAHealthCheckSchema {
			// mock ha_health_check meta
			ta := &schema.Table{
				Schema:  db,
				Name:    table,
				Columns: make([]schema.TableColumn, 0, 2),
				Indexes: make([]*schema.Index, 0),
			}
			ta.AddColumn("id", "bigint(20)", "", "")
			ta.AddColumn("type", "char(1)", "", "")
			rb.tableLock.Lock()
			rb.tables[key] = ta
			rb.tableLock.Unlock()
			return ta, target, nil
		}
		// if DiscardNoMetaRowEvent is true, we just log this error
		if rb.cfg.DiscardNoMetaRowEvent {
			rb.tableLock.Lock()
			rb.errorTablesGetTime[key] = time.Now()
			rb.tableLock.Unlock()
			// log error and return ErrMissingTableMeta
			return nil, nil, schema.ErrMissingTableMeta
		}
		moduleLogger.Errorf("get table meta failed: %s", err)
		return nil, nil, err
	} else {
		cols := []string{}
		for _, cl := range t.Columns {
			cols = append(cols, fmt.Sprintf("%s(%s)", cl.Name, cl.RawType))
		}
		moduleLogger.Debugf("table info %s.%s: %s", db, table, strings.Join(cols, " - "))
	}

	rb.tableLock.Lock()
	rb.tables[key] = t
	if rb.cfg.DiscardNoMetaRowEvent {
		// if get table info success, delete this key from errorTablesGetTime
		delete(rb.errorTablesGetTime, key)
	}
	rb.tableLock.Unlock()

	return t, target, nil
}

func (rb *RunningBinloger) checkTableMatch(db, table string) *DatabaseConfig {

	var destdb *DatabaseConfig
	for _, t := range rb.cfg.Databases {
		if t.Database == db {
			destdb = t
			break
		}
	}

	if destdb == nil {
		return nil
	}

	bmatch := false

	if len(destdb.Tables) > 0 {
		for _, tbl := range destdb.Tables {
			//fmt.Printf("tbname: %s\n", tbl.Name)
			if tbl.Name == table {
				bmatch = true
				break
			}
		}
	} else {
		bmatch = true
	}

	// if len(destdb.ExcludeTables) > 0 {
	// 	for _, tblname := range destdb.ExcludeTables {
	// 		if table == tblname {
	// 			bmatch = false
	// 			break
	// 		}
	// 	}
	// }

	if !bmatch {
		return nil
	}

	return destdb
}

func (rb *RunningBinloger) clearTableCache(db []byte, table []byte) {
	key := fmt.Sprintf("%s.%s", db, table)
	rb.tableLock.Lock()
	delete(rb.tables, key)
	if rb.cfg.DiscardNoMetaRowEvent {
		delete(rb.errorTablesGetTime, key)
	}
	rb.tableLock.Unlock()
}
