package config

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

var tomlParseCases = []struct {
	in string
}{
	{
		in: `
		[[inputs.abc]]
			key1 = "1-line-string"
			key2 = '''multili
			string
			'''`,
	},

	{
		in: `
[[inputs.abc]]
	key1 = 1
	key2 = "a"
	key3 = 3.14`,
	},

	{
		in: `
[[inputs.abc]]
	key1 = 11
	key2 = "aa"
	key3 = 6.28`,
	},

	{
		in: `
[[inputs.abc]]
	key1 = 22
	key2 = "aaa"
	key3 = 6.18`,
	},
	{
		in: `
[[inputs.def]]
	key1 = 22
	key2 = "aaa"
	key3 = 6.18`,
	},
}

func TestTomlParse(t *testing.T) {

	type obj struct {
		Key1 int     `toml:"key1"`
		Key2 string  `toml:"key2"`
		Key3 float64 `toml:"key3"`
	}

	for _, tcase := range tomlParseCases {
		tbl, err := toml.Parse([]byte(tcase.in))
		if err != nil {
			t.Fatal(err)
		}

		if tbl.Fields == nil {
			t.Fatal("empty data")
		}

		for f, v := range tbl.Fields {

			switch f {

			default:
				// ignore
				t.Logf("ignore %+#v", f)

			case "inputs":
				switch tpe := v.(type) {
				case *ast.Table:
					stbl := v.(*ast.Table)

					for _, vv := range stbl.Fields {
						switch tt := vv.(type) {
						case []*ast.Table:
							for idx, elem := range vv.([]*ast.Table) {
								t.Logf("[%d] %+#v, source: %s", idx, elem, elem.Source())
							}
						case *ast.Table:
							t.Logf("%+#v, source: %s", vv.(*ast.Table), vv.(*ast.Table).Source())
						default:
							t.Logf("bad data: %v", tt)
						}
					}

				default:
					t.Logf("unknown type: %v", tpe)
				}
			}
		}
	}
}

//func TestEnableInputs(t *testing.T) {
//	fpath, sample, err := doEnableInput("timezone")
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	t.Logf("fpath: %s, sample: %s", fpath, sample)
//}

func TestBuildInputCfg(t *testing.T) {

	data := `
# Read metrics about disk IO by device
[[inputs.diskio]]
  # By default, telegraf will gather stats for all devices including
  # disk partitions.
  # Setting devices will restrict the stats to the specified devices.
   devices = ["sda", "sdb", "vd*"]
  # Uncomment the following line if you need disk serial numbers.
   skip_serial_number = false

  # On systems which support it, device metadata can be added in the form of
  # tags.
  # Currently only Linux is supported via udev properties. You can view
  # available properties for a device by running:
  # 'udevadm info -q property -n /dev/sda'
  # Note: Most, but not all, udev properties can be accessed this way. Properties
  # that are currently inaccessible include DEVTYPE, DEVNAME, and DEVPATH.
   device_tags = ["ID_FS_TYPE", "ID_FS_USAGE"]

  # Using the same metadata source as device_tags, you can also customize the
  # name of the device via templates.
  # The 'name_templates' parameter is a list of templates to try and apply to
  # the device. The template may contain variables in the form of '$PROPERTY' or
  # '${PROPERTY}'. The first template which does not contain any variables not
  # present for the device is used as the device name tag.
  # The typical use case is for LVM volumes, to get the VG/LV name instead of
  # the near-meaningless DM-0 name.
   name_templates = ["$ID_FS_LABEL","$DM_VG_NAME/$DM_LV_NAME"]

	[inputs.diskio.tags]
	host = '{{.Hostname}}'`

	datakit.Cfg.MainCfg.Hostname = "this-is-the-test-host-name"
	sample, err := BuildInputCfg([]byte(data), datakit.Cfg.MainCfg)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("sample: %s", sample)
}

//func TestLoadMainCfg(t *testing.T) {
//
//	c := datakit.Cfg
//	if err := c.LoadMainConfig(); err != nil {
//		t.Errorf("%s", err)
//	}
//}
//
func TestTomlUnmarshal(t *testing.T) {
	x := []byte(`
global = "global config"
[[inputs.abc]]
	key1 = 1
	key2 = "a"
	key3 = 3.14

[[inputs.abc]]
	key1 = 11
	key2 = "aa"
	key3 = 6.28

[[inputs.def]]
	key1 = 22
	key2 = "aaa"
	key3 = 6.18

[inputs.xyz]
	key1 = 22
	key2 = "aaa"
	key3 = 6.18

	[[inputs.xyz.tags]]
		key1 = 22
		key2 = "aaa"
		key3 = 6.18
		#key4 = 7.18
	`)

	tbl, err := toml.Parse(x)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("tbl: %+#v", tbl)
	t.Log(tbl.Source())

	for f, v := range tbl.Fields {

		switch f {

		default:
			// ignore
			t.Logf("ignore %+#v", f)

		case "inputs":
			switch tpe := v.(type) {
			case *ast.Table:
				stbl := v.(*ast.Table)

				for _, vv := range stbl.Fields {
					switch tt := vv.(type) {
					case []*ast.Table:
						for idx, elem := range vv.([]*ast.Table) {
							t.Logf("[%d] %+#v, source: %s", idx, elem, elem.Source())
						}
					case *ast.Table:
						t.Logf("%+#v, source: %s", vv.(*ast.Table), vv.(*ast.Table).Source())
					default:
						t.Logf("bad data: %v", tt)
					}
				}

			default:
				t.Logf("unknown type: %v", tpe)
			}
		}
	}
}

func TestInitCfg(t *testing.T) {
	// TODO
}

func TestBlackWhiteList(t *testing.T) {
	wlists := []*datakit.InputHostList{
		&datakit.InputHostList{
			Hosts:  []string{"host1", "host2"},
			Inputs: []string{"input1", "input2"},
		},
		&datakit.InputHostList{
			Hosts:  []string{"hostx", "hosty"},
			Inputs: []string{"inputx", "inputy"},
		},
	}

	blists := []*datakit.InputHostList{
		&datakit.InputHostList{
			Hosts:  []string{"host_3", "host_4"},
			Inputs: []string{"input_3", "input_4"},
		},
		&datakit.InputHostList{
			Hosts:  []string{"host_i", "host_j"},
			Inputs: []string{"input_i", "input_j"},
		},
	}

	t.Logf("host1.inputx on? %v", !isDisabled(wlists, blists, "host1", "inputx"))
	t.Logf("host2.inputy on? %v", !isDisabled(wlists, blists, "host2", "inputy"))
	t.Logf("host2.input1 on? %v", !isDisabled(wlists, blists, "host2", "input1"))
	t.Logf("host2.input_foo on? %v", !isDisabled(wlists, blists, "host2", "input_foo"))
	t.Logf("host_bar.input_foo on? %v", !isDisabled(wlists, blists, "host_bar", "input_foo"))

	t.Logf("host_3.input_foo on? %v", !isDisabled(wlists, blists, "host_3", "input_foo"))
	t.Logf("host_3.input_4 on? %v", !isDisabled(wlists, blists, "host_3", "input_4"))
	t.Logf("host_3.input_j on? %v", !isDisabled(wlists, blists, "host_3", "input_j"))
}
