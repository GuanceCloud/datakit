package hostdir

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/disk"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

// 并发用到的channel数据类型.
type dataChan struct {
	fileSize int64
	dirCount int64
}

const (
	FSTypeUnknown = "unknown"
)

func GetFileSystemType(path string) (string, error) {
	ptr := 0
	if runtime.GOOS == datakit.OSWindows {
		info, err := disk.Partitions(true)
		if err != nil {
			return FSTypeUnknown, fmt.Errorf("error get windows disk information: %w", err)
		}
		for i := 0; i < len(info); i++ {
			if strings.Contains(path, info[i].Device) {
				ptr = i
			}
		}
		return info[ptr].Fstype, nil
	} else {
		infos, err := disk.Usage(path)
		if err != nil {
			return FSTypeUnknown, fmt.Errorf("error get %s disk information: %w", runtime.GOOS, err)
		}
		return infos.Fstype, nil
	}
}

func GetFileOwnership(path string, host string) (string, error) {
	uid, err := Getuid(path, host)
	if err != nil {
		return FSTypeUnknown, fmt.Errorf("error get uid: %w", err)
	}
	u, err := user.LookupId(uid)
	if err != nil {
		return FSTypeUnknown, fmt.Errorf("error look for uid: %w", err)
	}
	return u.Username, nil
}

func Getuid(path string, host string) (string, error) {
	var uid string
	info, err := os.Stat(path)
	if host == "linux" || host == "darwin" {
		a := reflect.ValueOf(info.Sys()).Elem()
		uid = strconv.Itoa(int(a.FieldByName("Uid").Uint()))
	}
	return uid, err
}

func Getdirmode(path string) (string, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return FSTypeUnknown, fmt.Errorf("error get os stat information:%w", err)
	}
	mode := fileInfo.Mode()
	return mode.String(), nil
}

func dirents(path string) ([]os.FileInfo, bool) {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
		return nil, false
	}
	return entries, true
}

func walkDir(path string, chans chan dataChan, regslice []string) {
	entries, _ := dirents(path)
	for _, e := range entries {
		if e.IsDir() {
			chans <- dataChan{
				dirCount: 1,
			}
			walkDir(filepath.Join(path, e.Name()), chans, regslice)
		} else {
			flag := isreg(filepath.Join(path, e.Name()), regslice)
			if !flag {
				chans <- dataChan{
					fileSize: e.Size(),
				}
			}
		}
	}
}

func Startcollect(dir string, reslice []string) (int, int, int) {
	mychan := make(chan dataChan)
	var sizeCount int64

	var fileCount int

	var dirNum int64
	dirNum = 0

	go func() {
		walkDir(dir, mychan, reslice)
		defer close(mychan)
	}()

	for count := range mychan {
		fileCount++
		sizeCount += count.fileSize
		dirNum += count.dirCount
	}

	return int(sizeCount), fileCount, int(dirNum)
}

func isreg(filename string, regslice []string) bool {
	buf := filename
	flag := false
	for i := 0; i < len(regslice); i++ {
		reg := regexp.MustCompile(`^.+\.` + regslice[i] + `$`)
		result := reg.FindAllStringSubmatch(buf, 1)
		if len(result) != 0 {
			flag = true
			break
		}
	}
	return flag
}
