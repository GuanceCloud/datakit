package hostdir

import (
	"fmt"
	"github.com/shirou/gopsutil/disk"
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
)

var Dir_count int

func GetFileSystemType(path string) (string, error) {
	ptr := 0
	if runtime.GOOS == "windows" {
		info, err := disk.Partitions(true)
		if err != nil {
			return "unknown", fmt.Errorf("error get windows disk information:%s", err)
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
			return "unknown", fmt.Errorf("error get %s disk information:%s", runtime.GOOS, err)
		}
		return infos.Fstype, nil
	}
}

func GetFileOwnership(path string, host string) (string, error) {
	uid, err := Getuid(path, host)
	if err != nil {
		return "unknown", fmt.Errorf("error get uid:%s", err)
	}
	u, err := user.LookupId(uid)
	if err != nil {
		return "unknown", fmt.Errorf("error look for uid:%s", err)
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
		return "unknown", fmt.Errorf("error get os stat information:%s", err)
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

func walkDir(path string, fileSize chan<- int64, regslice []string) {
	var flag bool
	entries, _ := dirents(path)
	for _, e := range entries {
		if e.IsDir() {
			Dir_count++
			walkDir(filepath.Join(path, e.Name()), fileSize, regslice)
		} else {
			flag = false
			flag = isreg(filepath.Join(path, e.Name()), regslice)
			if !flag {
				fileSize <- e.Size()
			}
		}
	}
}

func Startcollect(dir string, reslice []string) (int, int, int) {

	fileSize := make(chan int64)

	var sizeCount int64

	var fileCount int

	//*dirCount = 0
	go func() {
		walkDir(dir, fileSize, reslice)
		defer close(fileSize)
	}()
	for size := range fileSize {
		fileCount++
		sizeCount += size
	}
	return int(sizeCount), fileCount, Dir_count

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
