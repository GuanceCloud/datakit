// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package filecollector

import (
	"testing"
)

func TestFsnotify(t *testing.T) {
	/*
		dir, _ := os.Getwd()

		watch, err := fsnotify.NewWatcher()
		if err != nil {
			l.Fatal(err)
		}
		err = watch.Add(dir)
		if err != nil {
			l.Fatal(err)
		}

		go func() {
			time.Sleep(time.Second * 1)

			f, err := os.Create(filepath.Join(dir, "123.txt"))
			if err != nil {
				l.Fatal(err)
			}
			f.Close() //nolint:errcheck
		}()

		for ev := range watch.Events {
			t.Log(ev.String())
		}

		_ = os.Remove(filepath.Join(dir, "123.txt")) //nolint:errcheck
	*/
}
