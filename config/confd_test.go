// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"os"
	"testing"

	tu "github.com/GuanceCloud/cliutils/testutil"
)

func TestConfd(t *testing.T) {
	dirname := "test123"
	if err := os.MkdirAll(dirname, 0o600); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.RemoveAll(dirname); err != nil {
			t.Error(err)
		}
	}()

	tu.Assert(t, emptyDir(dirname) == true, "dir %s should be empty", dirname)
}

// func Test_doWithContext(t *testing.T) {
// 	timeout1s, cancel := context.WithTimeout(context.Background(), time.Second*2)
// 	_ = cancel
// 	type args struct {
// 		ctx context.Context
// 		fn  func() (interface{}, error)
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    interface{}
// 		wantErr bool
// 	}{
// 		{
// 			"ok",
// 			args{
// 				ctx: context.Background(),
// 				fn: func() (interface{}, error) {
// 					return nil, nil
// 				},
// 			},
// 			nil,
// 			false,
// 		},
// 		// {
// 		// 	"timeout",
// 		// 	args{
// 		// 		ctx: timeout1s,
// 		// 		fn: func() (interface{}, error) {
// 		// 			<-time.After(time.Second * 6)
// 		// 			return nil, nil
// 		// 		},
// 		// 	},
// 		// 	nil,
// 		// 	true,
// 		// },
// 		{
// 			"timeout",
// 			args{
// 				ctx: timeout1s,
// 				fn: func() (interface{}, error) {
// 					for i := 0; i < 20; i++ {
// 						fmt.Println("A-", i)
// 						time.Sleep(time.Microsecond * 300)
// 					}

// 					return nil, nil
// 				},
// 			},
// 			nil,
// 			true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := doWithContext(tt.args.ctx, tt.args.fn)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("doWithContext() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("doWithContext() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
