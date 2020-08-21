package traceSkywalking

import (
	"bytes"
	"encoding/binary"
	"path/filepath"
	"sync"
	"time"

	"github.com/boltdb/bolt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	BoltDb         *bolt.DB
	DbRWLocker     sync.RWMutex
	RegService     = &sync.Map{} //key: id,           value: serviceName
	RegServiceRev  = &sync.Map{} //key: serviceName,  value: id
	RegInstance    = &sync.Map{} //key: id,           value: instanceUUID
	RegInstanceRev = &sync.Map{} //key: instanceUUID, value: id
	RegEndpoint    = &sync.Map{} //key: id,           value: endpointName
	RegEndpointRev = &sync.Map{} //key: endpointName, value: id

)

const (
	DbFile         = "skywalking-v2.db"
	ServiceBucket  = "service"
	InstanceBucket = "instance"
	EndpointBucket = "endpoint"
)

func BoltDbInit() {
	dbAddr := getBoltDbAddr()
	if err := ConnectBoltDb(dbAddr); err != nil {
		log.Error(err)
		return
	}

	CreateAllDbBucket()
	ReadFromDbByBucket(ServiceBucket, RegService, RegServiceRev)
	ReadFromDbByBucket(InstanceBucket, RegInstance, RegInstanceRev)
	ReadFromDbByBucket(EndpointBucket, RegEndpoint, RegEndpointRev)
}

func getBoltDbAddr() string {
	return filepath.Join(datakit.InstallDir, inputName, DbFile)
}

func ConnectBoltDb(addr string) error {
	var err error
	BoltDb, err = bolt.Open(addr, 0600, &bolt.Options{Timeout: 10 * time.Second})
	return err
}

func CreateAllDbBucket() {
	buckets := []string{ServiceBucket, InstanceBucket, EndpointBucket}
	for _, bucket := range buckets {
		CreateDbBucket(bucket)
	}
}

func CreateDbBucket(bucket string) {
	BoltDb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			log.Errorf("create bucket: %s", err)
			return err
		}
		return nil
	})

}
func ReadFromDbByBucket(bucket string, mapKV *sync.Map, mapRevKV *sync.Map) {
	BoltDb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var x int32
			msg := string(v)
			bytesBuffer := bytes.NewBuffer(k)
			binary.Read(bytesBuffer, binary.BigEndian, &x)
			mapKV.Store(x, msg)
			mapRevKV.Store(msg, x)
			log.Infof("%s -> key: %v value: %v", bucket, x, msg)
		}

		return nil
	})
}

func SaveRegInfo(bucket, msg string) (int32, error) {
	var gID int32

	DbRWLocker.Lock()
	defer DbRWLocker.Unlock()

	err := BoltDb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		id, _ := b.NextSequence()
		gID = int32(id)

		bytesBuffer := bytes.NewBuffer([]byte{})
		binary.Write(bytesBuffer, binary.BigEndian, gID)
		return b.Put(bytesBuffer.Bytes(), []byte(msg))
	})
	return gID, err
}
