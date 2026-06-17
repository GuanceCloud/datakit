#define _GNU_SOURCE

#include <stdbool.h>
#include <stdint.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdarg.h>
#include <unistd.h>
#include <sys/syscall.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/wait.h>
#include <signal.h>
#include <spawn.h>
#include <fcntl.h>
#include <errno.h>

#if defined(__GLIBC__) && defined(__x86_64__)
#define DATAKIT_GLIBC_POSIX_SPAWN_SYMVER 1
#endif

#ifndef DATAKIT_GLIBC_POSIX_SPAWN_SYMVER
#include <dlfcn.h>
#endif

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
SYM_VER_GLIBC2_2_5(strchr);
SYM_VER_GLIBC2_2_5(getenv);
SYM_VER_GLIBC2_2_5(dlsym);
SYM_VER_GLIBC2_2_5(euidaccess);
SYM_VER_GLIBC2_2_5(getdelim);
#undef SYM_VER_GLIBC2_2_5
#elif defined(__aarch64__)
#define SYM_VER_GLIBC2_17(sym) SYM_VER_GLIBC(sym, 2.17)
SYM_VER_GLIBC2_17(atoi);
SYM_VER_GLIBC2_17(strlen);
SYM_VER_GLIBC2_17(strdup);
SYM_VER_GLIBC2_17(strcat);
SYM_VER_GLIBC2_17(strchr);
SYM_VER_GLIBC2_17(getenv);
SYM_VER_GLIBC2_17(dlsym);
SYM_VER_GLIBC2_17(euidaccess);
SYM_VER_GLIBC2_17(getdelim);
#undef SYM_VER_GLIBC2_17
#endif
#undef SYM_VER_GLIBC
#endif

static const char launcher_version[] __attribute__((used)) = "1.0";
static const char default_exec_path[] = "/bin:/usr/bin";

extern char **environ;

static inline int old_execve(char const *path, char *const *argv, char *const *envp)
{
    return syscall(SYS_execve, path, argv, envp);
}

static int effective_exec_access(const char *path)
{
#ifdef __GLIBC__
    return euidaccess(path, X_OK);
#else
    return faccessat(AT_FDCWD, path, X_OK, AT_EACCESS);
#endif
}

typedef int (*posix_spawn_func_t)(pid_t *pid, const char *path, const posix_spawn_file_actions_t *file_actions,
                                  const posix_spawnattr_t *attrp, char *const argv[], char *const envp[]);

#ifdef DATAKIT_GLIBC_POSIX_SPAWN_SYMVER
extern int old_posix_spawn(pid_t *pid, const char *path, const posix_spawn_file_actions_t *file_actions,
                           const posix_spawnattr_t *attrp, char *const argv[], char *const envp[]);
extern int old_posix_spawnp(pid_t *pid, const char *file, const posix_spawn_file_actions_t *file_actions,
                            const posix_spawnattr_t *attrp, char *const argv[], char *const envp[]);

#if defined(__x86_64__)
__asm__(".symver old_posix_spawn,posix_spawn@GLIBC_2.2.5");
__asm__(".symver old_posix_spawnp,posix_spawnp@GLIBC_2.2.5");
#endif
#else
static posix_spawn_func_t resolve_posix_spawn_func(const char *name)
{
    union
    {
        void *obj;
        posix_spawn_func_t fn;
    } sym = {0};

    sym.obj = dlsym(RTLD_NEXT, name);
    return sym.fn;
}

static int old_posix_spawn(pid_t *pid, const char *path, const posix_spawn_file_actions_t *file_actions,
                           const posix_spawnattr_t *attrp, char *const argv[], char *const envp[])
{
    posix_spawn_func_t fn = resolve_posix_spawn_func("posix_spawn");
    if (fn == NULL)
    {
        return ENOSYS;
    }
    return fn(pid, path, file_actions, attrp, argv, envp);
}

static int old_posix_spawnp(pid_t *pid, const char *file, const posix_spawn_file_actions_t *file_actions,
                            const posix_spawnattr_t *attrp, char *const argv[], char *const envp[])
{
    posix_spawn_func_t fn = resolve_posix_spawn_func("posix_spawnp");
    if (fn == NULL)
    {
        return ENOSYS;
    }
    return fn(pid, file, file_actions, attrp, argv, envp);
}
#endif

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

    return WIFEXITED(stat) && WEXITSTATUS(stat) == 0 ? true : false;
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

struct rewrite_exec_result
{
    char *const *multi_args[3];
};

static void rewrite_exec_result_free(struct rewrite_exec_result *result)
{
    if (result == NULL)
    {
        return;
    }
    multi_args_free(result->multi_args, 3);
    __builtin_memset(result, 0, sizeof(*result));
}

