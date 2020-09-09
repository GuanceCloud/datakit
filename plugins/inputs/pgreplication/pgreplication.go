package pgreplication

import (
	"context"
	"database/sql"
	"fmt"
	ioeof "io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx"
	_ "github.com/lib/pq"
	"github.com/nickelser/parselogical"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "postgresql_replication"

	defaultMeasurement = "postgresql_replication"

	sampleCfg = `
[[inputs.postgresql_replication]]
    # required
    host="127.0.0.1"

    # required
    port=25432

    # postgres user (need replication privilege)
    # required
    user="testuser"

    # required
    password="pwd"

    # required
    database="testdb"

    table="test_table"

    # there are 3 events: "INSERT","UPDATE","DELETE"
    # required
    events=["INSERT"]

    # tags
    tag_colunms=[]

    # fields. required
    field_colunms=["fieldName"]

    # [inputs.postgresql_replication.tags]
    # tags1 = "value1"
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Replication{}
	})
}

type Replication struct {
	Host      string            `toml:"host"`
	Port      uint16            `toml:"port"`
	User      string            `toml:"user"`
	Password  string            `toml:"password"`
	Database  string            `toml:"database"`
	Table     string            `toml:"table"`
	Events    []string          `toml:"events"`
	TagList   []string          `toml:"tag_colunms"`
	FieldList []string          `toml:"field_colunms"`
	Tags      map[string]string `toml:"tags"`

	// slice to map cache
	eventsOperation map[string]interface{}
	tagKeys         map[string]interface{}
	fieldKeys       map[string]interface{}

	slotName string
	// 当前 wal 位置
	receivedWal uint64
	// 复制连接
	replicationConn *pgx.ReplicationConn
	// pgConfig
	pgConfig pgx.ConnConfig
	// ack 锁
	sendStatusLock sync.Mutex
}

func (*Replication) Catalog() string {
	return "db"
}

func (*Replication) SampleConfig() string {
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

	l.Infof("postgresql replication input started.")

	r.runloop()
}

const sendHeartbeatInterval = 5

func (r *Replication) runloop() {
	tick := time.NewTicker(time.Second * sendHeartbeatInterval)

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
			if err := r.sendStatus(); err != nil {
				l.Error(err)
			}

		default:
			msg, err := r.replicationConn.WaitForReplicationMessage(context.Background())
			if err != nil {
				// filter useless log info
				if err != ioeof.EOF {
					l.Errorf("get replication msg err: %s", err.Error())
				}

				if err := r.checkAndResetConn(); err != nil {
					l.Errorf("reset connection: %s", err.Error())
					time.Sleep(time.Second)
				} else {
					l.Info("reconnect success")
				}
				continue
			}
			r.replicationMsgHandle(msg)
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

	if r.slotName != "" {
		if err := r.deleteSelfSlot(); err != nil {
			l.Errorf("delete self slot err: %s", err.Error())
		}
	}

	r.slotName = fmt.Sprintf("datakit_slot_%d", time.Now().UnixNano())
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

func (r *Replication) getStatus() (*pgx.StandbyStatus, error) {
	return pgx.NewStandbyStatus(r.getReceivedWal())
}

// ReplicationMsgHandle handle replication msg
func (r *Replication) replicationMsgHandle(msg *pgx.ReplicationMessage) {
	if msg == nil {
		l.Error("replication message is empty")
		return
	}

	if msg.ServerHeartbeat != nil {

		if msg.ServerHeartbeat.ServerWalEnd > r.getReceivedWal() {
			r.setReceivedWal(msg.ServerHeartbeat.ServerWalEnd)
		}
		if msg.ServerHeartbeat.ReplyRequested == 1 {
			if err := r.sendStatus(); err != nil {
				l.Error(err)
			}
		}
	}

	if msg.WalMessage != nil {
		result := parselogical.NewParseResult(string(msg.WalMessage.WalData))
		if err := result.Parse(); err != nil {
			l.Errorf("parse message err: %s", err)
			return
		}

		if _, ok := r.eventsOperation[result.Operation]; !ok {
			l.Debugf("ignore this event %s", result.Operation)
			return
		}

		if r.Table != "" {
			schemaAndTable := strings.Split(result.Relation, ".")
			if len(schemaAndTable) != 2 || schemaAndTable[1] != r.Table {
				l.Debugf("ignore this table %s", schemaAndTable[1])
				return
			}
		}

		data, err := r.buildPoint(result)
		if err != nil {
			l.Errorf("build point err: %s", err)
			return
		}

		if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
			l.Errorf("io feed err: %s", err.Error())
		} else {
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (r *Replication) sendStatus() error {
	r.sendStatusLock.Lock()
	defer r.sendStatusLock.Unlock()

	status, err := r.getStatus()
	if err != nil {
		return fmt.Errorf("unable to create StandbyStatus object: %s", err)
	}

	err = r.replicationConn.SendStandbyStatus(status)
	if err != nil {
		return fmt.Errorf("unable to send StandbyStatus object: %s", err)
	}

	return nil
}

func (r *Replication) buildPoint(result *parselogical.ParseResult) ([]byte, error) {

	var tags = make(map[string]string)
	var fields = make(map[string]interface{}, len(r.fieldKeys))

	for key, column := range result.Columns {
		if _, ok := r.tagKeys[key]; ok {
			tags[key] = column.Value
		}
		if _, ok := r.fieldKeys[key]; ok {
			fields[key] = typeAssert(column.Type, column.Value)
		}
	}

	tags["relation"] = result.Relation
	tags["database"] = r.Database
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

func (r *Replication) deleteSelfSlot() error {

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", r.User, r.Password, r.Host, r.Port, r.Database)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("SELECT pg_drop_replication_slot('?')", r.slotName)
	if err != nil {
		return err
	}

	return nil
}
