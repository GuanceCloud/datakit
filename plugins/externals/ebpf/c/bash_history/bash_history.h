#ifndef __BASH_READLINE_H
#define __BASH_READLINE_H

#include <linux/types.h>
struct bash_event
{
    __u64 pid_tgid;
    __u64 uid_gid;
    __u8 line[128];
};

#endif // __BASH_READLINE_H
