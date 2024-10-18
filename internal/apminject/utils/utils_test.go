// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package utils

import (
	"debug/elf"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	musl125 = `musl libc (x86_64)
Version 1.2.5
Dynamic Program Loader
Usage: /lib/ld-musl-x86_64.so.1 [options] [--] pathname`

	glibc235 = `ldd (Ubuntu GLIBC 2.35-0ubuntu3.8) 2.35
Copyright (C) 2022 Free Software Foundation, Inc.
This is free software; see the source for copying conditions.  There is NO
warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
Written by Roland McGrath and Ulrich Drepper.`

	glic212 = `ldd (GNU libc) 2.12
Copyright (C) 2010 Free Software Foundation, Inc.
This is free software; see the source for copying conditions.  There is NO
warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
Written by Roland McGrath and Ulrich Drepper.`

	glibc236 = `ldd (Debian GLIBC 2.36-9+deb12u8) 2.36
Copyright (C) 2022 Free Software Foundation, Inc.
This is free software; see the source for copying conditions.  There is NO
warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
Written by Roland McGrath and Ulrich Drepper.`
)

func TestUtils(t *testing.T) {
	v1, v2, ok := libcInfo(glibc235)
	assert.Equal(t, v1, glibc)
	assert.Equal(t, v2, "2.35")
	assert.True(t, ok)

	v1, v2, ok = libcInfo(glic212)
	assert.Equal(t, v1, glibc)
	assert.Equal(t, v2, "2.12")
	assert.True(t, ok)

	v1, v2, ok = libcInfo(glibc236)
	assert.Equal(t, v1, glibc)
	assert.Equal(t, v2, "2.36")
	assert.True(t, ok)

	v1, v2, ok = libcInfo(musl125)
	assert.Equal(t, v1, muslc)
	assert.Equal(t, v2, "1.2.5")
	assert.True(t, ok)
}

func TestLddInfo(t *testing.T) {
	syms := []elf.Symbol{
		{Library: "libc.so.6", Version: "GLIBC_2.35"},
		{Library: "ld-linux-x86-64.so.2", Version: "GLIBC_2.35"},
	}
	v, err := requiredGLIBCVersion(syms)
	assert.NoError(t, err)
	assert.Equal(t, Version{2, 35, 0}, v)
}
