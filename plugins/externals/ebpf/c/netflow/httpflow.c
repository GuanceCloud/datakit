#include "bpf_helpers.h"
#include "bpfmap_l7.h"
#include "conn_stats.h"
#include "l7_stats.h"
#include "l7_utils.h"

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

    // TODO 判断本机是客户端还是服务端
    if (l7http.req_status == HTTP_REQ_REQ)
    { // request
        record_http_req(&conn_info, &stats, l7http.method);
    }
    else if (l7http.req_status == HTTP_REQ_RESP)
    {
        swap_conn_src_dst(&conn_info); // src -> client; dst -> server
        record_http_resp(&conn_info, &stats, &l7http);
    }
    return DROPPACKET;
}

SEC("kretprobe/tcp_sendmsg")
int kretprobe__tcp_sendmsg(struct pt_regs *ctx)
{
    send_httpreq_fin_event(ctx);
    return 0;
}

SEC("uprobe/SSL_set_fd")
int uprobe__SSL_set_fd(struct pt_regs *ctx)
{
    void *ssl_ctx = (void *)PT_REGS_PARM1(ctx);
    __u32 fd = (__u32)PT_REGS_PARM2(ctx);

    init_ssl_sockfd(ssl_ctx, fd);

    return 0;
}

SEC("uprobe/BIO_new_socket")
int uprobe__BIO_new_socket(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 fd = PT_REGS_PARM1(ctx);
    bpf_map_update_elem(&bpf_map_bio_new_socket_args, &pid_tgid, &fd, BPF_ANY);
    return 0;
}

SEC("uretprobe/BIO_new_socket")
int uretprobe__BIO_new_socket(struct pt_regs *ctx)
{
    __u64 pid_tgid = (__u64)bpf_get_current_pid_tgid();
    __u32 *fd_map_value = (__u32 *)bpf_map_lookup_elem(&bpf_map_bio_new_socket_args, &pid_tgid);
    if (fd_map_value == NULL)
    {
        goto cleanup;
    }

    void *bio = (void *)PT_REGS_RC(ctx);
    if (bio == NULL)
    {
        goto cleanup;
    }

    __u32 fd = *fd_map_value;
    bpf_map_update_elem(&bpf_map_ssl_bio_fd, &bio, &fd, BPF_ANY);

cleanup:
    bpf_map_delete_elem(&bpf_map_bio_new_socket_args, &pid_tgid);
    return 0;
}

SEC("uprobe/SSL_set_bio")
int uprobe__SSL_set_bio(struct pt_regs *ctx)
{
    void *ssl_ctx = (void *)PT_REGS_PARM1(ctx);
    void *bio = (void *)PT_REGS_PARM2(ctx);

    __u32 *fd = bpf_map_lookup_elem(&bpf_map_ssl_bio_fd, &bio);
    if (fd == NULL)
    {
        goto cleanup;
    }
    init_ssl_sockfd(ssl_ctx, *fd);

cleanup:
    bpf_map_delete_elem(&bpf_map_ssl_bio_fd, &bio);
    return 0;
}

SEC("uprobe/SSL_read")
int uprobe__SSL_read(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    struct ssl_read_args args = {0};
    args.ctx = (void *)PT_REGS_PARM1(ctx);
    args.buf = (void *)PT_REGS_PARM2(ctx);
    bpf_map_update_elem(&bpfmap_ssl_read_args, &pid_tgid, &args, BPF_ANY);
    return 0;
}

SEC("uretprobe/SSL_read")
int uretprobe__SSL_read(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    struct ssl_read_args *args = bpf_map_lookup_elem(&bpfmap_ssl_read_args, &pid_tgid);
    if (args == NULL)
    {
        goto cleanup;
    }

    struct http_stats stats = {0};

    bpf_probe_read(stats.payload, sizeof(__u8) * HTTP_PAYLOAD_MAXSIZE, args->buf);

    struct connection_info *conn = (struct connection_info *)read_conn_ssl(args->ctx, pid_tgid);
    if (conn == NULL)
    {
        return 0;
    }

    struct layer7_http l7http = {0};
    if (parse_layer7_http(stats.payload, &l7http) != 0)
    {

        goto cleanup;
    }

    switch (l7http.req_status)
    {
    case HTTP_REQ_REQ:
        // read is server
        swap_conn_src_dst(conn); // src -> client; dst -> server

        record_http_req(conn, &stats, l7http.method);
        // bpf_printk("ssl read: req: src_ip_port:%x:%d", conn->saddr[3], conn->sport);
        // bpf_printk("dst_ip_port:%x:%d", conn->daddr[3], conn->dport);
        break;
    case HTTP_REQ_RESP:
        record_http_resp(conn, &stats, &l7http);
        // bpf_printk("ssl read: resp: src_ip_port:%x:%d", conn->saddr[3], conn->sport);
        // bpf_printk("dst_ip_port:%x:%d", conn->daddr[3], conn->dport);
        break;
    default:
        break;
    }

cleanup:
    bpf_map_delete_elem(&bpfmap_ssl_read_args, &pid_tgid);
    return 0;
}

SEC("uprobe/SSL_write")
int uprobe__SSL_write(struct pt_regs *ctx)
{
    __u64 pid_tgid = bpf_get_current_pid_tgid();

    void *write_buf = (void *)PT_REGS_PARM2(ctx);
    void *ssl_ctx = (void *)PT_REGS_PARM1(ctx);

    struct http_stats stats = {0};
    bpf_probe_read(stats.payload, sizeof(__u8) * HTTP_PAYLOAD_MAXSIZE, write_buf);

    struct connection_info *conn = read_conn_ssl(ssl_ctx, pid_tgid);
    if (conn == NULL)
    {
        return 0;
    }

    struct layer7_http l7http = {0};
    if (parse_layer7_http(stats.payload, &l7http) != 0)
    {
        return 0;
    }

    switch (l7http.req_status)
    {
    case HTTP_REQ_REQ:
        record_http_req(conn, &stats, l7http.method);
        // bpf_printk("ssl write: req: src_ip_port:%x:%d", conn->saddr[3], conn->sport);
        // bpf_printk("dst_ip_port:%x:%d", conn->daddr[3], conn->dport);
        break;
    case HTTP_REQ_RESP:
        swap_conn_src_dst(conn); // src -> client; dst -> server

        record_http_resp(conn, &stats, &l7http);
        // bpf_printk("ssl write: resp: src_ip_port:%x:%d", conn->saddr[3], conn->sport);
        // bpf_printk("dst_ip_port:%x:%d", conn->daddr[3], conn->dport);
        break;
    default:
        break;
    }
    return 0;
}

SEC("uprobe/SSL_shutdown")
int uprobe__SSL_shutdown(struct pt_regs *ctx)
{
    void *ssl_ctx = (void *)PT_REGS_PARM1(ctx);
    bpf_map_delete_elem(&bpfmap_ssl_ctx_sockfd, &ssl_ctx);
    return 0;
}

char _license[] SEC("license") = "GPL";
// this number will be interpreted by eBPF(Cilium) elf-loader
// to set the current running kernel version
__u32 _version SEC("version") = 0xFFFFFFFE;
