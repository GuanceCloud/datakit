package replication

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/jackc/pgx"
	"github.com/nickelser/parselogical"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "replication"

	defaultMeasurement = "replication"

	sampleCfg = `
# [inputs.replication]
# 	# postgres host
# 	host="127.0.0.1"
#
# 	# postgres port
# 	port=25432
#
# 	# postgres user (need replication privilege)
# 	user="testuser"
#
# 	password="pwd"
#
# 	database="testdb"
#
# 	table="testable"
#
# 	# replication slot name, only
# 	slotname="slot_for_datakit"
#
# 	# exlcude the events of postgres
# 	# there are 3 events: "insert","update","delete"
# 	events=['insert']
#
# 	# tags
# 	tagList=['colunm1']
#
# 	# fields
# 	fieldList=['colunm0']
#
# 	# [inputs.replication.tags]
# 	# tags1 = "tags1"
`
)

var l *logger.Logger

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Replication{}
	})
}

type Replication struct {
	Host      string            `toml:"host"`
	Port      uint16            `toml:"port"`
	Database  string            `toml:"database"`
	User      string            `toml:"user"`
	Password  string            `toml:"password"`
	Table     string            `toml:"table"`
	SlotName  string            `toml:"slotname"`
	Events    []string          `toml:"events"`
	TagList   []string          `toml:"tagList"`
	FieldList []string          `toml:"fieldList"`
	Tags      map[string]string `toml:"tags"`

	// slice to map cache
	eventsOperation map[string]interface{}
	tagKeys         map[string]interface{}
	fieldKeys       map[string]interface{}

	// 当前 wal 位置
	receivedWal uint64
	flushWal    uint64
	// 复制连接
	replicationConn *pgx.ReplicationConn
	// pgConfig
	pgConfig pgx.ConnConfig
	// ack 锁
	sendStatusLock sync.Mutex
}

func (_ *Replication) Catalog() string {
	return "postgresql"
}

func (_ *Replication) SampleConfig() string {
	return sampleCfg
}

