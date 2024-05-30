//go:build linux
// +build linux

package protodec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

func TestPrefix(t *testing.T) {
	t.Run("POST", func(t *testing.T) {
		data := []byte("POST / HTTP/1.1\r\n\r\n")
		v, ok := httpMethod(data)
		if !ok {
			t.Fatal("not req start")
		}
		assert.Equal(t, "POST", v)
	})

	t.Run("GET", func(t *testing.T) {
		data := []byte("GET / HTTP/1.1\r\n\r\n")
		v, ok := httpMethod(data)
		if !ok {
			t.Fatal("not req start")
		}
		assert.Equal(t, "GET", v)
	})

	t.Run("HTTP", func(t *testing.T) {
		data := []byte("HTTP/1.1 200 OK\r\n\r\n")
		v, code, ok := httpProtoVersion(data)
		if !ok {
			t.Fatal("not resp start")
		}
		assert.Equal(t, "1.1", v)
		assert.Equal(t, 200, code)
	})
}

func TestMysql(t *testing.T) {
	m := &mysqlInfo{}
	t.Run("DetectMysql", func(t *testing.T) {
		mysqlReqMsg := "\x17\x00\x00\x00\x03\x43\x52\x45\x41\x54\x45\x20\x44\x41\x54\x41\x42\x41\x53\x45\x20" +
			"\x74\x65\x73\x74\x64\x62"
		seq, err := detectMysql([]byte(mysqlReqMsg), len(mysqlReqMsg))
		expectedSeq := 0
		if err != nil {
			t.Fatal("not client req create")
		}
		assert.Equal(t, expectedSeq, int(seq))
	})

	t.Run("Client", func(t *testing.T) {
		mysqlReqMsg := "\x03\x43\x52\x45\x41\x54\x45\x20\x44\x41\x54\x41\x42\x41\x53\x45\x20" +
			"\x74\x65\x73\x74\x64\x62"
		expected := "CREATE DATABASE testdb"
		msg, stmtId, err := m.isClientMsg([]byte(mysqlReqMsg), comm.FnSysRecvfrom)
		if err != nil {
			t.Fatal("not client req")
		}
		assert.Equal(t, expected, msg)
		assert.Equal(t, 0, stmtId)
	})

	t.Run("InsertPreStmt", func(t *testing.T) {
		insertReqMsg := "\x16\x49\x4e\x53\x45\x52\x54\x20\x49\x4e\x54\x4f\x20" +
			"\x63\x69\x74\x69\x65\x73\x28\x6e\x61\x6d\x65\x2c\x20\x70\x6f\x70\x75\x6c\x61\x74\x69" +
			"\x6f\x6e\x29\x20\x56\x41\x4c\x55\x45\x53\x28\x3f\x2c\x20\x3f\x29\x3b"
		expected := "INSERT INTO cities(name, population) VALUES(?, ?);"
		msg, stmtId, err := m.isClientMsg([]byte(insertReqMsg), comm.FnSysRecvfrom)
		if err != nil {
			t.Fatal("not client req")
		}
		assert.Equal(t, expected, msg)
		assert.Equal(t, -1, stmtId)
	})

	t.Run("InsertPreStmtServer", func(t *testing.T) {
		insertResMsg := "\x00\x01\x00\x00\x00\x00\x00\x02\x00\x00\x00\x00"
		msg, code, err := m.isServerMsg([]byte(insertResMsg), comm.FnSysSendto)
		if err != nil {
			t.Fatal("not server req")
		}
		t.Log(msg)
		assert.Equal(t, 0, code)
	})

	t.Run("Server", func(t *testing.T) {
		mysqlRespMsg := "\xff\xef\x03\x23\x48\x59\x30\x30\x30\x43\x61\x6e\x27\x74\x20" +
			"\x63\x72\x65\x61\x74\x65\x20\x64\x61\x74\x61\x62\x61\x73\x65\x20\x27\x74\x65" +
			"\x73\x74\x64\x62\x27\x3b\x20\x64\x61\x74\x61\x62\x61\x73\x65\x20\x65\x78\x69" +
			"\x73\x74\x73"
		expectedMsg := "Error Code: 1007, Error Msg: Can't create database 'testdb'; database exists"
		expectedCode := 1007
		msg, code, err := m.isServerMsg([]byte(mysqlRespMsg), 0)
		if err != nil {
			t.Fatal("not server req")
		}
		assert.Equal(t, expectedCode, code)
		assert.Equal(t, expectedMsg, msg)
	})
}
