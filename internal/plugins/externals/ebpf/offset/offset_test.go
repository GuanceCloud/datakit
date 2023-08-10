//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package offset

// const offsetStr = `{"offset_sk_num":14,"offset_inet_sport":784,"offset_sk_family":16,"offset_sk_rcv_saddr":4,"offset_sk_daddr":0,"offset_sk_v6_rcv_saddr":72,"offset_sk_v6_daddr":56,"offset_sk_dport":12,"offset_tcp_sk_srtt_us":1600,"offset_tcp_sk_mdev_us":1604,"offset_flowi4_saddr":48,"offset_flowi4_daddr":52,"offset_flowi4_sport":58,"offset_flowi4_dport":56,"offset_flowi6_saddr":64,"offset_flowi6_daddr":48,"offset_flowi6_sport":86,"offset_flowi6_dport":84,"offset_skaddr_sin_port":0,"offset_skaddr6_sin6_port":0,"offset_sk_net":48,"offset_ns_common_inum":136,"offset_socket_sk":24}`
//
// func TestGuess(t *testing.T) {
// o, err := LoadOffset(offsetStr)
// if err != nil {
// t.Error(err)
// }
//
// e := NewConstEditor(&o)
// editor, _, err := GuessOffsetTCPSeq(e)
// if err != nil {
// t.Fatal(err)
// }
//
// t.Log(editor)
// }
//
