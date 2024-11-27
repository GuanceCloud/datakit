// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package exporter collect RealTime data.
package exporter

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"gopkg.in/yaml.v2"

	_ "github.com/go-sql-driver/mysql"
)

var interfaceMain = map[int]string{
	0: "default",
	1: "non-default",
}

var interfaceType = map[int]string{
	1: "Agent",
	2: "SNMP",
	3: "IPMI",
	4: "JMX",
}

var itemTypes = map[int]string{
	0:  "Zabbix agent",
	1:  "unknown",
	2:  "Zabbix trapper",
	3:  "Simple check",
	5:  "Zabbix internal",
	7:  "Zabbix agent(active)",
	9:  "Web item",
	10: "External check",
	11: "Database monitor",
	12: "IPMI agent",
	13: "SSH agent",
	14: "TELNET agent",
	15: "Calculated",
	16: "JMX agent",
	17: "SNMP trap",
	18: "Dependent item",
	19: "HTTP agent",
	20: "SNMP agent",
	21: "Script",
	22: "Browser",
}

type Mysql struct {
	DBHost string `toml:"db_host"`
	DBPort string `toml:"db_port"`
	User   string `toml:"user"`
	PW     string `toml:"pw"`
}

type ItemC struct {
	ItemID    int64  `json:"itemid"`
	Name      string `json:"name"`
	Key_      string `json:"key_"` // net.if.in[if,mode]
	HostID    int64  `json:"hostid"`
	Units     string `json:"units"`
	TypeC     int    `json:"type"`
	itemType  string
	key       string   // 将 key_ 字符串的中括号去掉得到的就是key。
	tagValues []string // 按逗号分隔
}

func (ic *ItemC) setValues() {
	index := strings.Index(ic.Key_, "[")
	if index > 0 {
		ic.key = ic.Key_[0:index]
	}

	index = strings.Index(ic.Key_, "[")
	if index > 0 && strings.HasSuffix(ic.Key_, "]") {
		tag := ic.Key_[index+1 : len(ic.Key_)-1]
		// log.Debugf("from key_,tag = %s", tag)
		ic.tagValues = strings.Split(tag, ",")
	}
	if t, ok := itemTypes[ic.TypeC]; ok && t != "" {
		ic.itemType = t
	}
}

type InterfaceC struct {
	HostID        int64
	Main          int
	Type          int
	IP            string
	interfaceMain string
	interfaceType string
}

type HostC struct {
	// todo 全量获取host表
}

type Measurement struct {
	Measurement string   `yaml:"measurement"`
	Metric      string   `yaml:"metric"`
	Key         string   `yaml:"key"`
	Params      []string `yaml:"params"`
	Values      []string `yaml:"values"`
}

// CacheData : mysql tables: items,interface,hosts. and measurement yaml file.
type CacheData struct {
	lock         sync.RWMutex
	Items        map[int64]*ItemC
	Interfaces   map[int64]*InterfaceC
	Measurements map[string]*Measurement
	// Hosts      map[string]*HostC

	api  *ZabbixAPI
	cron *cron.Cron
}

func (cd *CacheData) start(my *Mysql, measurementDir string, crontab string) error {
	err := cd.readItemsAndInterface(my)
	if err != nil {
		log.Errorf("read item or interface err=%v", err)
		return err
	}
	err = cd.readMeasurementFromDir(measurementDir)
	if err != nil {
		log.Errorf("init MeasurementConfigDir err=%v", err)
		return err
	}

	cd.api.getToken()

	cd.cron = cron.New()
	_, err = cd.cron.AddFunc(crontab, func() {
		cd.crontab(my)
		cd.api.getToken() // 随着没次全量数据更新，token也更新一次
	})
	if err != nil {
		log.Errorf("cron err=%v", err)
		return err
	}
	cd.cron.Start()
	return nil
}

func (cd *CacheData) crontab(my *Mysql) {
	err := cd.readItemsAndInterface(my)
	if err != nil {
		log.Errorf("error read item or interface from mysql err=%v", err)
	}
}

func (cd *CacheData) readItemsAndInterface(my *Mysql) error {
	if my == nil {
		return fmt.Errorf("mysql is nil")
	}
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/zabbix", my.User, my.PW, my.DBHost, my.DBPort)
	// db, err := sql.Open("mysql", "root:123456#..A@tcp(49.232.153.84:3306)/zabbix")
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		log.Errorf("Error connecting to the database:%v", err)
		return err
	}
	defer db.Close() //nolint

	// 测试数据库连接
	err = db.Ping()
	if err != nil {
		log.Errorf("Error pinging the database:err=%v", err)
		return err
	}
	log.Infof("Successfully connected to MySQL!")

	err = cd.readItemsFromDB(db)
	if err != nil {
		return err
	}
	err = cd.readInterfaceFromDB(db)
	if err != nil {
		return err
	}
	return nil
}

