#define _GNU_SOURCE
#include <stdbool.h>
#include <stdint.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/syscall.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/wait.h>
#include <signal.h>
#include <fcntl.h>
#include <errno.h>

#ifdef DATAKIT_DEBUG
#define debug_perror(msg) perror(msg)
#define debug_log(msg, ...) fprintf(stderr, "%s:%d: %s: " msg, \
                                    __FILE__, __LINE__, __func__, ##__VA_ARGS__)
#else
#define debug_perror(msg)
#define debug_log(msg, ...)
#endif

#ifndef DATAKIT_INJ_RAND_BYTES
#define DATAKIT_INJ_RAND_BYTES 16
#endif

#ifndef DATAKIT_INJ_FILENAME_PREFIX
#define DATAKIT_INJ_FILENAME_PREFIX "/tmp/dk_inject_rewrite_"
#endif

#ifndef DATAKIT_INJ_REWRITE_PROC
#define DATAKIT_INJ_REWRITE_PROC "/usr/local/datakit/apm_inject/inject/rewriter"
#endif

#ifdef __GLIBC__
#define SYM_VER_GLIBC(sym, ver) __asm__(".symver " #sym "," #sym "@GLIBC_" #ver)
#if defined(__x86_64__)
#define SYM_VER_GLIBC2_2_5(sym) SYM_VER_GLIBC(sym, 2.2.5)
SYM_VER_GLIBC2_2_5(atoi);
SYM_VER_GLIBC2_2_5(strlen);
SYM_VER_GLIBC2_2_5(strdup);
SYM_VER_GLIBC2_2_5(strcat);
SYM_VER_GLIBC2_2_5(getdelim);
#undef SYM_VER_GLIBC2_2_5
#elif defined(__aarch64__)
#define SYM_VER_GLIBC2_17(sym) SYM_VER_GLIBC(sym, 2.17)
SYM_VER_GLIBC2_17(atoi);
SYM_VER_GLIBC2_17(strlen);
SYM_VER_GLIBC2_17(strdup);
SYM_VER_GLIBC2_17(strcat);
SYM_VER_GLIBC2_17(getdelim);
#undef SYM_VER_GLIBC2_17
#endif
#undef SYM_VER_GLIBC
#endif

static const char launcher_version[] = "1.0";

static inline int old_execve(char const *path, char *const *argv, char *const *envp)
{
    return syscall(SYS_execve, path, argv, envp);
}

static ssize_t read_rewrite_record(char **line, size_t *n, FILE *fp)
{
    return getdelim(line, n, 0x1E, fp);
}

static char *const *read_args(FILE *fp)
{
    char **arr = NULL;
    int total_lines = 0;

    // total_lines<RS>arg_1<RS>...<RS>arg_n<RS><GS><RS>
    for (int ln = 0;; ln++)
    {
        size_t len = 0;
        char *line = NULL;
        // 0x1E, record separator
        ssize_t nread = read_rewrite_record(&line, &len, fp);
        if (nread == -1)
        {
            if (line != NULL)
            {
                free(line);
            }
            goto err_check;
        }
        if (nread == 2)
        {
            // 0x1D, group separator
            if (line[0] == 0x1D && line[1] == 0x1E)
            {
                free(line);
                break;
            }
        }

        if (ln == 0)
        {
            line[nread - 1] = '\0';
            total_lines = atoi(line);
            free(line);

            size_t nsize = (total_lines + 1) * sizeof(char *);
            if (arr = malloc(nsize), arr == NULL)
            {
                debug_perror("malloc rewrite record");
                return NULL;
            }
            __builtin_memset(arr, 0, nsize);
            continue;
        }

        if (ln > total_lines)
        {
            free(line);
            return NULL; // error, need check content format
        }
        line[nread - 1] = '\0';
        arr[ln - 1] = line;
    }

    return arr;

err_check:
    if (feof(fp))
    {
        return arr;
    }
    else
    {
        debug_perror("read rewrite record");
        return NULL;
    }
}

static void multi_args_free(char *const *ptr[], int len)
{
    if (ptr == NULL)
    {
        return;
    }
    for (int i = 0; i < len; i++)
    {
        if (ptr[i] == NULL)
        {
            continue;
        }
        for (int j = 0; ptr[i][j] != NULL; j++)
        {
            free(ptr[i][j]);
        }
        free((void *)ptr[i]);
    }
    return;
}

