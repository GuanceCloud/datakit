package zabbix

import (
	"errors"
	"io/ioutil"
	"encoding/json"
	"sync"
)

type Registry struct {
	Table     string
	Startdate string
}

type MapTable map[string]string

//var mapTables = make(MapTable)

var mu sync.Mutex


func ReadRegistry(regName string, mapTables *MapTable) error {

	if _, err := ioutil.ReadFile(regName); err != nil {
		CreateRegistry() // create file if not exist
	}

	registryJson, err := ioutil.ReadFile(regName)
	if err != nil {
		return err
	}

	// parse JSON
	regEntries := make([]Registry, 0)
	if err := json.Unmarshal(registryJson, &regEntries); err != nil {
		return err
	}

	for i := 0; i < len(regEntries); i++ {
		tableName := regEntries[i].Table
		startdate := regEntries[i].Startdate
		SetValueByKey(mapTables, tableName, startdate)
	}

	return nil
}

func CreateRegistry() error {

	if len(zabbix.Tables) == 0 {
		return errors.New("No tables in configuration")
	}

	regEntries := make([]Registry, len(zabbix.Tables))

	var idx int = 0
	for _, table := range zabbix.Tables {
		var reg Registry
		reg.Table = table.Name
		reg.Startdate = table.Startdate
		regEntries[idx] = reg
		idx += 1
	}

	// write JSON file
	registryOutJson, _ := json.MarshalIndent(regEntries, "", "    ")
	ioutil.WriteFile(registryPath, registryOutJson, 0777)
	return nil
}

func SaveRegistry(registryPath string, tableName string, lastClock string) error {

	// read  file
	registryJson, err := ioutil.ReadFile(registryPath)
	if err != nil {
		return err
	}

	// parse JSON
	regEntries := make([]Registry, 0)
	if err := json.Unmarshal(registryJson, &regEntries); err != nil {
		return err
	}
	var found bool = false
	for i := 0; i < len(regEntries); i++ {
		if regEntries[i].Table == tableName {
			regEntries[i].Startdate = lastClock
			found = true
		}
	}
	// if not found, create it
	if found == false {
		regEntries = append(regEntries, Registry{tableName, lastClock})
	}

	// write JSON file
	registryOutJson, _ := json.MarshalIndent(regEntries, "", "    ")
	ioutil.WriteFile(registryPath, registryOutJson, 0777)

	return nil
}

func SetValueByKey(mt *MapTable, key string, value string) {
	mu.Lock()
	defer mu.Unlock()
	(*mt)[key] = value
}

func GetValueFromKey(mt MapTable, key string) string {
	if len(mt) > 0 {
		mu.Lock()
		defer mu.Unlock()
		return mt[key]
	}
	return ""
}