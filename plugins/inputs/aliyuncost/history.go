package aliyuncost

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type historyInfo struct {
	Start   string
	End     string
	Statue  int
	PageNum int

	key string
}

func SetAliyunCostHistory(key string, info *historyInfo) error {
	if data, err := json.Marshal(info); err != nil {
		return err
	} else {
		return ioutil.WriteFile(filepath.Join(historyCacheDir, key), data, 0755)
	}
}

func DelAliyunCostHistory(key string) {
	path := filepath.Join(historyCacheDir, key)
	os.Remove(path)
}

func GetAliyunCostHistory(key string) (*historyInfo, error) {
	path := filepath.Join(historyCacheDir, key)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
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
