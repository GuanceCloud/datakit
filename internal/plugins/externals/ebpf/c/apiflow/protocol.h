#ifndef __PROTOCOL_H__
#define __PROTOCOL_H__

#include "bpf_helpers.h"
#include "l7_utils.h"

typedef enum app_protocol
{

} app_proto_t;

typedef struct protocol_detect_buffer
{
    __u8 data[32];
} proto_dect_buf_t;


static __always_inline void read_dect_buf();

static __always_inline app_proto_t detect_protocol(proto_dect_buf_t *buf);

static __always_inline int parse_proto_buf(proto_buffer_t *buf);


static __always_inline int apiflow_gather(void *ctx);
#endif