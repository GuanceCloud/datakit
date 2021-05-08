package httpPacket

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/luascript"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

func TestPoints(t *testing.T) {
	io.TestOutput()

	var luaCode = `
function handle(points)
	table.insert(points,
		{
			name="create_name1",
			time=1601285330806946000,
			tags={ t1="create_tags_01", t2="create_tags_02" },
			fields={ f1="create_fields_01", f5=555555, f6=true }
		}
	)
	for _, pt in pairs(points) do
		print("name", pt.name)
		print("time", pt.time)
		print("-------\ntags:")
		for k, v in pairs(pt.tags) do
			print(k, v)
		end
		print("-------\nfields:")
		for k, v in pairs(pt.fields) do
			print(k, v)
		end
		print("-----------------------")
	end
	return points
end
`
	pt1, _ := influxdb.NewPoint("point01",
		map[string]string{
			"t1": "tags10",
			"t2": "tags20",
		},
		map[string]interface{}{
			"f1": uint(11),
			"f2": true,
			"f3": "hello",
		},
		time.Now(),
	)
	pt2, _ := influxdb.NewPoint("point02",
		map[string]string{
			"t1": "tags11",
			"t2": "tags21",
		},
		map[string]interface{}{
			"f1": uint(33),
			"f2": int32(444),
			"f4": "world",
		},
		time.Now(),
	)

	var err error
	err = luascript.AddLuaCodes("ptdata", []string{luaCode})
	if err != nil {
		t.Fatal(err)
	}

	luascript.Run()

	p, err := NewPointsData("ptdata", datakit.Logging, []*influxdb.Point{pt1, pt2})
	if err != nil {
		t.Fatal(err)
	}

	luascript.SendData(p)

	time.Sleep(time.Second * 1)

	luascript.Stop()
}
