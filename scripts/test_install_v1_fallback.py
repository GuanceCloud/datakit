"""Offline Linux installer routing tests; never execute a real installer."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


TEMPLATE = Path(__file__).resolve().parents[1] / "templates/install.template.sh"


class LinuxFallbackTest(unittest.TestCase):
    def run_script(self, kernel="2.6.32", **variables):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            # Exercise the real preflight, stopping before any installation work.
            script = TEMPLATE.read_text().split('if [ -n "$DK_HOSTNAME" ]; then', 1)[0]
            script = script.replace('{{.Version}}', '2.1.0').replace(
                '{{.InstallBaseURL}}', 'release.example/datakit-v2')
            script = script.replace('v1_fallback=0\n',
                'updateHosts() { printf "HOSTMAP:%s:%s\\n" "$1" "$2"; }\nv1_fallback=0\n', 1)
            script += '\nprintf "CONTINUE\\n"\n'
            (root / 'install.sh').write_text(script)
            mocks = {
                'uname': '#!/bin/sh\ncase "$1" in -m) echo x86_64;; -r) echo "$TEST_KERNEL";; esac\n',
                'curl': '''#!/bin/sh
printf '%s\\n' "$@" > "$TEST_CURL_ARGS"
if [ "$TEST_DOWNLOAD_FAIL" = 1 ]; then exit 22; fi
if [ "$TEST_EMPTY" = 1 ]; then exit 0; fi
cat <<'CHILD'
printf 'CHILD:%s|%s|%s\\n' "$DK_INSTALLER_BASE_URL" "$DK_ALLOW_V1_FALLBACK" "$DK_DATAWAY"
exit "${TEST_CHILD_EXIT:-0}"
CHILD
''',
            }
            for name, content in mocks.items():
                path = root / name
                path.write_text(content)
                path.chmod(0o755)
            env = {k: v for k, v in os.environ.items()
                   if not k.startswith('DK_') and not k.lower().endswith('_proxy')}
            env.update(PATH=str(root) + ':' + os.environ['PATH'],
                       OSTYPE='linux-gnu', TEST_KERNEL=kernel,
                       TEST_CURL_ARGS=str(root / 'curl-args'))
            env.update(variables)
            result = subprocess.run(['bash', str(root / 'install.sh')],
                                    env=env, cwd=root, capture_output=True, text=True)
            args = (root / 'curl-args').read_text() if (root / 'curl-args').exists() else ''
            return result, args

    def test_disabled(self):
        for value in ('0', 'true', ''):
            result, args = self.run_script(DK_ALLOW_V1_FALLBACK=value)
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual(args, '')

    def test_compatible_kernel(self):
        for kernel in ('3.2.0', '3.10.0-1160.el7.x86_64', '6.8.0'):
            result, args = self.run_script(kernel=kernel, DK_ALLOW_V1_FALLBACK='1')
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn('CONTINUE', result.stdout)
            self.assertEqual(args, '')

    def test_mirror_environment_and_exit_status(self):
        for source in ('https://mirror.example/datakit-v2/', 'https://mirror.example/nested/datakit-v2'):
            for status in ('0', '7'):
                result, args = self.run_script(
                    DK_ALLOW_V1_FALLBACK='1', DK_INSTALLER_BASE_URL=source,
                    DK_DATAWAY='https://dataway.example?token=test', TEST_CHILD_EXIT=status)
                v1 = source.rstrip('/').removesuffix('/datakit-v2') + '/datakit'
                self.assertEqual(result.returncode, int(status), result.stderr)
                self.assertIn(v1 + '/install.sh', args)
                self.assertIn('CHILD:' + v1 + '|1|https://dataway.example?token=test', result.stdout)
                self.assertNotIn('CONTINUE', result.stdout)

    def test_default_source(self):
        result, args = self.run_script(DK_ALLOW_V1_FALLBACK='1')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('https://release.example/datakit/install.sh', args)

    def test_download_errors(self):
        for variables, status in (({'TEST_DOWNLOAD_FAIL': '1'}, 22), ({'TEST_EMPTY': '1'}, 1)):
            result, _ = self.run_script(DK_ALLOW_V1_FALLBACK='1', **variables)
            self.assertEqual(result.returncode, status)
            self.assertNotIn('CHILD:', result.stdout)
            self.assertNotIn('CONTINUE', result.stdout)

    def test_unknown_kernel_preserves_existing_behavior(self):
        for kernel in ('unknown', '', '3.unknown'):
            result, args = self.run_script(kernel=kernel, DK_ALLOW_V1_FALLBACK='1')
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn('CONTINUE', result.stdout)
            self.assertEqual(args, '')

    def test_custom_path(self):
        result, args = self.run_script(DK_ALLOW_V1_FALLBACK='1',
                                       DK_INSTALLER_BASE_URL='https://mirror.example/custom')
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(args, '')

    def test_proxy(self):
        for variables, expected in (
            ({}, None),
            ({'HTTP_PROXY': 'http://http.example:8080'}, 'http://http.example:8080'),
            ({'HTTPS_PROXY': 'http://https.example:8080'}, 'http://https.example:8080'),
            ({'HTTP_PROXY': 'http://http.example:8080', 'HTTPS_PROXY': 'http://https.example:8080'}, 'http://https.example:8080'),
        ):
            result, args = self.run_script(DK_ALLOW_V1_FALLBACK='1', **variables)
            self.assertEqual(result.returncode, 0, result.stderr)
            if expected:
                self.assertIn('-x\n' + expected + '\n', args)
            else:
                self.assertNotIn('-x\n', args)

    def test_nginx_proxy(self):
        result, args = self.run_script(DK_ALLOW_V1_FALLBACK='1',
            DK_PROXY_TYPE='nginx', DK_NGINX_IP='192.0.2.10',
            HTTP_PROXY='http://http.example:8080', HTTPS_PROXY='http://https.example:8080')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('HOSTMAP:192.0.2.10:static.guance.com', result.stdout)
        self.assertLess(result.stdout.index('HOSTMAP:'), result.stdout.index('requires DataKit 1.x'))
        self.assertIn('--noproxy\n*\n', args)
        self.assertNotIn('-x\n', args)

    def test_disabled_fallback_does_not_configure_proxy(self):
        result, args = self.run_script(DK_PROXY_TYPE='nginx', DK_NGINX_IP='192.0.2.10')
        self.assertEqual(result.returncode, 1)
        self.assertNotIn('HOSTMAP:', result.stdout)
        self.assertNotIn('Set nginx proxy', result.stdout)
        self.assertEqual(args, '')

if __name__ == '__main__':
    unittest.main()
