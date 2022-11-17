// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

package ipmi

import (
	"reflect"
	"testing"
)

func TestStrings2StringSlice(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name     string
		args     args
		wantStrs []string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name:     "good",
			args:     args{`["fan","slot","drive"]`},
			wantStrs: []string{"fan", "slot", "drive"},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStrs, err := Strings2StringSlice(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("Strings2StringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotStrs, tt.wantStrs) {
				t.Errorf("Strings2StringSlice() = %v, want %v", gotStrs, tt.wantStrs)
			}
		})
	}
}

func TestInts2IntSlice(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name     string
		args     args
		wantInts []int
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name:     "good",
			args:     args{`[1, 2, 3, 444 ,  0]`},
			wantInts: []int{1, 2, 3, 444, 0},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInts, err := Ints2IntSlice(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ints2IntSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotInts, tt.wantInts) {
				t.Errorf("Ints2IntSlice() = %v, want %v", gotInts, tt.wantInts)
			}
		})
	}
}
