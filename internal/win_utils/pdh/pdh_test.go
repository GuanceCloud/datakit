package pdh

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestGetDataThrPDH(t *testing.T) {

	objNameList := map[string]bool{
		"Web Service":  true,
		"APP_POOL_WAS": true,
	}
	for objName := range objNameList {

		// Enum the instance and counter of object
		instanceList, counterList, ret := PdhEnumObjectItems(objName)
		if ret != uint32(windows.ERROR_SUCCESS) {
			t.Errorf("Failed to enumerate the instance and counter of object %s", objName)
		}

		// Make counter full path and check it
		var pathList = make([]string, 0)
		for i := range instanceList {
			for j := range counterList {
				if r := PdhValidatePath(MakeFullCounterPath(objName, instanceList[i], counterList[j])); r != uint32(windows.ERROR_SUCCESS) {
					// Check full counter path
					t.Errorf("path %s invalid", MakeFullCounterPath(objName, instanceList[i], counterList[j]))
				}
				pathList = append(pathList, MakeFullCounterPath(objName, instanceList[i], counterList[j]))
			}
		}

		// Open query
		var handle PDH_HQUERY
		var counterHandle PDH_HCOUNTER
		ret = PdhOpenQuery(0, 0, &handle)
		if ret != uint32(windows.ERROR_SUCCESS) {
			t.Error("open query failed")
			return
		}

		// Add (english) counter
		var counterHandleList = make([]PDH_HCOUNTER, len(pathList))
		for i := range pathList {
			ret = PdhAddEnglishCounter(handle, pathList[i], 0, &counterHandle)
			counterHandleList[i] = counterHandle
			if ret != uint32(windows.ERROR_SUCCESS) {
				t.Error("error: add counter", pathList[i])
			}
		}

		// Call PDH query function
		// For some counter, it need to call twice
		if ret = PdhCollectQueryData(handle); ret != uint32(windows.ERROR_SUCCESS) {
			t.Errorf("PdhCollectQueryData return: %x\n", ret)
		}

		// If object name is `Web Service` and only call once,
		// will cause func such as PdhGetFormattedCounterValueDouble
		// return PDH_INVALID_DATA
		if objName == "Web Service" {
			if ret = PdhCollectQueryData(handle); ret != uint32(windows.ERROR_SUCCESS) {
				t.Errorf("PdhCollectQueryData return: %x\n", ret)
			}
		}

		// Get value
		var derp PDH_FMT_COUNTERVALUE_DOUBLE
		for i := 0; i < len(counterHandleList); i++ {
			ret = PdhGetFormattedCounterValueDouble(counterHandleList[i], nil, &derp)
			if ret != uint32(windows.ERROR_SUCCESS) {
				t.Errorf("PdhGetFormattedCounterValueDouble return: %x\nCounterFullPath: %s\n", ret, pathList[i])
			}
		}

		// Close query
		ret = PdhCloseQuery(handle)
		if ret != uint32(windows.ERROR_SUCCESS) {
			t.Error("PdhCloseQuery return: ", ret)
		}
	}
}
