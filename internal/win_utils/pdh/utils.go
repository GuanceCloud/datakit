package pdh

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func MakeFullCounterPath(objectName, instance, counter string) string {
	return fmt.Sprintf("\\%s(%s)\\%s", objectName, instance, counter)
}

func BufferToStringList(bufferLength uint32, buffer []uint16) []string {
	var dstList = make([]string, 0)
	if bufferLength < 2 {
		return nil
	} else {
		head := 0
		for tail := 0; tail < int(bufferLength)-1; tail++ {
			if buffer[tail] == 0 {
				dstList = append(dstList, windows.UTF16ToString(buffer[head:tail]))
				head = tail + 1
				if buffer[tail+1] == 0 {
					break
				}
			}
		}
	}
	return dstList
}
