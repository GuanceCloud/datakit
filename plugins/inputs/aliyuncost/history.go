package aliyuncost

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type historyInfo struct {
	Start   string
	End     string
	Statue  int
	PageNum int

	//for bills
	StartTime time.Time
	EndTime   time.Time

	key string
}

func setAliyunCostHistory(key string, info *historyInfo) error {
	if data, err := json.Marshal(info); err != nil {
		return err
	} else {
		os.MkdirAll(historyCacheDir, 0755)
		return ioutil.WriteFile(filepath.Join(historyCacheDir, key), data, 0755)
	}
}

func delAliyunCostHistory(key string) {
	path := filepath.Join(historyCacheDir, key)
	os.Remove(path)
}

func getAliyunCostHistory(key string) (*historyInfo, error) {
	path := filepath.Join(historyCacheDir, key)
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	var info historyInfo

	if err = json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}
