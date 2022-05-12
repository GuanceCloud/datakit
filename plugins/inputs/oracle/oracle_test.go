// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

/* test: failed
func TestRun(t *testing.T) {
	arr, err := config.LoadInputConfigFile("./oracle.conf", func() inputs.Input {
		return &Input{}
	})
	if err != nil {
		t.Fatalf("%s", err)
	}

	o, ok := arr[0].(*Input)
	if !ok {
		t.Error("expect *Input")
	}

	t.Log("args ====>", o.Args)
	o.Run()
} */
