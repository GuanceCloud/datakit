//go:build windows
// +build windows

package winevent

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	bufferSize := 1 << 14
	w := Input{
		buf: make([]byte, bufferSize),
	}
	w.Query = query

	w.Run()
}

func TestHandle(t *testing.T) {
	msg := `{
    "Source":{
        "Name":"Microsoft-Windows-Security-Auditing"
    },
    "EventID":4624,
    "Version":2,
    "Level":0,
    "Task":12544,
    "Opcode":0,
    "Keywords":"审核成功",
    "TimeCreated":{
        "SystemTime":"2021-06-15T11:33:59.981585000Z"
    },
    "EventRecordID":10771,
    "Correlation":{
        "ActivityID":"{2161fd5e-61c8-0001-cbfd-6121c861d701}",
        "RelatedActivityID":""
    },
    "Execution":{
        "ProcessID":632,
        "ThreadID":676,
        "ProcessName":""
    },
    "Channel":"Security",
    "Computer":"MSEDGEWIN10",
    "Security":{
        "UserID":""
    },
    "UserData":{
        "InnerXML":null
    },
    "EventData":{
        "InnerXML":"PERhdGEgTmFtZT0nU3ViamVjdFVzZXJTaWQnPlMtMS01LTE4PC9EYXRhPjxEYXRhIE5hbWU9J1N1YmplY3RVc2VyTmFtZSc+TVNFREdFV0lOMTAkPC9EYXRhPjxEYXRhIE5hbWU9J1N1YmplY3REb21haW5OYW1lJz5XT1JLR1JPVVA8L0RhdGE+PERhdGEgTmFtZT0nU3ViamVjdExvZ29uSWQnPjB4M2U3PC9EYXRhPjxEYXRhIE5hbWU9J1RhcmdldFVzZXJTaWQnPlMtMS01LTE4PC9EYXRhPjxEYXRhIE5hbWU9J1RhcmdldFVzZXJOYW1lJz5TWVNURU08L0RhdGE+PERhdGEgTmFtZT0nVGFyZ2V0RG9tYWluTmFtZSc+TlQgQVVUSE9SSVRZPC9EYXRhPjxEYXRhIE5hbWU9J1RhcmdldExvZ29uSWQnPjB4M2U3PC9EYXRhPjxEYXRhIE5hbWU9J0xvZ29uVHlwZSc+NTwvRGF0YT48RGF0YSBOYW1lPSdMb2dvblByb2Nlc3NOYW1lJz5BZHZhcGkgIDwvRGF0YT48RGF0YSBOYW1lPSdBdXRoZW50aWNhdGlvblBhY2thZ2VOYW1lJz5OZWdvdGlhdGU8L0RhdGE+PERhdGEgTmFtZT0nV29ya3N0YXRpb25OYW1lJz4tPC9EYXRhPjxEYXRhIE5hbWU9J0xvZ29uR3VpZCc+ezAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMH08L0RhdGE+PERhdGEgTmFtZT0nVHJhbnNtaXR0ZWRTZXJ2aWNlcyc+LTwvRGF0YT48RGF0YSBOYW1lPSdMbVBhY2thZ2VOYW1lJz4tPC9EYXRhPjxEYXRhIE5hbWU9J0tleUxlbmd0aCc+MDwvRGF0YT48RGF0YSBOYW1lPSdQcm9jZXNzSWQnPjB4MjcwPC9EYXRhPjxEYXRhIE5hbWU9J1Byb2Nlc3NOYW1lJz5DOlxXaW5kb3dzXFN5c3RlbTMyXHNlcnZpY2VzLmV4ZTwvRGF0YT48RGF0YSBOYW1lPSdJcEFkZHJlc3MnPi08L0RhdGE+PERhdGEgTmFtZT0nSXBQb3J0Jz4tPC9EYXRhPjxEYXRhIE5hbWU9J0ltcGVyc29uYXRpb25MZXZlbCc+JSUxODMzPC9EYXRhPjxEYXRhIE5hbWU9J1Jlc3RyaWN0ZWRBZG1pbk1vZGUnPi08L0RhdGE+PERhdGEgTmFtZT0nVGFyZ2V0T3V0Ym91bmRVc2VyTmFtZSc+LTwvRGF0YT48RGF0YSBOYW1lPSdUYXJnZXRPdXRib3VuZERvbWFpbk5hbWUnPi08L0RhdGE+PERhdGEgTmFtZT0nVmlydHVhbEFjY291bnQnPiUlMTg0MzwvRGF0YT48RGF0YSBOYW1lPSdUYXJnZXRMaW5rZWRMb2dvbklkJz4weDA8L0RhdGE+PERhdGEgTmFtZT0nRWxldmF0ZWRUb2tlbic+JSUxODQyPC9EYXRhPg=="
    },
    "Message":"已成功登录帐户。",
    "LevelText":"信息",
    "TaskText":"Logon",
    "OpcodeText":"信息"
}`
	e := Event{}
	err := json.Unmarshal([]byte(msg), &e)
	if err != nil {
		l.Error(err.Error())
		return
	}
	c := Input{}
	c.handleEvent(e)
	for _, v := range c.collectCache {
		if line, err := v.LineProto(); err != nil {
			l.Error(err.Error())
		} else {
			fmt.Println(line.String())
		}
	}
}
