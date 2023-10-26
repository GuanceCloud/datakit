//go:build !linux
// +build !linux

// Package conntrack place probes on kernel functions
// `__nf_conntrack_hash_insert` and `nf_ct_delete`
package conntrack
