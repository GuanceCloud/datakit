package consumerLibrary

import (
	"reflect"
	"testing"
)

func TestSet(t *testing.T) {
	type args struct {
		slc []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"TestSet_1", args{[]int{1, 1, 2, 2, 3, 3}}, []int{1, 2, 3}},
		{"TestSet_2", args{[]int{0, 1, 2, 2, 2, 3, 3, 4}}, []int{0, 1, 2, 3, 4}},
	}
	for _, tt := range tests {
		if got := Set(tt.args.slc); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Set() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSubtract(t *testing.T) {
	type args struct {
		a []int
		b []int
	}
	tests := []struct {
		name          string
		args          args
		wantDiffSlice []int
	}{
		{"TestSubtract_1", args{[]int{0, 1, 2}, []int{1, 2, 3}}, []int{3}},
		{"TestSubtract_2", args{[]int{4, 5, 9}, []int{1, 2, 3}}, []int{1, 2, 3}},
		{"TestSubtract_3", args{[]int{99, 99, 99}, []int{88, 89, 90}}, []int{88, 89, 90}},
	}
	for _, tt := range tests {
		if gotDiffSlice := Subtract(tt.args.a, tt.args.b); !reflect.DeepEqual(gotDiffSlice, tt.wantDiffSlice) {
			t.Errorf("%q. Subtract() = %v, want %v", tt.name, gotDiffSlice, tt.wantDiffSlice)
		}
	}
}

func TestMin(t *testing.T) {
	type args struct {
		a int64
		b int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{"TestMin_1", args{0, 1}, 0},
		{"TestMin_2", args{1, 1}, 1},
	}
	for _, tt := range tests {
		if got := Min(tt.args.a, tt.args.b); got != tt.want {
			t.Errorf("%q. Min() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIntSliceReflectEqual(t *testing.T) {
	type args struct {
		a []int
		b []int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TestIntSliceReflectEqual_1", args{[]int{0, 1, 2}, []int{0, 1, 2}}, true},
		{"TestIntSliceReflectEqual_2", args{[]int{11, 12, 23, 24}, []int{0, 1, 2}}, false},
	}
	for _, tt := range tests {
		if got := IntSliceReflectEqual(tt.args.a, tt.args.b); got != tt.want {
			t.Errorf("%q. IntSliceReflectEqual() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestContain(t *testing.T) {
	type args struct {
		obj    int
		target []int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"TestContain_1", args{2, []int{1, 2, 3}}, true},
		{"TestContain_2", args{2, []int{1, 4, 3}}, false},
	}
	for _, tt := range tests {
		if got := Contain(tt.args.obj, tt.args.target); got != tt.want {
			t.Errorf("%q. Contain() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
