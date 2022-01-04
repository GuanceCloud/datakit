package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestDefaultTime(t *testing.T) {
	// local timezone: utc+0800
	cst := time.FixedZone("CST", 8*3600)
	time.Local = cst

	cases := []struct {
		name, pl, in string
		outkey       string
		expect       interface{}
		fail         bool
	}{
		{
			name: "nginx log datetime, 02/Jan/2006:15:04:05 -0700",
			in:   `{"time":"02/Dec/2021:11:55:34 +0800"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "nginx log datetime, 02/Jan/2006:15:04:05 -0700",
			in:   `{"time":"02/Dec/2021:11:55:34 +0800"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "redis log datetime-tokyo",
			in:   `{"time":"02 Dec 2021 12:55:34.000"}`,
			pl: `
			json(_, time)
			default_time(time, "Asia/Tokyo")
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},

		{
			name: "redis log datetime-default tz",
			in:   `{"time":"02 Dec 2021 11:55:34.000"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "redis log datetime no year",
			in:   `{"time":"02 Dec 11:55:34.000"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: time.Date(time.Now().Year(), time.Month(12), 2, 11, 55, 34, 0, time.Now().Location()).UnixNano(),
			fail:   false,
		},
		{
			name: "mysql, 171113 14:14:20",
			in:   `{"time":"211202 11:55:34"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "gin, 2021/02/27 - 14:14:20",
			in:   `{"time":"2021/12/02 - 11:55:34"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "apache,  Tue May 18 06:25:05.176170 2021",
			in:   `{"time":"Tue Dec 2 11:55:34.000000 2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "postgresql, 2021-05-27 06:54:14.760 UTC",
			in:   `{"time":"2021-12-02 11:55:34.000 UTC"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Dec 2, 2021 11:55:34 AM",
			in:   `{"time":"Dec 2, 2021 11:55:34 AM"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Dec 2, 2021",
			in:   `{"time":"Dec 2, 2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "Dec 2, 2021 (tz: UTC)",
			in:   `{"time":"Dec 2, 2021"}`,
			pl: `
			json(_, time)
			default_time(time, "UTC")
		`,
			outkey: "time",
			expect: int64(1638403200000000000),
			fail:   false,
		},
		{
			name: "dec 2, '21 (tz: UTC)",
			in:   `{"time":"Dec 2, '21"}`,
			pl: `
			json(_, time)
			default_time(time, "UTC")
		`,
			outkey: "time",
			expect: int64(1638403200000000000),
			fail:   false,
		},
		{
			name: "Dec. 2, 2021 (tz: UTC)",
			in:   `{"time":"Dec 2, 2021"}`,
			pl: `
			json(_, time)
			default_time(time, "UTC")
		`,
			outkey: "time",
			expect: int64(1638403200000000000),
			fail:   false,
		},
		{
			name: "Dec. 2, 21 (tz: UTC)",
			in:   `{"time":"Dec 2, 21"}`,
			pl: `
			json(_, time)
			default_time(time, "UTC")
		`,
			outkey: "time",
			expect: int64(1638403200000000000),
			fail:   false,
		},
		{
			name: "Tue Dec  2 11:55:34 2021",
			in:   `{"time":"Tue Dec  2 11:55:34 2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Tue Dec  2 11:55:34 CST 2021",
			in:   `{"time":"Tue Dec  2 11:55:34 CST 2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Tue Dec 02 11:55:34 +0800 2021",
			in:   `{"time":"Tue Dec 02 11:55:34 +0800 2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Tuesday, 02-Dec-21 11:55:34 CST",
			in:   `{"time":"Tuesday, 02-Dec-21 11:55:34 CST"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Tue, 02 Dec 2021 11:55:34 CST",
			in:   `{"time":"Tue, 02 Dec 2021 11:55:34 CST"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Tue, 02 Dec 2021 11:55:34 +0800 (CST)",
			in:   `{"time":"Tue, 02 Dec 2021 11:55:34 +0800 (CST)"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Tue, 02 Dec 2021 11:55:34 +0800",
			in:   `{"time":"Tue, 02 Dec 2021 11:55:34 +0800"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Tue 02 Dec 2021 11:55:34 AM CST",
			in:   `{"time":"Tue 02 Dec 2021 11:55:34 AM CST"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Tue Dec 2 11:55:34 UTC+0800 2021",
			in:   `{"time":"Tue 02 Dec 2021 11:55:34 AM CST"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Tue, 2 Dec 2021 11:55:34 +0800",
			in:   `{"time":"Tue, 2 Dec 2021 11:55:34 +0800"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "Tue Dec 02 2021 11:55:34 GMT+0800 (GMT Daylight Time)",
			in:   `{"time":"Tue Dec 02 2021 11:55:34 GMT+0800 (GMT Daylight Time)"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "December 02, 2021 11:55:34am",
			in:   `{"time":"December 02, 2021 11:55:34am"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "December 02, 2021 at 11:55:34am CST+08",
			in:   `{"time":"December 02, 2021 at 11:55:34am CST+08"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "December 02, 2021 at 11:55:34am CST+08",
			in:   `{"time":"December 02, 2021 at 11:55:34am CST+08"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "December 02, 2021 at 11:55am CST+08",
			in:   `{"time":"December 02, 2021 at 11:55am CST+08"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417300000000000),
			fail:   false,
		},
		{
			name: "December 02, 2021, 11:55:34",
			in:   `{"time":"December 02, 2021, 11:55:34"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
		{
			name: "December 2, 2021",
			in:   `{"time":"December 2, 2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "December 2th, 2021",
			in:   `{"time":"December 2th, 2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "02 Dec 2021, 11:55",
			in:   `{"time":"02 Dec 2021, 11:55"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417300000000000),
			fail:   false,
		},
		{
			name: "2 Dec 2021, 11:55",
			in:   `{"time":"2 Dec 2021, 11:55"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417300000000000),
			fail:   false,
		},
		{
			name: "2 Dec 21",
			in:   `{"time":"2 Dec 21"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "2 Dec 2021",
			in:   `{"time":"2 Dec 2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "02 December 2021",
			in:   `{"time":"02 December 2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "2 December 2021",
			in:   `{"time":"2 December 2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "2021-Dec-02",
			in:   `{"time":"2021-Dec-02"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		// mm/dd/yy
		{
			name: "12/2/2021",
			in:   `{"time":"12/2/2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "12/02/2021",
			in:   `{"time":"12/02/2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "12/2/21",
			in:   `{"time":"12/2/21"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "12/2/2021 11:55",
			in:   `{"time":"12/2/2021 11:55"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417300000000000),
			fail:   false,
		},
		{
			name: "12/02/2021 11:55",
			in:   `{"time":"12/02/2021 11:55"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417300000000000),
			fail:   false,
		},
		{
			name: "12/02/2021 11:55:43",
			in:   `{"time":"12/02/2021 11:55:43"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417343000000000),
			fail:   false,
		},
		{
			name: "12/02/2021 11:55:43 AM",
			in:   `{"time":"12/02/2021 11:55:43 AM"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417343000000000),
			fail:   false,
		},
		{
			name: "12/02/2021 11:55 AM",
			in:   `{"time":"12/02/2021 11:55 AM"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417300000000000),
			fail:   false,
		},
		{
			name: "12/02/2021 11:55 11:55:43",
			in:   `{"time":"12/02/2021 11:55:43"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417343000000000),
			fail:   false,
		},
		{
			name: "12/02/2021 11:55 11:55:43.9999999",
			in:   `{"time":"12/02/2021 11:55:43.9999999"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417343999999900),
			fail:   false,
		},
		//yyyy:mm:dd
		{
			name: "2021:12:2",
			in:   `{"time":"2021:12:2"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "2021:12:02",
			in:   `{"time":"2021:12:02"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "2021:12:2 11:55",
			in:   `{"time":"2021:12:2 11:55"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417300000000000),
		},
		{
			name: "2021:12:02 11:55",
			in:   `{"time":"2021:12:02 11:55"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417300000000000),
		},
		{
			name: "2021:12:02 11:55:43",
			in:   `{"time":"2021:12:02 11:55:43"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021:12:2 12:55:43",
			in:   `{"time":"2021:12:2 12:55:43"}`,
			pl: `
			json(_, time)
			default_time(time, "Asia/Tokyo")
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021:12:2 11:55:43",
			in:   `{"time":"2021:12:2 11:55:43.89555"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343895550000),
		},
		{
			name: "2021年11月02日",
			in:   `{"time":"2021年12月02日"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
		},
		// yyyy-mm-ddThhs
		{
			name: "2021-12-2T11:55:43+0800",
			in:   `{"time":"2021:12:2 11:55:43"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2T11:55:43+08:00",
			in:   `{"time":"2021-12-2T11:55:43+08:00"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2T11:55:43",
			in:   `{"time":"2021-12-2T11:55:43"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2T3:55:43Z",
			in:   `{"time":"2021-12-2T3:55:43Z"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		// yyyy-mm-dd hh:mm:ss
		{
			name: "2021-12-2 11:55:43.12223",
			in:   `{"time":"2021-12-2 11:55:43.12223"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343122230000),
		},
		{
			name: "2021-12-2 11:55:43.122230000",
			in:   `{"time":"2021-12-2 11:55:43.122230000"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343122230000),
		},
		{
			name: "2021-12-2 11:55",
			in:   `{"time":"2021-12-2 11:55"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417300000000000),
		},
		{
			name: "2021-12-2 11:55:43",
			in:   `{"time":"2021-12-2 11:55:43"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2 11:55:43 UTC",
			in:   `{"time":"2021-12-2 11:55:43 UTC"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2 11:55:43 GMT",
			in:   `{"time":"2021-12-2 11:55:43 GMT"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2 11:55:43 CST",
			in:   `{"time":"2021-12-2 11:55:43 CST"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2 11:55:43 AM",
			in:   `{"time":"2021-12-2 11:55:43 AM"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2 11:55:43 +0800",
			in:   `{"time":"2021-12-2 11:55:43 +0800"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2 11:55:43 +0800 +08",
			in:   `{"time":"2021-12-2 11:55:43 +0800 +08"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2 11:55:43 +08:00",
			in:   `{"time":"2021-12-2 11:55:43 +08:00"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2 3:55:43.1 +0000 UTC",
			in:   `{"time":"2021-12-2 3:55:43.1 +0000 UTC"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343100000000),
		},
		{
			name: "2021-12-2 11:55:43.1 +0800 UTC",
			in:   `{"time":"2021-12-2 11:55:43.1 +0800 UTC"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343100000000),
		},
		{
			name: "2021-12-2 11:55:43 +0800 UTC",
			in:   `{"time":"2021-12-2 11:55:43 +0800 UTC"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2 11:55:43 +0800 GMT",
			in:   `{"time":"2021-12-2 11:55:43 +0800 GMT"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2 11:55:43 +0800 GMT m=+0.000000001",
			in:   `{"time":"2021-12-2 11:55:43 +0800 GMT m=+0.000000001"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-2 11:55:43.0001 +0800 GMT m=+0.000000001",
			in:   `{"time":"2021-12-2 11:55:43.0001 +0800 GMT m=+0.000000001"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000100000),
		},
		{
			name: "2021-12-2 11:55:43+08:00",
			in:   `{"time":"2021-12-2 11:55:43+08:00"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343000000000),
		},
		{
			name: "2021-12-02",
			in:   `{"time":"2021-12-02"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "2021-12",
			in:   `{"time":"2021-12"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638288000000000000),
			fail:   false,
		},
		{
			name: "2021",
			in:   `{"time":"2021"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1609430400000000000),
			fail:   false,
		},
		{
			name: "2021-12-2 11:55:43,212",
			in:   `{"time":"2021-12-2 11:55:43,212"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638417343212000000),
		},
		// mm.dd.yy
		{
			name: "12.2.2021",
			in:   `{"time":"12.2.2021"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638374400000000000),
		},
		{
			name: "12.02.2021",
			in:   `{"time":"12.02.2021"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638374400000000000),
		},
		{
			name: "12.02.21",
			in:   `{"time":"12.02.21"}`,
			pl: `
			json(_, time)
			default_time(time)
			`,
			outkey: "time",
			expect: int64(1638374400000000000),
		},
		{
			name: "2021.12",
			in:   `{"time":"2021.12"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638288000000000000),
			fail:   false,
		},
		{
			name: "2021.12.02",
			in:   `{"time":"2021.12.02"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		// yyyymmdd and similar
		{
			name: "20211202",
			in:   `{"time":"20211202"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638374400000000000),
			fail:   false,
		},
		{
			name: "20211202115543",
			in:   `{"time":"20211202115543"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417343000000000),
			fail:   false,
		},
		// unix seconds, ms, micro, nano
		{
			name: "1638417343",
			in:   `{"time":"1638417343"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417343000000000),
			fail:   false,
		},
		{
			name: "1638417343001",
			in:   `{"time":"1638417343001"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417343001000000),
			fail:   false,
		},
		{
			name: "1638417343001002",
			in:   `{"time":"1638417343001002"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417343001002000),
			fail:   false,
		},
		{
			name: "1638417343001002003",
			in:   `{"time":"1638417343001002003"}`,
			pl: `
			json(_, time)
			default_time(time)
		`,
			outkey: "time",
			expect: int64(1638417343001002003),
			fail:   false,
		},
		{
			name: "Dec 2, 2021 (tz: UTC)",
			in:   `{"time":"Dec 2, 2021"}`,
			pl: `
			json(_, time)
			default_time(time, "+0")
		`,
			outkey: "time",
			expect: int64(1638403200000000000),
			fail:   false,
		},
		{
			name: "Tue Dec  2 11:55:34 CST 2021",
			in:   `{"time":"Tue Dec  2 11:55:34 CST 2021"}`,
			pl: `
			json(_, time)
			default_time(time, "+8")
		`,
			outkey: "time",
			expect: int64(1638417334000000000),
			fail:   false,
		},
	}

	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.pl)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}

			if err := runner.Run(tc.in); err != nil {
				t.Error(err)
				return
			}
			t.Log(runner.Result())
			v, _ := runner.GetContent(tc.outkey)
			tu.Equals(t, tc.expect, v)
			t.Logf("[%d] PASS", idx)
		})
	}
}