func (r *Replication) Run() {
	l = logger.SLogger(inputName)

	if r.Tags == nil {
		r.Tags = make(map[string]string)
	}
	r.updateParamList()

	r.pgConfig = pgx.ConnConfig{
		Host:     r.Host,
		Port:     r.Port,
		Database: r.Database,
		User:     r.User,
		Password: r.Password,
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		default:
			// nil
		}

		if err := r.checkAndResetConn(); err != nil {
			l.Errorf("failed to connect, err: %s", err.Error())
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	_ = r.sendStatus()

	r.runloop()
}

func (r *Replication) runloop() {

	tick := time.NewTicker(time.Second * 5)

	for {
		select {
		case <-datakit.Exit.Wait():
			if err := r.replicationConn.Close(); err != nil {
				l.Error(err)
			}
			l.Info("exit")
			return

		case <-tick.C:
			_ = r.sendStatus()

		default:
			msg, err := r.replicationConn.WaitForReplicationMessage(context.Background())
			if err != nil {
				l.Error(err)
				if err := r.checkAndResetConn(); err != nil {
					l.Error(err)
					time.Sleep(time.Second)
				}
				continue
			}

			if msg == nil {
				l.Error(err)
				continue
			}

			if err := r.replicationMsgHandle(msg); err != nil {
				l.Error(err)
				continue
			}
		}
	}
}

func (r *Replication) updateParamList() {
	r.eventsOperation = make(map[string]interface{})
	r.tagKeys = make(map[string]interface{})
	r.fieldKeys = make(map[string]interface{})

	for _, event := range r.Events {
		r.eventsOperation[event] = nil
	}
	for _, tag := range r.TagList {
		r.tagKeys[tag] = nil
	}
	for _, field := range r.FieldList {
		r.fieldKeys[field] = nil
	}
}

func (r *Replication) checkAndResetConn() error {

	if r.replicationConn != nil && r.replicationConn.IsAlive() {
		return nil
	}

	conn, err := pgx.ReplicationConnect(r.pgConfig)
	if err != nil {
		return err
	}

	if _, _, err := conn.CreateReplicationSlotEx(r.SlotName, "test_decoding"); err != nil {
		if pgerr, ok := err.(pgx.PgError); !ok || pgerr.Code != "42710" {
			return fmt.Errorf("failed to create replication slot: %s", err)
		}
	}

	if err := conn.StartReplication(r.SlotName, 0, -1); err != nil {
		_ = conn.Close()
		return err
	}

	r.replicationConn = conn
	return nil
}

func (r *Replication) getReceivedWal() uint64 {
	return atomic.LoadUint64(&r.receivedWal)
}

func (r *Replication) setReceivedWal(val uint64) {
	atomic.StoreUint64(&r.receivedWal, val)
}

func (r *Replication) getFlushWal() uint64 {
	return atomic.LoadUint64(&r.flushWal)
}

func (r *Replication) getStatus() (*pgx.StandbyStatus, error) {
	return pgx.NewStandbyStatus(r.getReceivedWal(), r.getFlushWal(), r.getFlushWal())
}

// ReplicationMsgHandle handle replication msg
func (r *Replication) replicationMsgHandle(msg *pgx.ReplicationMessage) error {

	// 回复心跳
	if msg.ServerHeartbeat != nil {

		if msg.ServerHeartbeat.ServerWalEnd > r.getReceivedWal() {
			r.setReceivedWal(msg.ServerHeartbeat.ServerWalEnd)
		}
		if msg.ServerHeartbeat.ReplyRequested == 1 {
			_ = r.sendStatus()
		}
	}

	if msg.WalMessage != nil {

		data, err := r.parseMsg(msg.WalMessage)
		if err != nil {
			return fmt.Errorf("invalid pgoutput msg: %s", err)
		}
		return io.NamedFeed(data, io.Metric, inputName)
	}

	return nil
}

// sendStatus send heartbeat
func (r *Replication) sendStatus() error {
	r.sendStatusLock.Lock()
	defer r.sendStatusLock.Unlock()

	status, err := r.getStatus()
	if err != nil {
		return err
	}

	return r.replicationConn.SendStandbyStatus(status)
}

// Parse test_decoding format wal to WalData
func (r *Replication) parseMsg(msg *pgx.WalMessage) ([]byte, error) {
	result := parselogical.NewParseResult(*(*string)(unsafe.Pointer(&msg.WalData)))
	if err := result.Parse(); err != nil {
		return nil, err
	}

	if _, ok := r.eventsOperation[result.Operation]; !ok {
		// INSERT, UPDATE, DELETE, BEGIN, COMMIT
		return nil, nil
	}

	var tags = make(map[string]string, len(r.tagKeys))
	var fields = make(map[string]interface{}, len(r.fieldKeys))

	for key, column := range result.Columns {
		if _, ok := r.tagKeys[key]; ok {
			tags[key] = column.Value
		}

		if _, ok := r.fieldKeys[key]; ok {
			fields[key] = typeAssert(column.Type, column.Value)
		}
	}

	for k, v := range r.Tags {
		tags[k] = v
	}

	return io.MakeMetric(defaultMeasurement, tags, fields, time.Now())
}

func typeAssert(typer, value string) interface{} {

	switch typer {
	case "smallint", "integer", "bigint", "real", "double precision", "smallserial", "serial", "bigserial", "money":
		if val, err := strconv.ParseInt(value, 10, 64); err == nil {
			return val
		}
	case "decimal", "numeric":
		if val, err := strconv.ParseFloat(value, 64); err == nil {
			return val
		}
	case "boolean":
		// FIXME: pg boolean has 'true','false','unknow'
		if val, err := strconv.ParseBool(value); err == nil {
			return val
		}
	default:
		return value
	}

	return nil
}
