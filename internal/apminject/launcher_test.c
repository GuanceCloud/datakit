#include <assert.h>

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
    assert(arr!= NULL);
}

int main(void)
{
    test_malloc();

    char tmpid[DATAKIT_INJ_RAND_BYTES * 2 + 1] = {0};
    test_gen_inject_tmpid_b16(tmpid);

    char *const args[] = {
        tmpid,
        "/bin/ldd",
        "ldd",
        "--verson",
        NULL,
    };
    char *const *envp[] = {
        NULL,
    };
    // try_apm_inject_process(args, envp);

    printf("PASS\n");
    return 0;
}