static bool build_rewritten_exec_args(const char *path, char *const argv[], char *const envp[],
                                      struct rewrite_exec_result *result)
{
    __builtin_memset(result, 0, sizeof(*result));

    char tmpid[DATAKIT_INJ_RAND_BYTES * 2 + 1] = {0};
    // retrieve modified launch parameters from file

    if (!gen_inject_tmpid_b16(tmpid))
    {
        debug_log("gen tmpid failed\n");
        return false;
    }

    int count_argv = varb_array_len(argv);
    char **dup_argv = malloc((count_argv + 2 + 1) * sizeof(char const *));
    if (dup_argv == NULL)
    {
        debug_perror("malloc argv");
        return false;
    }
    __builtin_memset(dup_argv, 0, (count_argv + 2 + 1) * sizeof(char const *));

    dup_argv[0] = strdup(tmpid);
    dup_argv[1] = strdup(path);
    dup_argv[count_argv + 2] = NULL;

    for (int i = 0; i < count_argv; i++)
    {
        dup_argv[i + 2] = strdup(argv[i]);
    }

    for (int i = 0; i < count_argv + 2; i++)
    {
        if (dup_argv[i] == NULL)
        {
            for (int j = 0; j < count_argv + 2; j++)
            {
                free((void *)dup_argv[j]);
            }
            free(dup_argv);
            debug_perror("strdup argv");
            return false;
        }
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
        return false;
    }

    char rewrite_args_fname[sizeof(DATAKIT_INJ_FILENAME_PREFIX) + sizeof(tmpid)] = {0};
    strcat(rewrite_args_fname, DATAKIT_INJ_FILENAME_PREFIX);
    strcat(rewrite_args_fname, tmpid);

    FILE *fp = fopen(rewrite_args_fname, "r");
    if (fp == NULL)
    {
        debug_perror(rewrite_args_fname);
        return false;
    }

    for (int i = 0; i < 3; i++)
    {
        char *const *val = read_args(fp);
        if ((val == NULL) || (i == 0 && val[0] == NULL))
        {
            if (val != NULL)
            {
                char *const *invalid_args[] = {val};
                multi_args_free(invalid_args, 1);
            }
            multi_args_free(result->multi_args, i);
            fclose(fp);
#ifdef DATAKIT_DEBUG
#else
            remove(rewrite_args_fname);
#endif
            return false;
        }
        result->multi_args[i] = val;
    }
    fclose(fp);
#ifdef DATAKIT_DEBUG
#else
    remove(rewrite_args_fname);
#endif

    debug_log("rewrite args: %s %s %s\n", result->multi_args[0][0], result->multi_args[1][0], result->multi_args[2][0]);

    return true;
}

static int rewrite_execve(const char *path, char *const argv[], char *const envp[])
{
    if (path == NULL)
    {
        return old_execve(path, argv, envp);
    }

    struct rewrite_exec_result rewrite_result = {0};
    if (!build_rewritten_exec_args(path, argv, envp, &rewrite_result))
    {
        return old_execve(path, argv, envp);
    }

    int ret = old_execve(rewrite_result.multi_args[0][0], rewrite_result.multi_args[1], rewrite_result.multi_args[2]);
    int saved_errno = errno;
    rewrite_exec_result_free(&rewrite_result);
    errno = saved_errno;
    return ret;
}

static char *build_exec_path(const char *dir, size_t dir_len, const char *file)
{
    size_t file_len = strlen(file);
    size_t need_slash = dir_len > 0 ? 1 : 0;
    char *out = malloc(dir_len + need_slash + file_len + 1);
    if (out == NULL)
    {
        return NULL;
    }

    size_t pos = 0;
    for (size_t i = 0; i < dir_len; i++)
    {
        out[pos++] = dir[i];
    }
    if (need_slash)
    {
        out[pos++] = '/';
    }
    for (size_t i = 0; i < file_len; i++)
    {
        out[pos++] = file[i];
    }
    out[pos] = '\0';
    return out;
}

static char **build_shell_argv(const char *file, char *const argv[])
{
    int argc = varb_array_len(argv);
    int extra_argc = argc > 0 ? argc - 1 : 0;
    char **shell_argv = malloc((extra_argc + 3) * sizeof(char *));
    if (shell_argv == NULL)
    {
        errno = ENOMEM;
        return NULL;
    }

    int pos = 0;
    shell_argv[pos++] = "/bin/sh";
    shell_argv[pos++] = (char *)file;
    for (int i = 1; i < argc; i++)
    {
        shell_argv[pos++] = argv[i];
    }
    shell_argv[pos] = NULL;

    return shell_argv;
}

