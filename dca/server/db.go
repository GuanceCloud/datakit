// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

const (
	maxUpdatedAtInterval = 24 * time.Hour // list datakits which updated_at > now - maxUpdatedAtInterval
)

type (
	DB struct {
		db *sqlx.DB
	}
)

const (
	sqlCreateTable = `
create table if not exists datakit (
	id integer primary key autoincrement not null,
	runtime_id text not null,
	arch text not null,
	host_name text not null,
	os text not null,
	run_mode text not null,
	usage_cores int not null,
	version text not null,
	updated_at int not null,
	workspace_uuid text not null,
	ip text not null,
	start_time int not null,
	run_in_container bool not null,
	conn_id string,
	url string,
	status text not null
);

create unique index if not exists datakit_conn_id_index on datakit(conn_id,runtime_id);
`
)

const DefaultDBPath = ":memory:"

func (db *DB) Init() error {
	// open db
	if v, err := sqlx.Open("sqlite", dbPath); err != nil {
		return fmt.Errorf("open sqlite db failed: %w", err)
	} else {
		l.Infof("open sqlite db %s success", dbPath)
		v.SetMaxOpenConns(1)
		db.db = v
	}

	// create table
	if _, err := db.Exec(sqlCreateTable); err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}

	// update datakit status
	if _, err := db.Exec("update datakit set status=?", ws.StatusOffline); err != nil {
		l.Errorf("update datakit failed: %s", err.Error())
		return fmt.Errorf("Init db error:%w ", err)
	}

	if _, err := db.Exec("delete from datakit where updated_at < ? or (run_in_container=true)",
		time.Now().Add(-maxUpdatedAtInterval).UnixMilli()); err != nil {
		return fmt.Errorf("delete old datakit error:%w ", err)
	}

	l.Info("init db success")

	return nil
}

func (db *DB) Select(sql string, res any, args ...interface{}) error {
	return db.db.Select(res, sql, args...)
}

func (db *DB) Exec(sql string, args ...interface{}) (sql.Result, error) {
	return db.db.Exec(sql, args...)
}

// ForceUpdate delete datakit by connID and insert new datakit.
func (db *DB) ForceUpdate(dk *ws.DataKit) error {
	if dk == nil {
		l.Warnf("dks is empty")
		return nil
	}
	if err := db.DeleteByConnID(dk.ConnID, true); err != nil {
		return fmt.Errorf("failed to delete datakit: %w", err)
	}

	return db.Insert(dk)
}

func (db *DB) Update(dk *ws.DataKit) error {
	updatedAt := time.Now().UnixMilli()
	//nolint:lll
	_, err := db.Exec("update datakit set arch=?,host_name=?,os=?,version=?,ip=?,start_time=?,run_in_container=?,run_mode=?,usage_cores=?,updated_at=?,workspace_uuid=?,status=?,url=? where conn_id=?",
		dk.Arch, dk.HostName, dk.OS, dk.Version, dk.IP, dk.StartTime, dk.RunInContainer, dk.RunMode, dk.UsageCores, updatedAt, dk.WorkspaceUUID, dk.Status.String(), dk.URL, dk.ConnID)
	return err
}

func (db *DB) UpdateInsert(dk *ws.DataKit) error {
	if v, err := db.Find(dk); err != nil {
		return fmt.Errorf("find dk failed: %w", err)
	} else if v == nil {
		return db.Insert(dk)
	}
	return db.Update(dk)
}

func (db *DB) UpdateByConnID(dk *ws.DataKit, connID string) error {
	updatedAt := time.Now().UnixMilli()
	//nolint:lll
	_, err := db.Exec("update datakit set runtime_id=?,arch=?,host_name=?,os=?,version=?,ip=?,start_time=?,run_in_container=?,run_mode=?,usage_cores=?,updated_at=?,workspace_uuid=?,status=?,url=? where conn_id=?",
		dk.RunTimeID, dk.Arch, dk.HostName, dk.OS, dk.Version, dk.IP, dk.StartTime, dk.RunInContainer, dk.RunMode, dk.UsageCores, updatedAt, dk.WorkspaceUUID, dk.Status.String(), dk.URL, connID)
	return err
}

func (db *DB) BatchUpdate(dks []*ws.DataKit) error {
	for _, dk := range dks {
		if err := db.UpdateInsert(dk); err != nil {
			return fmt.Errorf("failed to update datakit: %w", err)
		}
	}

	return nil
}

