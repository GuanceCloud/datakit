// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func Test_aesEncrypt(t *testing.T) {
	key := []byte("0123456789abcdef")
	plaintext := "HelloAES9*&."

	aesEncrypted, err := AESEncrypt(key, plaintext)
	if err != nil {
		t.Errorf("aes encrypt err=%v", err)
		return
	}

	t.Logf("AES Encrypted: %s\n", aesEncrypted)

	aesDecrypted, err := AESDecrypt(key, aesEncrypted)
	if err != nil {
		t.Errorf("aes decrypt err=%v", err)
		return
	}
	t.Logf("AES Decrypted: %s\n", aesDecrypted)
	assert.Equal(t, plaintext, aesDecrypted)
}

func TestReadFromFile(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "read hosts",
			args: args{filepath: "/etc/hosts"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS == datakit.OSLinux {
				assert.NotEmpty(t, readFromFile(tt.args.filepath), "ReadFromFile(%v)", tt.args.filepath)
			}
		})
	}
}

func TestMaskStringAfterFifth(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "len5",
			args: args{s: "abcde"},
			want: "abcde",
		},
		{
			name: "len10",
			args: args{s: "abcdefghij"},
			want: "abcde*****",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, maskPassword(tt.args.s), "MaskStringAfterFifth(%v)", tt.args.s)
		})
	}
}
