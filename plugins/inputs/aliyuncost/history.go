package aliyuncost

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

type historyInfo struct {
	Start   string
	End     string
	Current string
	Statue  int
	PageNum int
}

func SetAliyunCostHistory(key string, info *historyInfo) error {
	if data, err := json.Marshal(info); err != nil {
		return err
	} else {
		return ioutil.WriteFile(filepath.Join(config.ExecutableDir, key), data, 0755)
	}
}

func GetAliyunCostHistory(key string) (*historyInfo, error) {
	path := filepath.Join(config.ExecutableDir, key)
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
