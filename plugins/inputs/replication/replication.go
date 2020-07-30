package replication

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"database/sql"
	"github.com/jackc/pgx"
	_ "github.com/lib/pq"
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
# [[inputs.replication]]
#	# required
# 	host="127.0.0.1"
#
#	# required
# 	port=25432
#
# 	# postgres user (need replication privilege)
#	# required
# 	user="testuser"
#
#	# required
# 	password="pwd"
#
#	# required
# 	database="testdb"
#
# 	# exlcude the events of postgres
# 	# there are 3 events: "INSERT","UPDATE","DELETE"
#	# required
# 	events=["INSERT"]
#
# 	# tags
# 	tagList=[]
#
# 	# fields. required
# 	fieldList=["fieldName"]
#
# 	# [inputs.replication.tags]
# 	# tags1 = "value1"
`
)

var (
	l *logger.Logger

	testAssert bool
)

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
	Events    []string          `toml:"events"`
	TagList   []string          `toml:"tagList"`
	FieldList []string          `toml:"fieldList"`
	Tags      map[string]string `toml:"tags"`

	// slice to map cache
	eventsOperation map[string]interface{}
	tagKeys         map[string]interface{}
	fieldKeys       map[string]interface{}

	slotName string
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

	r.slotName = fmt.Sprintf("datakit_slot_%d", time.Now().UnixNano())
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

	l.Infof("postgresql replication input started.")

	r.runloop()
}

func (r *Replication) runloop() {

	tick := time.NewTicker(time.Second * 5)

	for {
		select {
		case <-datakit.Exit.Wait():
			if err := r.replicationConn.Close(); err != nil {
				l.Errorf("replication connection close err: %s", err.Error())
			}
			if err := r.deleteSelfSlot(); err != nil {
				l.Errorf("delete self slot err: %s", err.Error())
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

			if err := r.replicationMsgHandle(msg); err != nil {
				l.Error(err)
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

	if _, _, err := conn.CreateReplicationSlotEx(r.slotName, "test_decoding"); err != nil {
		return fmt.Errorf("failed to create replication slot: %s", err)
	}

	if err := conn.StartReplication(r.slotName, 0, -1); err != nil {
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
	if msg == nil {
		return fmt.Errorf("msg is empty")
	}

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

		// filter other events
		if testAssert && len(data) > 0 {
			l.Debugf("Data: %s", string(data))
			return nil
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
	tags["relation"] = result.Relation

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

func (r *Replication) deleteSelfSlot() error {

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", r.User, r.Password, r.Host, r.Port, r.Database)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	sqlstr := fmt.Sprintf("SELECT pg_drop_replication_slot('%s')", r.slotName)
	_, err = db.Exec(sqlstr)
	if err != nil {
		return err
	}

	return nil
}