static int shell_exec_script(const char *file, char *const argv[], char *const envp[])
{
    char **shell_argv = build_shell_argv(file, argv);
    if (shell_argv == NULL)
    {
        return -1;
    }

    int ret = rewrite_execve("/bin/sh", shell_argv, envp);
    int saved_errno = errno;
    free(shell_argv);
    errno = saved_errno;
    return ret;
}

static int rewrite_execvpe(const char *file, char *const argv[], char *const envp[])
{
    if (file == NULL || file[0] == '\0')
    {
        errno = ENOENT;
        return -1;
    }

    if (strchr(file, '/') != NULL)
    {
        int ret = rewrite_execve(file, argv, envp);
        if (ret == -1 && errno == ENOEXEC)
        {
            return shell_exec_script(file, argv, envp);
        }
        return ret;
    }

    char *path_env = getenv("PATH");
    if (path_env == NULL)
    {
        path_env = (char *)default_exec_path;
    }

    bool got_eacces = false;
    const char *cur = path_env;
    for (;;)
    {
        const char *end = strchr(cur, ':');
        size_t dir_len = end == NULL ? strlen(cur) : (size_t)(end - cur);

        char *candidate = build_exec_path(cur, dir_len, file);
        if (candidate == NULL)
        {
            errno = ENOMEM;
            return -1;
        }

        if (effective_exec_access(candidate) != 0)
        {
            int saved_errno = errno;
            free(candidate);
            if (saved_errno == EACCES)
            {
                got_eacces = true;
            }
            else if (saved_errno != ENOENT && saved_errno != ENOTDIR)
            {
                errno = saved_errno;
                return -1;
            }

            if (end == NULL)
            {
                break;
            }
            cur = end + 1;
            continue;
        }

        int ret = rewrite_execve(candidate, argv, envp);
        int saved_errno = errno;
        if (ret == -1 && saved_errno == ENOEXEC)
        {
            ret = shell_exec_script(candidate, argv, envp);
            saved_errno = errno;
        }
        free(candidate);
        errno = saved_errno;

        if (ret != -1)
        {
            return ret;
        }
        if (saved_errno == EACCES)
        {
            got_eacces = true;
        }
        else if (saved_errno != ENOENT && saved_errno != ENOTDIR)
        {
            return -1;
        }

        if (end == NULL)
        {
            break;
        }
        cur = end + 1;
    }

    errno = got_eacces ? EACCES : ENOENT;
    return -1;
}

static char **build_argv_from_va(const char *arg, va_list ap)
{
    int argc = 0;
    if (arg != NULL)
    {
        argc = 1;
        va_list count_ap;
        va_copy(count_ap, ap);
        while (va_arg(count_ap, char *) != NULL)
        {
            argc++;
        }
        va_end(count_ap);
    }

    char **argv = malloc((argc + 1) * sizeof(char *));
    if (argv == NULL)
    {
        errno = ENOMEM;
        return NULL;
    }

    if (arg != NULL)
    {
        argv[0] = (char *)arg;

        va_list fill_ap;
        va_copy(fill_ap, ap);
        for (int i = 1; i < argc; i++)
        {
            argv[i] = va_arg(fill_ap, char *);
        }
        va_end(fill_ap);
    }
    argv[argc] = NULL;

    return argv;
}

static char **build_execle_argv_envp_from_va(const char *arg, va_list ap, char *const **envp)
{
    int argc = 0;
    if (arg != NULL)
    {
        argc = 1;
        va_list count_ap;
        va_copy(count_ap, ap);
        while (va_arg(count_ap, char *) != NULL)
        {
            argc++;
        }
        va_end(count_ap);
    }

    char **argv = malloc((argc + 1) * sizeof(char *));
    if (argv == NULL)
    {
        errno = ENOMEM;
        return NULL;
    }

    va_list fill_ap;
    va_copy(fill_ap, ap);
    if (arg != NULL)
    {
        argv[0] = (char *)arg;
        for (int i = 1; i < argc; i++)
        {
            argv[i] = va_arg(fill_ap, char *);
        }
        (void)va_arg(fill_ap, char *);
    }
    *envp = va_arg(fill_ap, char *const *);
    va_end(fill_ap);

    argv[argc] = NULL;
    return argv;
}

int execve(const char *path, char *const argv[], char *const envp[])
{
    return rewrite_execve(path, argv, envp);
}

int execv(const char *path, char *const argv[])
{
    return rewrite_execve(path, argv, environ);
}

int execvp(const char *file, char *const argv[])
{
    return rewrite_execvpe(file, argv, environ);
}

