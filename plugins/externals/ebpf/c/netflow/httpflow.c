#include "bpf_helpers.h"
#include "bpfmap_l7.h"
#include "conn_stats.h"
#include "l7_stats.h"
#include "l7_utils.h"

#define KEEPPACKET -1
#define DROPPACKET 0

SEC("socket/http_filter")
int socket__http_filter(struct __sk_buff *skb)
{

    struct conn_skb_l4_info skbl4_info = {0};
    struct http_stats stats = {0};
    struct connection_info conn_info = {0};

    if (read_connection_info_skb(skb, &skbl4_info, &conn_info) != 0)
    {
        return DROPPACKET;
    }

    if ((conn_info.meta & CONN_L4_MASK) != CONN_L4_TCP)
    {
        return DROPPACKET;
    }

    // 跳过 https netflow
    // TODO 解析 7 层 payload 获取 TLS/SSL 版本信息
    if (conn_info.sport == 443 || conn_info.dport == 443)
    {
        return DROPPACKET;
    }

#pragma unroll
    // 使用 load_byte 时当 HTTP_PAYLOAD_MAXSIZE 值超出一定长度后分片读取异常，
    // 可出现在使用代理程序时 TCP 分片需重组的情况下读取 response 数据异常；
    // 使用 bpf_skb_load_bytes 能获取一部分数据
    // TODO
    for (int i = 0; i < HTTP_PAYLOAD_MAXSIZE - 1; i++) // arr[HTTP_PAYLOAD_MAXSIZE - 1] == EOF
    {
        // stats.payload[i] = load_byte(skb, skbl4_info.hdr_len + i);
        bpf_skb_load_bytes(skb, skbl4_info.hdr_len + i, stats.payload + i, sizeof(__u8));
    }

    struct layer7_http l7http = {0};

    if (parse_layer7_http(stats.payload, &l7http) != 0)
    {
        return DROPPACKET;
    }

    // 会出现两份一样的数据，本机同时是客户端和服务端
    // TODO 判断本机是客户端还是服务端
    if (l7http.req_status == HTTP_REQ_REQ)
    { // request
        stats.req_ts = bpf_ktime_get_ns();
        stats.req_method = l7http.method;
        bpf_map_update_elem(&bpfmap_http_stats, &conn_info, &stats, BPF_NOEXIST);

        bpf_printk("xxsrc, dst %lx %lx", conn_info.saddr[3], conn_info.daddr[3]);
        bpf_printk("xxsport, dport %d %d", conn_info.sport, conn_info.dport);
        bpf_printk("%s %d", stats.payload, bpf_get_smp_processor_id());
    }
    else if (l7http.req_status == HTTP_REQ_RESP)
    {
        // response
        swap_conn_src_dst(&conn_info); // src -> client; dst -> server
        struct http_stats *stats_cached = bpf_map_lookup_elem(&bpfmap_http_stats, &conn_info);
        if (stats_cached == NULL)
        {
            return DROPPACKET;
        }
        // bpf_printk("src, dst %lx %lx", conn_info.saddr[3], conn_info.daddr[3]);
        // bpf_printk("sport, dport %d %d", conn_info.sport, conn_info.dport);
        // bpf_printk("%s %d", stats.resp_code, bpf_get_smp_processor_id());

        struct http_req_finished_info http_finished = {0};

        __builtin_memcpy(&http_finished.conn_info, &conn_info, sizeof(struct connection_info));
        __builtin_memcpy(&http_finished.http_stats, stats_cached, sizeof(struct http_stats));

        bpf_map_delete_elem(&bpfmap_http_stats, &conn_info);

        http_finished.http_stats.resp_code = l7http.status_code;
        http_finished.http_stats.resp_ts = bpf_ktime_get_ns();

        // 需要较新的内核支持，
        // https://github.com/torvalds/linux/commit/7c4b90d79d0f4ee6f2c01a114ef0a83a09dfc900
        // bpf_perf_event_output(ctx, map_ptr, cpu, &http_finished, sizeof(http_finished));
        // unknown func bpf_skb_output#111
        // __u64 cpu  = bpf_get_smp_processor_id();
        // bpf_skb_output(skb, &bpfmap_httpreq_fin_event, cpu,  &http_finished, sizeof(http_finished));
        map_cache_finished_http_req(&http_finished);
    }
    return DROPPACKET;
}

SEC("kretprobe/tcp_sendmsg")
int kretprobe__tcp_sendmsg(struct pt_regs *ctx)
{
    send_httpreq_fin_event(ctx);
    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
