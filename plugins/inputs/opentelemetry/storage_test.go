package opentelemetry

import "testing"

func Test_byteToInt64(t *testing.T) {
	type args struct {
		bts []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "nil", args: args{bts: []byte{}}, want: "0"},
		{name: "100", args: args{bts: []byte{1, 0, 0}}, want: "010000"},
		{name: "a1", args: args{bts: []byte{0xa1}}, want: "a1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := byteToInt64(tt.args.bts); got != tt.want {
				t.Errorf("byteToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}