// checkStatus check if the status transition is valid.
// running -> all state
// all state -> offline
// upgrading | restarting | offline | stopped -> running.
func checkStatus(from, to ws.DataKitStatus) bool {
	if (from == to) || (from == ws.StatusRunning) || (to == ws.StatusOffline) {
		return true
	}

	return to == ws.StatusRunning
}

func (db *DB) UpdateStatus(dk *ws.DataKit, status ws.DataKitStatus) error {
	if dk == nil {
		return nil
	}

	targetDK, err := db.Find(dk)
	if err != nil {
		return fmt.Errorf("failed to find datakit: %w", err)
	}

	if targetDK == nil {
		return fmt.Errorf("datakit not found")
	}

	if !checkStatus(targetDK.Status, status) {
		return fmt.Errorf("invalid status transition: %s -> %s", targetDK.Status, status)
	}

	_, err = db.Exec("update datakit set status=? where conn_id=?", status.String(), dk.ConnID)
	return err
}

func (db *DB) Heartbeat(connID string) error {
	updatedAt := time.Now().UnixMilli()
	_, err := db.Exec("update datakit set updated_at=? where conn_id=?", updatedAt, connID)
	return err
}

func (db *DB) Insert(dk *ws.DataKit) error {
	if dk == nil {
		return nil
	}

	updatedAt := time.Now().UnixMilli()
	//nolint:lll
	_, err := db.Exec("insert into datakit(runtime_id,arch,host_name,os,version,ip,start_time,run_in_container,run_mode,usage_cores,updated_at,workspace_uuid,conn_id,status,url) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		dk.RunTimeID, dk.Arch, dk.HostName, dk.OS, dk.Version, dk.IP, dk.StartTime, dk.RunInContainer, dk.RunMode, dk.UsageCores, updatedAt, dk.WorkspaceUUID, dk.ConnID, ws.StatusRunning, dk.URL)

	return err
}

func (db *DB) BatchInsert(dks []*ws.DataKit) error {
	if len(dks) == 0 {
		l.Warnf("dks is empty")
		return nil
	}
	args := []interface{}{}
	updatedAt := time.Now().UnixMilli()

	//nolint:lll
	sql := "insert into datakit(runtime_id,arch,host_name,os,version,ip,start_time,run_in_container,run_mode,usage_cores,updated_at,workspace_uuid,conn_id,status,url) values"
	for i, dk := range dks {
		if i == 0 {
			sql += "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		} else {
			sql += ",(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		}
		args = append(args, dk.RunTimeID, dk.Arch, dk.HostName,
			dk.OS, dk.Version, dk.IP, dk.StartTime, dk.RunInContainer,
			dk.RunMode, dk.UsageCores, updatedAt, dk.WorkspaceUUID, dk.ConnID,
			ws.StatusRunning, dk.URL)
	}

	_, err := db.Exec(sql, args...)

	return err
}

// Delete update datakit status to offline when datakit is online.
// Don't delete datakit when datakit is upgrading.
func (db *DB) Delete(dk *ws.DataKit, isHard bool) error {
	if dk == nil {
		return nil
	}
	if isHard {
		_, err := db.Exec("delete from datakit where conn_id=?", dk.ConnID)
		return err
	}

	_, err := db.Exec("update datakit set status=? where conn_id=?", ws.StatusOffline, dk.ConnID)
	return err
}

func (db *DB) DeleteByConnID(connID string, isHard bool) error {
	if isHard {
		_, err := db.Exec("delete from datakit where conn_id=?", connID)
		return err
	}

	_, err := db.Exec("update datakit set status=? where conn_id=?", ws.StatusOffline, connID)
	return err
}

func (db *DB) Find(dk *ws.DataKit) (*ws.DataKit, error) {
	rows := []ws.DataKit{}
	if err := db.Select("select * from datakit where conn_id=?", &rows, dk.ConnID); err != nil {
		return nil, fmt.Errorf("failed to query datakit: %w", err)
	}

	switch {
	case len(rows) == 0:
		return nil, nil
	case len(rows) > 1:
		return nil, fmt.Errorf("duplicated datakit found")
	default:
		return &rows[0], nil
	}
}

// IsDuplicatedConn checks whether the connection id is duplicated.
func (db *DB) IsDuplicatedConn(dk *ws.DataKit) (bool, error) {
	if dk == nil {
		return false, nil
	}
	connID := dk.ConnID

	rows := []ws.DataKit{}
	if err := db.Select("select * from datakit where conn_id=? and status<>?", &rows, connID, ws.StatusOffline); err != nil {
		return false, fmt.Errorf("failed to query datakit: %w", err)
	}

	return len(rows) > 0, nil
}
