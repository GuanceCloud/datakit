package binlog

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/siddontang/go-log/log"
	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"github.com/siddontang/go-mysql/schema"
	"gitlab.jiagouyun.com/cloudcare-tools/ftcollector/config"
	"gitlab.jiagouyun.com/cloudcare-tools/ftcollector/uploader"
)

type Binloger struct {
	m sync.Mutex

	cfg *config.BinlogDatasource

	parser *parser.Parser
	master *masterInfo

	syncer *replication.BinlogSyncer

	eventHandler EventHandler

	connLock sync.Mutex
	conn     *client.Conn

	tableLock          sync.RWMutex
	tables             map[string]*schema.Table
	errorTablesGetTime map[string]time.Time

	tableMatchCache map[string]bool
	// includeTableRegex []*regexp.Regexp
	// excludeTableRegex []*regexp.Regexp

	delay *uint32

	ctx    context.Context
	cancel context.CancelFunc

	storage *uploader.Uploader
}

var (
	UnknownTableRetryPeriod = time.Second * time.Duration(10)
	ErrExcludedTable        = errors.New("excluded table meta")

	HeartbeatPeriod = 60 * time.Second
	ReadTimeout     = 90 * time.Second
)

func Start(cfg *config.BinlogConfig) error {

	if cfg == nil || cfg.Disable {
		return nil
	}

	var wg sync.WaitGroup

	var err error

	for _, dt := range cfg.Datasources {
		dt.ServerID = uint32(rand.New(rand.NewSource(time.Now().Unix())).Intn(1000)) + 1001
		dt.Charset = mysql.DEFAULT_CHARSET
		dt.Flavor = mysql.MySQLFlavor
		dt.HeartbeatPeriod = HeartbeatPeriod
		dt.DiscardNoMetaRowEvent = true
		dt.ReadTimeout = ReadTimeout
		dt.UseDecimal = true
		dt.ParseTime = true
		dt.SemiSyncEnabled = false

		binloger, err := NewBinloger(dt)
		if err != nil {
			log.Fatalf("%s", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			binloger.run()
		}()
	}

	wg.Wait()

	return err
}

func (c *Binloger) run() error {
	defer func() {
		c.cancel()
	}()

	up := uploader.New(c.cfg.FTGateway)
	up.Start()

	c.storage = up

	defer func() {
		up.Stop()
	}()

	c.master.UpdateTimestamp(uint32(time.Now().Unix()))

	if err := c.runSyncBinlog(); err != nil {
		if errors.Cause(err) != context.Canceled {
			log.Errorf("start sync binlog err: %v", err)
			return err
		}
	}

	return nil
}

func (c *Binloger) Close() {
	c.m.Lock()
	defer c.m.Unlock()

	c.cancel()
	c.syncer.Close()
	c.connLock.Lock()
	c.conn.Close()
	c.conn = nil
	c.connLock.Unlock()

	c.eventHandler.OnPosSynced(c.master.Position(), c.master.GTIDSet(), true)
}

func NewBinloger(cfg *config.BinlogDatasource) (*Binloger, error) {
	c := &Binloger{}
	c.cfg = cfg

	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.eventHandler = &MainEventHandler{
		binloger: c,
	}
	c.parser = parser.New()
	c.tables = make(map[string]*schema.Table)

	if c.cfg.DiscardNoMetaRowEvent {
		c.errorTablesGetTime = make(map[string]time.Time)
	}
	c.master = &masterInfo{}

	c.delay = new(uint32)

	var err error

	if err = c.prepareSyncer(); err != nil {
		return nil, errors.Trace(err)
	}

	if err = c.checkMysqlVersion(); err != nil {
		return nil, errors.Trace(err)
	}

	if err := c.checkBinlogRowFormat(); err != nil {
		return nil, errors.Trace(err)
	}

	if err := c.CheckBinlogRowImage("FULL"); err != nil {
		return nil, errors.Trace(err)
	}

	// if c.includeTableRegex != nil || c.excludeTableRegex != nil {
	c.tableMatchCache = make(map[string]bool)
	// }

	return c, nil
}

func (c *Binloger) prepareSyncer() error {
	cfg := replication.BinlogSyncerConfig{
		ServerID:                c.cfg.ServerID,
		Flavor:                  "mysql",
		User:                    c.cfg.User,
		Password:                c.cfg.Password,
		Charset:                 c.cfg.Charset,
		HeartbeatPeriod:         c.cfg.HeartbeatPeriod,
		ReadTimeout:             c.cfg.ReadTimeout,
		UseDecimal:              c.cfg.UseDecimal,
		ParseTime:               c.cfg.ParseTime,
		SemiSyncEnabled:         c.cfg.SemiSyncEnabled,
		MaxReconnectAttempts:    c.cfg.MaxReconnectAttempts,
		TimestampStringLocation: c.cfg.TimestampStringLocation,
	}

	if strings.Contains(c.cfg.Addr, "/") {
		cfg.Host = c.cfg.Addr
	} else {
		seps := strings.Split(c.cfg.Addr, ":")
		if len(seps) != 2 {
			return errors.Errorf("invalid mysql addr format %s, must host:port", c.cfg.Addr)
		}

		port, err := strconv.ParseUint(seps[1], 10, 16)
		if err != nil {
			return errors.Trace(err)
		}

		cfg.Host = seps[0]
		cfg.Port = uint16(port)
	}

	c.syncer = replication.NewBinlogSyncer(cfg)

	return nil
}

func (c *Binloger) GetTable(db string, table string) (*schema.Table, *config.BinlogInput, error) {
	key := fmt.Sprintf("%s.%s", db, table)
	// if table is excluded, return error and skip parsing event or dump
	target := c.checkTableMatch(db, table)
	if target == nil {
		return nil, nil, ErrExcludedTable
	}

	c.tableLock.RLock()
	t, ok := c.tables[key]
	c.tableLock.RUnlock()

	if ok {
		return t, target, nil
	}

	if c.cfg.DiscardNoMetaRowEvent {
		c.tableLock.RLock()
		lastTime, ok := c.errorTablesGetTime[key]
		c.tableLock.RUnlock()
		if ok && time.Now().Sub(lastTime) < UnknownTableRetryPeriod {
			return nil, nil, schema.ErrMissingTableMeta
		}
	}

	t, err := schema.NewTable(c, db, table)
	if err != nil {
		// check table not exists
		if ok, err1 := schema.IsTableExist(c, db, table); err1 == nil && !ok {
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
			c.tableLock.Lock()
			c.tables[key] = ta
			c.tableLock.Unlock()
			return ta, target, nil
		}
		// if DiscardNoMetaRowEvent is true, we just log this error
		if c.cfg.DiscardNoMetaRowEvent {
			c.tableLock.Lock()
			c.errorTablesGetTime[key] = time.Now()
			c.tableLock.Unlock()
			// log error and return ErrMissingTableMeta
			log.Errorf("get table meta err: %v", errors.Trace(err))
			return nil, nil, schema.ErrMissingTableMeta
		}
		return nil, nil, err
	}

	c.tableLock.Lock()
	c.tables[key] = t
	if c.cfg.DiscardNoMetaRowEvent {
		// if get table info success, delete this key from errorTablesGetTime
		delete(c.errorTablesGetTime, key)
	}
	c.tableLock.Unlock()

	return t, target, nil
}

// ClearTableCache clear table cache
func (c *Binloger) ClearTableCache(db []byte, table []byte) {
	key := fmt.Sprintf("%s.%s", db, table)
	c.tableLock.Lock()
	delete(c.tables, key)
	if c.cfg.DiscardNoMetaRowEvent {
		delete(c.errorTablesGetTime, key)
	}
	c.tableLock.Unlock()
}

// CheckBinlogRowImage checks MySQL binlog row image, must be in FULL, MINIMAL, NOBLOB
func (c *Binloger) CheckBinlogRowImage(image string) error {
	// need to check MySQL binlog row image? full, minimal or noblob?
	// now only log
	if c.cfg.Flavor == mysql.MySQLFlavor {
		if res, err := c.Execute(`SHOW GLOBAL VARIABLES LIKE "binlog_row_image"`); err != nil {
			return errors.Trace(err)
		} else {
			// MySQL has binlog row image from 5.6, so older will return empty
			rowImage, _ := res.GetString(0, 1)
			if rowImage != "" && !strings.EqualFold(rowImage, image) {
				return errors.Errorf("MySQL uses %s binlog row image, but we want %s", rowImage, image)
			}
		}
	}

	return nil
}

func (c *Binloger) checkBinlogRowFormat() error {
	res, err := c.Execute(`SHOW GLOBAL VARIABLES LIKE "binlog_format";`)
	if err != nil {
		return errors.Trace(err)
	} else if f, _ := res.GetString(0, 1); f != "ROW" {
		return errors.Errorf("binlog must ROW format, but %s now", f)
	}

	return nil
}

func (c *Binloger) getMasterStatus(m *masterInfo) error {
	res, err := c.Execute(`show master status;`)
	if err != nil {
		return errors.Trace(err)
	}

	filename, err := res.GetString(0, 0)
	if err != nil {
		return errors.Trace(err)
	}
	pos, err := res.GetUint(0, 1)
	if err != nil {
		return errors.Trace(err)
	}

	m.Update(mysql.Position{
		Name: filename,
		Pos:  uint32(pos),
	})

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

func (c *Binloger) checkMysqlVersion() error {
	es, err := c.Execute(`SELECT version();`)
	if err != nil {
		return errors.Trace(err)
	}

	ver, _ := es.GetString(0, 0)
	log.Infof("server version: %s", ver)
	if strings.Contains(strings.ToLower(ver), "maria") {
		c.cfg.Flavor = "mariadb"
	}

	return nil
}

// Execute a SQL
func (c *Binloger) Execute(cmd string, args ...interface{}) (rr *mysql.Result, err error) {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	retryNum := 3
	for i := 0; i < retryNum; i++ {
		if c.conn == nil {
			c.conn, err = client.Connect(c.cfg.Addr, c.cfg.User, c.cfg.Password, "")
			if err != nil {
				return nil, errors.Trace(err)
			}
		}

		rr, err = c.conn.Execute(cmd, args...)
		if err != nil && !mysql.ErrorEqual(err, mysql.ErrBadConn) {
			return
		} else if mysql.ErrorEqual(err, mysql.ErrBadConn) {
			c.conn.Close()
			c.conn = nil
			continue
		} else {
			return
		}
	}
	return
}

func (c *Binloger) checkTableMatch(db, table string) *config.BinlogInput {

	var destdb *config.BinlogInput
	for _, t := range c.cfg.Inputs {
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
			if tbl.Table == table {
				bmatch = true
				break
			}
		}
	} else {
		bmatch = true
	}

	if len(destdb.ExcludeTables) > 0 {
		for _, tblname := range destdb.ExcludeTables {
			if table == tblname {
				bmatch = false
				break
			}
		}
	}

	if !bmatch {
		return nil
	}

	return destdb
}
