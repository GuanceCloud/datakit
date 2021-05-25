package consumerLibrary

import (
	"github.com/aliyun/aliyun-log-go-sdk"
	"reflect"
	"time"
)

// List removal of duplicate elements
func Set(slc []int) []int {
	result := []int{}
	tempMap := map[int]byte{}
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l {
			result = append(result, e)
		}
	}
	return result
}

// Get the difference between the two lists
func Subtract(a []int, b []int) (diffSlice []int) {

	lengthA := len(a)

	if len(a) == 0 {
		return b
	}

	for _, valueB := range b {

		temp := valueB
		for j := 0; j < lengthA; j++ {
			if temp == a[j] {
				break
			} else {
				if lengthA == (j + 1) {
					diffSlice = append(diffSlice, temp)

				}
			}
		}
	}

	return diffSlice
}

// Returns the smallest of two numbers
func Min(a, b int64) int64 {
	if a > b {
		return b
	}
	return a
}

// Determine whether two lists are equal
func IntSliceReflectEqual(a, b []int) bool {
	if a == nil {
		a = []int{}
	}
	if b == nil {
		b = []int{}
	}
	return reflect.DeepEqual(a, b)
}

// Determine whether obj is in target object
func Contain(obj interface{}, target interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}

	return false
}

// Get the total number of logs
func GetLogCount(logGroupList *sls.LogGroupList) int {
	if logGroupList == nil {
		return 0
	}
	count := 0
	for _, logGroup := range logGroupList.LogGroups {
		count = count + len(logGroup.Logs)
	}
	return count
}

func GetLogGroupCount(logGroupList *sls.LogGroupList) int {
	return len(logGroupList.LogGroups)
}

func TimeToSleepInMillsecond(intervalTime, lastCheckTime int64, condition bool) {
	timeToSleep := intervalTime - (time.Now().UnixNano()/1000/1000 - lastCheckTime)
	for timeToSleep > 0 && !condition {
		time.Sleep(time.Duration(Min(timeToSleep, 100)) * time.Millisecond)
		timeToSleep = intervalTime - (time.Now().UnixNano()/1000/1000 - lastCheckTime)
	}
}

func TimeToSleepInSecond(intervalTime, lastCheckTime int64, condition bool) {
	timeToSleep := intervalTime*1000 - (time.Now().Unix()-lastCheckTime)*1000
	for timeToSleep > 0 && !condition {
		time.Sleep(time.Duration(Min(timeToSleep, 1000)) * time.Millisecond)
		timeToSleep = intervalTime*1000 - (time.Now().Unix()-lastCheckTime)*1000
	}
}