static bool gen_inject_tmpid_b16(char *tmpid)
{
    uint8_t buf[DATAKIT_INJ_RAND_BYTES] = {0};
    int fd = open("/dev/urandom", O_RDONLY);
    if (fd < 0)
    {
        return false;
    }
    read(fd, buf, DATAKIT_INJ_RAND_BYTES);
    close(fd);

    for (int i = 0; i < DATAKIT_INJ_RAND_BYTES; i++)
    {
        uint8_t c = buf[i];
        tmpid[(i * 2)] = (c & 0xf);
        tmpid[(i * 2) + 1] = (c >> 4);
    }
    for (int i = 0; i < DATAKIT_INJ_RAND_BYTES * 2; i++)
    {
        if (tmpid[i] > 9)
        {
            tmpid[i] += ('A' - 10);
        }
        else
        {
            tmpid[i] += '0';
        }
    }

    return true;
}

static bool try_apm_inject_process(char *const argv[], char *const envp[])
{
    pid_t sub_process = fork();

    switch (sub_process)
    {
    case -1:
        debug_perror("fork error");
        return false;
    case 0: // child process
        ;
        // a label can only be part of a statement
        // and a declaration is not a statement
        pid_t child_rewriter = fork();
        switch (child_rewriter)
        {
        case -1:
            debug_perror("fork rewrite process");
            _exit(1);
        case 0: // grandson process 1
            ;
            int ret = old_execve(DATAKIT_INJ_REWRITE_PROC, argv, envp);
            if (ret != 0)
            {
                debug_perror("run rewrite process `" DATAKIT_INJ_REWRITE_PROC "`");
            }
            _exit(0);
        }

        pid_t child_watchdog = fork();
        switch (child_watchdog)
        {
        case -1:
            debug_perror("fork watchdog timer");
            kill(child_rewriter, SIGKILL);
            wait(0);
            _exit(1);
        case 0: // grandson process 2
            // software watchdog timer
            sleep(1);
            _exit(0);
        }

        pid_t child_exited = wait(0); // skip check `errno`

        int stat = 0;
        if (child_exited == child_watchdog)
        {
            // timeout
            debug_log("rewrite process timeout\n");
            kill(child_rewriter, SIGKILL);
            stat = 1;
        }
        else
        {
            kill(child_watchdog, SIGKILL);
        }

        wait(0);
        _exit(stat);
    default:
        break;
    }

    // parent process
    int stat = 0;
    waitpid(sub_process, &stat, 0);

    return WEXITSTATUS(stat) == 0 ? true : false;
}

static int varb_array_len(char *const arr[])
{
    if (arr == NULL)
    {
        return 0;
    }
    int len = 0;
    for (int i = 0; arr[i] != NULL; i++)
    {
        len++;
    }
    return len;
}

int execve(const char *path, char *const argv[], char *const envp[])
{
    char tmpid[DATAKIT_INJ_RAND_BYTES * 2 + 1] = {0};
    // retrieve modified launch parameters from file

    if (!gen_inject_tmpid_b16(tmpid))
    {
        debug_log("gen tmpid failed\n");
        return old_execve(path, argv, envp);
    }

    int count_argv = varb_array_len(argv);
    char **dup_argv = malloc((count_argv + 2 + 1) * sizeof(char const *));
    dup_argv[0] = strdup(tmpid);
    dup_argv[1] = strdup(path);
    dup_argv[count_argv + 2] = NULL;

    for (int i = 0; i < count_argv; i++)
    {
        dup_argv[i + 2] = strdup(argv[i]);
    }

    // MT-Unsafe
    bool ok = try_apm_inject_process(dup_argv, envp);

    for (int i = 0; i < count_argv + 2; i++)
    {
        free((void *)dup_argv[i]);
    }
    free(dup_argv);

    if (!ok)
    {
        return old_execve(path, argv, envp);
    }

    char rewrite_args_fname[sizeof(DATAKIT_INJ_FILENAME_PREFIX) + sizeof(tmpid)] = {0};
    strcat(rewrite_args_fname, DATAKIT_INJ_FILENAME_PREFIX);
    strcat(rewrite_args_fname, tmpid);

    FILE *fp = fopen(rewrite_args_fname, "r");
    if (fp == NULL)
    {
        debug_perror(rewrite_args_fname);
        return old_execve(path, argv, envp);
    }

    char *const *multi_args[3] = {0};
    for (int i = 0; i < 3; i++)
    {
        char *const *val = read_args(fp);
        if ((val == NULL) ||
            (i == 0 && val[0] == NULL))
        {
            multi_args_free(multi_args, i);
            fclose(fp);
#ifdef DATAKIT_DEBUG
#else
            remove(rewrite_args_fname);
#endif
            return old_execve(path, argv, envp);
        }
        multi_args[i] = val;
    }
    fclose(fp);
#ifdef DATAKIT_DEBUG
#else
    remove(rewrite_args_fname);
#endif

    debug_log("rewrite args: %s %s %s\n", multi_args[0][0],
              multi_args[1][0], multi_args[2][0]);

    int ret = old_execve(multi_args[0][0], multi_args[1], multi_args[2]);
    multi_args_free(multi_args, 3);
    return ret;
}
