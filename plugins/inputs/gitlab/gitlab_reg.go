package gitlab

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

var locker sync.Mutex
var pBT map[string]string

func loadPBT(filename string) {
	pBT = make(map[string]string)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	json.Unmarshal(bytes, &pBT)
}

func updatePBT(key string, t time.Time) {
	tStr := t.Format(time.RFC3339)
	locker.Lock()
	pBT[key] = tStr
	locker.Unlock()
}

func getPBT(key string) (time.Time, error) {
	locker.Lock()
	tStr, ok := pBT[key]
	locker.Unlock()

	if !ok {
		err := fmt.Errorf("Key: %s not found", key)
		return time.Time{}, err
	}

	return parseTimeStr(tStr)
}

func parseTimeStr(timeStr string) (time.Time, error) {
	startTime, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		startTime, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return startTime, err
		}
	}
	return startTime, nil
}

func flushPBT(fileName string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		locker.Lock()
		byteS, err := json.Marshal(pBT)
		locker.Unlock()
		if err != nil {
			continue
		}

		var str bytes.Buffer
		err = json.Indent(&str, byteS, "", "    ")
		if err != nil {
			continue
		}
		ioutil.WriteFile(fileName, str.Bytes(), 0x666)
		internal.SleepContext(ctx, time.Duration(1)*time.Minute)
	}
}
