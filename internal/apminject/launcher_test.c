#define _GNU_SOURCE

#include <assert.h>
#include <stdarg.h>

#include "apm_launcher.c"

void test_gen_inject_tmpid_b16(char *tmpid)
{
    assert(gen_inject_tmpid_b16(tmpid));
    assert(strlen(tmpid) == DATAKIT_INJ_RAND_BYTES * 2);
}

void test_malloc(void)
{
    void *arr = NULL;
    if (arr = malloc(16), arr == NULL)
    {
        assert(false);
    }
    assert(arr != NULL);
    free(arr);
}

void test_build_exec_path(void)
{
    char *path = build_exec_path("/usr/bin", strlen("/usr/bin"), "java");
    assert(path != NULL);
    assert(strcmp(path, "/usr/bin/java") == 0);
    free(path);

    path = build_exec_path("", 0, "java");
    assert(path != NULL);
    assert(strcmp(path, "java") == 0);
    free(path);
}

void test_build_shell_argv(void)
{
    char *const empty[] = {
        NULL,
    };
    char **shell_argv = build_shell_argv("/tmp/noshebang", empty);
    assert(shell_argv != NULL);
    assert(strcmp(shell_argv[0], "/bin/sh") == 0);
    assert(strcmp(shell_argv[1], "/tmp/noshebang") == 0);
    assert(shell_argv[2] == NULL);
    free(shell_argv);

    char *const args[] = {
        "noshebang",
        "one",
        "two",
        NULL,
    };
    shell_argv = build_shell_argv("/tmp/noshebang", args);
    assert(shell_argv != NULL);
    assert(strcmp(shell_argv[0], "/bin/sh") == 0);
    assert(strcmp(shell_argv[1], "/tmp/noshebang") == 0);
    assert(strcmp(shell_argv[2], "one") == 0);
    assert(strcmp(shell_argv[3], "two") == 0);
    assert(shell_argv[4] == NULL);
    free(shell_argv);
}

static char **build_argv_from_test_va(const char *arg, ...)
{
    va_list ap;
    va_start(ap, arg);
    char **argv = build_argv_from_va(arg, ap);
    va_end(ap);
    return argv;
}

void test_build_argv_from_va(void)
{
    char **argv = build_argv_from_test_va("java", "-version", NULL);
    assert(argv != NULL);
    assert(strcmp(argv[0], "java") == 0);
    assert(strcmp(argv[1], "-version") == 0);
    assert(argv[2] == NULL);
    free(argv);

    argv = build_argv_from_test_va(NULL);
    assert(argv != NULL);
    assert(argv[0] == NULL);
    free(argv);
}

static char **build_execle_argv_from_test_va(char *const **envp, const char *arg, ...)
{
    va_list ap;
    va_start(ap, arg);
    char **argv = build_execle_argv_envp_from_va(arg, ap, envp);
    va_end(ap);
    return argv;
}

void test_build_execle_argv_envp_from_va(void)
{
    char *const envp[] = {
        "A=B",
        NULL,
    };
    char *const *got_envp = NULL;

    char **argv = build_execle_argv_from_test_va(&got_envp, "java", "-version", NULL, envp);
    assert(argv != NULL);
    assert(strcmp(argv[0], "java") == 0);
    assert(strcmp(argv[1], "-version") == 0);
    assert(argv[2] == NULL);
    assert(got_envp == envp);
    free(argv);

    got_envp = NULL;
    argv = build_execle_argv_from_test_va(&got_envp, NULL, envp);
    assert(argv != NULL);
    assert(argv[0] == NULL);
    assert(got_envp == envp);
    free(argv);
}

int main(void)
{
    test_malloc();
    test_build_exec_path();
    test_build_shell_argv();
    test_build_argv_from_va();
    test_build_execle_argv_envp_from_va();

    char tmpid[DATAKIT_INJ_RAND_BYTES * 2 + 1] = {0};
    test_gen_inject_tmpid_b16(tmpid);

    printf("PASS\n");
    return 0;
}
