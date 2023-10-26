#ifndef __NET_PROTOCOL_DETECT_H_
#define __NET_PROTOCOL_DETECT_H_

#include "bpf_helpers.h"

typedef bool __u8;

#define true 1
#define false 0

enum PROTO_NUM
{
    PROTO_UNKNOWN = 0x0,
    PROTO_HTTP1,
    PROTO_HTTP2,
    PROTO_REDIS,
    PROTO_MYSQL,
};

typedef struct
{
    __u8 buf[16];
    __u16 proto;
    __u16 is_req;
    __u32 _pad;
} proto_detect_data_t;

static __always_inline int detect_l7_proto(proto_detect_data_t *data)
{
}

static __always_inline bool detect_http(proto_detect_data_t *data)
{
    __u8 *buffer = data->buf;

    switch (buffer[0])
    {
    case 'G':
        if (buffer[1] == 'E' && buffer[2] == 'T' && buffer[3] == ' ') // HTTP GET
        {
            data->proto = PROTO_HTTP1;
            data->is_req = true;
            return true;
        }
        break;
    case 'P':
        switch (buffer[1])
        {
        case 'O':
            if (buffer[2] == 'S' && buffer[3] == 'T' &&
                buffer[4] == ' ') // HTTP POST
            {
                data->proto = PROTO_HTTP1;
                data->is_req = true;
                return true;
            }
            break;
        case 'U':
            if (buffer[2] == 'T' && buffer[3] == ' ') // HTTP PUT
            {
                data->proto = PROTO_HTTP1;
                data->is_req = true;
                return true;
            }
            break;
        case 'A':
            if (buffer[2] == 'T' && buffer[3] == 'C' && buffer[4] == 'H' &&
                buffer[5] == ' ') // HTTP PATCH
            {
                data->proto = PROTO_HTTP1;
                data->is_req = true;
                return true;
            }
            break;
        default:
            break;
        }
    case 'D':
        if (buffer[1] == 'E' && buffer[2] == 'L' && buffer[3] == 'E' &&
            buffer[4] == 'T' && buffer[5] == 'E' &&
            buffer[6] == ' ') // HTTP DELETE
        {
            data->proto = PROTO_HTTP1;
            data->is_req = true;
            return true;
        }
        break;
    case 'H':
        if (buffer[1] == 'T' && buffer[2] == 'T' &&
            buffer[3] == 'P') // response payload
        {
            data->proto = PROTO_HTTP1;
            data->is_req = false;
            return true;
        }
        else if (buffer[1] == 'E' && buffer[2] == 'A' && buffer[3] == 'D' &&
                 buffer[4] == ' ') // HTTP HEAD
        {
            data->proto = PROTO_HTTP1;
            data->is_req = true;
            return true;
        }
        break;
    case 'O':
        if (buffer[1] == 'P' && buffer[2] == 'T' && buffer[3] == 'I' &&
            buffer[4] == 'O' && buffer[5] == 'N' && buffer[6] == 'S' &&
            buffer[7] == ' ') // HTTP OPTIONS
        {
            data->proto = PROTO_HTTP1;
            data->is_req = true;
            return true;
        }
        break;
    // case 'C':
    //     if (buffer[1] == 'O' && buffer[2] == 'N' && buffer[3] == 'N' &&
    //         buffer[4] == 'E' && buffer[5] == 'C' && buffer[6] == 'T') // HTTP CONNECTION
    //         {
    //         }
    //     break;
    // case 'T':
    //     if (buffer[1] == 'R' && buffer[2] == 'A' && buffer[3] == 'C' && buffer[4] == 'E')
    //     { // HTTP TRACE l7http->method = HTTP_METHOD_TRACE; return
    //     }
    //     break;
    default:
        break;
    }
    return false;
}

#endif