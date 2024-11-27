// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux
// +build !linux

package openfile

func FileKey(file string) string  { return file }
func UniqueID(file string) string { return file }
func Inode(_ string) string       { return "not-supported" }
func Device(_ string) string      { return "not-supported" }
