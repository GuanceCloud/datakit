/* SPDX-License-Identifier: (LGPL-2.1 OR BSD-2-Clause) */
#ifndef __BPF_HELPERS__
#define __BPF_HELPERS__

#include <linux/version.h>
#include <uapi/linux/bpf.h>
#include "asm_goto_workaround.h"

#define __uint(name, val) int(*name)[val]
#define __type(name, val) typeof(val) *name

// #ifdef __DK_DEBUG__
/* helper macro to print out debug messages */
#define bpf_printk(fmt, ...)                       \
	({                                             \
		char ____fmt[] = fmt;                      \
		bpf_trace_printk(____fmt, sizeof(____fmt), \
						 ##__VA_ARGS__);           \
	})
// #else
// #define bpf_printk(fmt, ...)
// #endif

#ifdef __clang__

/* helper macro to place programs, maps, license in
 * different sections in elf_bpf file. Section names
 * are interpreted by elf_bpf loader
 */
#define SEC(NAME) __attribute__((section(NAME), used))

/* helper functions called from eBPF programs written in C */
static void *(*bpf_map_lookup_elem)(void *map, const void *key) =
	(void *)BPF_FUNC_map_lookup_elem;
static int (*bpf_map_update_elem)(void *map, const void *key, const void *value,
								  unsigned long long flags) =
	(void *)BPF_FUNC_map_update_elem;
static int (*bpf_map_delete_elem)(void *map, const void *key) =
	(void *)BPF_FUNC_map_delete_elem;
static int (*bpf_probe_read)(void *dst, int size, const void *unsafe_ptr) =
	(void *)BPF_FUNC_probe_read;
static unsigned long long (*bpf_ktime_get_ns)(void) =
	(void *)BPF_FUNC_ktime_get_ns;
static int (*bpf_trace_printk)(const char *fmt, int fmt_size, ...) =
	(void *)BPF_FUNC_trace_printk;
static void (*bpf_tail_call)(void *ctx, void *map, int index) =
	(void *)BPF_FUNC_tail_call;
static unsigned long long (*bpf_get_smp_processor_id)(void) =
	(void *)BPF_FUNC_get_smp_processor_id;
static unsigned long long (*bpf_get_current_pid_tgid)(void) =
	(void *)BPF_FUNC_get_current_pid_tgid;
static unsigned long long (*bpf_get_current_uid_gid)(void) =
	(void *)BPF_FUNC_get_current_uid_gid;
static int (*bpf_get_current_comm)(void *buf, int buf_size) =
	(void *)BPF_FUNC_get_current_comm;
static void *(*bpf_get_current_task)(void) =
	(void *)BPF_FUNC_get_current_task;
static unsigned long long (*bpf_perf_event_read)(void *map,
												 unsigned long long flags) =
	(void *)BPF_FUNC_perf_event_read;
static int (*bpf_clone_redirect)(void *ctx, int ifindex, int flags) =
	(void *)BPF_FUNC_clone_redirect;
static int (*bpf_redirect)(int ifindex, int flags) =
	(void *)BPF_FUNC_redirect;
static int (*bpf_redirect_map)(void *map, int key, int flags) =
	(void *)BPF_FUNC_redirect_map;
static int (*bpf_perf_event_output)(void *ctx, void *map,
									unsigned long long flags, void *data,
									int size) =
	(void *)BPF_FUNC_perf_event_output;
static int (*bpf_get_stackid)(void *ctx, void *map, int flags) =
	(void *)BPF_FUNC_get_stackid;
static int (*bpf_probe_write_user)(void *dst, const void *src, int size) =
	(void *)BPF_FUNC_probe_write_user;
static unsigned long long (*bpf_get_prandom_u32)(void) =
	(void *)BPF_FUNC_get_prandom_u32;

/* llvm builtin functions that eBPF C program may use to
 * emit BPF_LD_ABS and BPF_LD_IND instructions
 */
struct sk_buff;
unsigned long long load_byte(void *skb,
							 unsigned long long off) asm("llvm.bpf.load.byte");
unsigned long long load_half(void *skb,
							 unsigned long long off) asm("llvm.bpf.load.half");
unsigned long long load_word(void *skb,
							 unsigned long long off) asm("llvm.bpf.load.word");

/* a helper structure used by eBPF C program
 * to describe map attributes to elf_bpf loader
 */
#define BUF_SIZE_MAP_NS 256

struct bpf_map_def
{
	unsigned int type;
	unsigned int key_size;
	unsigned int value_size;
	unsigned int max_entries;
	unsigned int map_flags;
	unsigned int pinning;
	char namespace[BUF_SIZE_MAP_NS];
};

#define BPF_HASH_MAP(map_name, key_type, value_type, map_max_entries) \
	struct bpf_map_def SEC("maps/" #map_name) map_name = {            \
		.type = BPF_MAP_TYPE_HASH,                                    \
		.key_size = sizeof(key_type),                                 \
		.value_size = sizeof(value_type),                             \
		.max_entries = map_max_entries};

#define BPF_PERF_EVENT_MAP(map_name)                       \
	struct bpf_map_def SEC("maps/" #map_name) map_name = { \
		.type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,             \
		.key_size = sizeof(__u32),                         \
		.value_size = sizeof(__u32),                       \
		.max_entries = 0};

static int (*bpf_skb_load_bytes)(void *ctx, int off, void *to, int len) =
	(void *)BPF_FUNC_skb_load_bytes;
static int (*bpf_skb_store_bytes)(void *ctx, int off, void *from, int len, int flags) =
	(void *)BPF_FUNC_skb_store_bytes;
static int (*bpf_l3_csum_replace)(void *ctx, int off, int from, int to, int flags) =
	(void *)BPF_FUNC_l3_csum_replace;
static int (*bpf_l4_csum_replace)(void *ctx, int off, int from, int to, int flags) =
	(void *)BPF_FUNC_l4_csum_replace;

#define PT_REGS_STACK_PARM(x, n)                                     \
	({                                                               \
		unsigned long p = 0;                                         \
		bpf_probe_read(&p, sizeof(p), ((unsigned long *)x->sp) + n); \
		p;                                                           \
	})

#if defined(__x86_64__)

// https://go.googlesource.com/go/+/refs/heads/master/src/cmd/compile/abi-internal.md#amd64-architecture
#define PT_GO_REGS_PARAM1(x) ((x)->ax)
#define PT_GO_REGS_PARAM2(x) ((x)->bx)
#define PT_GO_REGS_PARAM3(x) ((x)->cx)
#define PT_GO_REGS_PARAM4(x) ((x)->di)
#define PT_GO_REGS_PARAM5(x) ((x)->si)
#define PT_GO_REGS_PARAM6(x) ((x)->r8)
#define PT_GO_REGS_PARAM7(x) ((x)->r9)
#define PT_GO_REGS_PARAM8(x) ((x)->r10)
#define PT_GO_REGS_PARAM9(x) ((x)->r11)
#define PT_GO_REGS_GOROUTINE(x) ((x)->r14)

#define PT_REGS_PARM1(x) ((x)->di)
#define PT_REGS_PARM2(x) ((x)->si)
#define PT_REGS_PARM3(x) ((x)->dx)
#define PT_REGS_PARM4(x) ((x)->cx)
#define PT_REGS_PARM5(x) ((x)->r8)
#define PT_REGS_PARM6(x) ((x)->r9)
#define PT_REGS_PARM7(x) PT_REGS_STACK_PARM(x, 1)
#define PT_REGS_PARM8(x) PT_REGS_STACK_PARM(x, 2)
#define PT_REGS_PARM9(x) PT_REGS_STACK_PARM(x, 3)
#define PT_REGS_RET(x) ((x)->sp)
#define PT_REGS_FP(x) ((x)->bp)
#define PT_REGS_RC(x) ((x)->ax)
#define PT_REGS_SP(x) ((x)->sp)
#define PT_REGS_IP(x) ((x)->ip)

#elif defined(__aarch64__)

// https://go.googlesource.com/go/+/refs/heads/master/src/cmd/compile/abi-internal.md#arm64-architecture
#define PT_GO_REGS_PARAM1(x) ((x))->regs[0])
#define PT_GO_REGS_PARAM2(x) ((x))->regs[1])
#define PT_GO_REGS_PARAM3(x) ((x))->regs[2])
#define PT_GO_REGS_PARAM4(x) ((x))->regs[3])
#define PT_GO_REGS_PARAM5(x) ((x))->regs[4])
#define PT_GO_REGS_PARAM6(x) ((x))->regs[5])
#define PT_GO_REGS_PARAM7(x) ((x))->regs[6])
#define PT_GO_REGS_PARAM8(x) ((x))->regs[7])
#define PT_GO_REGS_PARAM9(x) ((x))->regs[8])
#define PT_GO_REGS_GOROUTINE(x) ((x))->regs[28])

#define PT_REGS_PARM1(x) ((x)->regs[0])
#define PT_REGS_PARM2(x) ((x)->regs[1])
#define PT_REGS_PARM3(x) ((x)->regs[2])
#define PT_REGS_PARM4(x) ((x)->regs[3])
#define PT_REGS_PARM5(x) ((x)->regs[4])
#define PT_REGS_PARM6(x) ((x)->regs[5])
#define PT_REGS_PARM7(x) ((x)->regs[6])
#define PT_REGS_PARM8(x) ((x)->regs[7])
#define PT_REGS_PARM9(x) PT_REGS_STACK_PARM(x, 1)
#define PT_REGS_RET(x) ((x)->regs[30])
#define PT_REGS_FP(x) ((x)->regs[29]) /* Works only with CONFIG_FRAME_POINTER */
#define PT_REGS_RC(x) ((x)->regs[0])
#define PT_REGS_SP(x) ((x)->sp)
#define PT_REGS_IP(x) ((x)->pc)

#endif

#define BPF_KPROBE_READ_RET_IP(ip, ctx) ({ bpf_probe_read(&(ip), sizeof(ip), (void *)PT_REGS_RET(ctx)); })
#define BPF_KRETPROBE_READ_RET_IP(ip, ctx) ({ bpf_probe_read(&(ip), sizeof(ip), \
															 (void *)(PT_REGS_FP(ctx) + sizeof(ip))); })
#endif

#endif