int execvpe(const char *file, char *const argv[], char *const envp[])
{
    return rewrite_execvpe(file, argv, envp);
}

int execl(const char *path, const char *arg, ...)
{
    va_list ap;
    va_start(ap, arg);
    char **argv = build_argv_from_va(arg, ap);
    va_end(ap);

    if (argv == NULL)
    {
        return -1;
    }

    int ret = rewrite_execve(path, argv, environ);
    int saved_errno = errno;
    free(argv);
    errno = saved_errno;
    return ret;
}

int execlp(const char *file, const char *arg, ...)
{
    va_list ap;
    va_start(ap, arg);
    char **argv = build_argv_from_va(arg, ap);
    va_end(ap);

    if (argv == NULL)
    {
        return -1;
    }

    int ret = rewrite_execvpe(file, argv, environ);
    int saved_errno = errno;
    free(argv);
    errno = saved_errno;
    return ret;
}

int execle(const char *path, const char *arg, ...)
{
    char *const *envp = NULL;
    va_list ap;
    va_start(ap, arg);
    char **argv = build_execle_argv_envp_from_va(arg, ap, &envp);
    va_end(ap);

    if (argv == NULL)
    {
        return -1;
    }

    int ret = rewrite_execve(path, argv, envp);
    int saved_errno = errno;
    free(argv);
    errno = saved_errno;
    return ret;
}

int posix_spawn(pid_t *pid, const char *path, const posix_spawn_file_actions_t *file_actions,
                const posix_spawnattr_t *attrp, char *const argv[], char *const envp[])
{
    if (path == NULL)
    {
        return old_posix_spawn(pid, path, file_actions, attrp, argv, envp);
    }

    struct rewrite_exec_result rewrite_result = {0};
    if (!build_rewritten_exec_args(path, argv, envp, &rewrite_result))
    {
        return old_posix_spawn(pid, path, file_actions, attrp, argv, envp);
    }

    int ret = old_posix_spawn(pid, rewrite_result.multi_args[0][0], file_actions, attrp, rewrite_result.multi_args[1],
                              rewrite_result.multi_args[2]);
    rewrite_exec_result_free(&rewrite_result);
    return ret;
}

int posix_spawnp(pid_t *pid, const char *file, const posix_spawn_file_actions_t *file_actions,
                 const posix_spawnattr_t *attrp, char *const argv[], char *const envp[])
{
    if (file == NULL)
    {
        return old_posix_spawnp(pid, file, file_actions, attrp, argv, envp);
    }
    if (file[0] == '\0')
    {
        return ENOENT;
    }

    if (strchr(file, '/') != NULL)
    {
        struct rewrite_exec_result rewrite_result = {0};
        if (!build_rewritten_exec_args(file, argv, envp, &rewrite_result))
        {
            return old_posix_spawnp(pid, file, file_actions, attrp, argv, envp);
        }

        int ret = old_posix_spawn(pid, rewrite_result.multi_args[0][0], file_actions, attrp,
                                  rewrite_result.multi_args[1], rewrite_result.multi_args[2]);
        rewrite_exec_result_free(&rewrite_result);
        if (ret == 0)
        {
            return ret;
        }

        return old_posix_spawnp(pid, file, file_actions, attrp, argv, envp);
    }

    char *path_env = getenv("PATH");
    if (path_env == NULL)
    {
        path_env = (char *)default_exec_path;
    }

    const char *cur = path_env;
    for (;;)
    {
        const char *end = strchr(cur, ':');
        size_t dir_len = end == NULL ? strlen(cur) : (size_t)(end - cur);

        char *candidate = build_exec_path(cur, dir_len, file);
        if (candidate == NULL)
        {
            return ENOMEM;
        }

        if (effective_exec_access(candidate) != 0)
        {
            free(candidate);
            if (end == NULL)
            {
                break;
            }
            cur = end + 1;
            continue;
        }

        struct rewrite_exec_result rewrite_result = {0};
        bool rewritten = build_rewritten_exec_args(candidate, argv, envp, &rewrite_result);
        free(candidate);

        if (rewritten)
        {
            int ret = old_posix_spawn(pid, rewrite_result.multi_args[0][0], file_actions, attrp,
                                      rewrite_result.multi_args[1], rewrite_result.multi_args[2]);
            rewrite_exec_result_free(&rewrite_result);

            if (ret == 0)
            {
                return ret;
            }
            if (ret != ENOENT && ret != ENOTDIR && ret != EACCES)
            {
                return ret;
            }
        }

        if (end == NULL)
        {
            break;
        }
        cur = end + 1;
    }

    return old_posix_spawnp(pid, file, file_actions, attrp, argv, envp);
}