func (cd *CacheData) readItemsFromDB(db *sql.DB) error {
	cd.lock.Lock()
	defer cd.lock.Unlock()
	rows, err := db.Query("select itemid,type,name,key_,hostid,units from items")
	if err != nil {
		log.Errorf("err =%v", err)
		return err
	}

	var count int
	for rows.Next() {
		var itemID int64
		var typeC int
		var name string
		var key string
		var units string
		var hostID int64

		if err = rows.Scan(&itemID, &typeC, &name, &key, &hostID, &units); err != nil {
			log.Errorf("rows.Scan err=%v", err)
			continue
		}
		count++
		// 表数据转对象 再存到cacheData中。
		ic := &ItemC{
			ItemID: itemID,
			TypeC:  typeC,
			Name:   name,
			Key_:   key,
			key:    key,
			HostID: hostID,
			Units:  units,
		}
		ic.setValues()

		cd.Items[itemID] = ic
	}

	log.Infof("all items len=%d", count)

	return nil
}

func (cd *CacheData) readInterfaceFromDB(db *sql.DB) error {
	cd.lock.Lock()
	defer cd.lock.Unlock()
	// Query interface.
	rows, err := db.Query("select hostid,main,type,ip from interface")
	if err != nil {
		log.Errorf("query interface err=%v", err)
		return err
	}

	for rows.Next() {
		var hostID int64
		var main int
		var t int
		var ip string

		if err = rows.Scan(&hostID, &main, &t, &ip); err != nil {
			log.Errorf("row interface err=%v", err)
			continue
		}

		i := &InterfaceC{
			HostID: hostID,
			Main:   main,
			Type:   t,
			IP:     ip,
		}
		i.interfaceMain = interfaceMain[i.Main]
		i.interfaceType = interfaceType[i.Type]

		cd.Interfaces[hostID] = i
	}
	log.Infof("all interface len =%d", len(cd.Interfaces))
	return nil
}

func (cd *CacheData) readMeasurementFromDir(dir string) error {
	cd.lock.Lock()
	defer cd.lock.Unlock()
	fs, err := os.ReadDir(dir)
	if err != nil {
		log.Errorf("read dir=%s err=%v", dir, err)
		return err
	}
	for _, entry := range fs {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), "yaml") {
			yamlFile, err := os.ReadFile(filepath.Join(dir, entry.Name())) //nolint
			if err != nil {
				log.Errorf("error: %v", err)
				continue
			}

			// 创建一个MetricConfig切片来存储解析后的数据
			var metrics []*Measurement

			// 解析YAML文件到MetricConfig切片
			err = yaml.Unmarshal(yamlFile, &metrics)
			if err != nil {
				log.Errorf("yaml.Unmarshal error: %v", err)
				return err
			}

			for _, metric := range metrics {
				cd.Measurements[metric.Key] = metric
			}
		}
	}
	log.Infof("read from yaml file, measurement.len=%d", len(cd.Measurements))
	return nil
}

func (cd *CacheData) getKeyName(defaultName string, itemID int64) string {
	cd.lock.Lock()
	defer cd.lock.Unlock()
	item, ok := cd.Items[itemID]
	if !ok {
		log.Infof("request zabbix api itemid=%d", itemID)
		start := time.Now()
		ic := cd.api.getItemByID(itemID)
		if ic == nil {
			RequestAPIVec.WithLabelValues("failed").Observe(float64(time.Since(start)) / float64(time.Millisecond))
			return defaultName
		}
		RequestAPIVec.WithLabelValues("success").Observe(float64(time.Since(start)) / float64(time.Millisecond))
		cd.Items[itemID] = ic
	}

	m, ok := cd.Measurements[item.key]
	if !ok {
		return item.key
	}

	return m.Metric
}

func (cd *CacheData) getTagsByItemID(itemID int64) map[string]string {
	cd.lock.RLock()
	defer cd.lock.RUnlock()
	tags := make(map[string]string)
	item, ok := cd.Items[itemID]
	if !ok {
		// todo  通过api查询
		return tags
	}
	log.Debugf("itemid=%d key=%s", item.ItemID, item.key)
	m, ok := cd.Measurements[item.key]
	tags["units"] = item.Units
	tags["item_type"] = item.itemType

	if ok {
		if len(m.Params) == len(m.Values) {
			if len(m.Params) != 0 && len(m.Values) != 0 {
				// 先填充默认值。
				for i, param := range m.Params {
					tags[param] = m.Values[i]
				}

				// 再填充item值。
				for i, value := range item.tagValues {
					if value != "" && len(m.Params) > i {
						tags[m.Params[i]] = value
					}
				}
			}

			tags["measurement"] = m.Measurement
		}
	}

	// 填充ip
	if inter, ok := cd.Interfaces[item.HostID]; ok {
		tags["ip"] = inter.IP
		if v, ok := interfaceMain[inter.Main]; ok {
			tags["interface_main"] = v
		}
		if v, ok := interfaceType[inter.Type]; ok {
			tags["interface_type"] = v
		}
	} else {
		tags["ip"] = "127.0.0.1"
	}

	return tags
}
