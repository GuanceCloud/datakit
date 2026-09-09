"""Offline contract tests; no real deploy token or registry requests."""
import hashlib
import io
import json
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile
import unittest

SCRIPT = Path(__file__).with_name('download-pipeline-jit.sh').resolve()


class DownloadTest(unittest.TestCase):
    def test_download_contract(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            registry = root / 'registry'
            registry.mkdir()
            def package(arch, revision):
                archive = registry / f'platypus-jit-linux-{arch}.tar.gz'
                with tarfile.open(archive, 'w:gz') as tar:
                    for name in ('libplatypus_jit.so', 'libplatypus_jit.so.sha256',
                                 'manifest.json', 'glibc-ceiling.txt'):
                        info = tarfile.TarInfo(f'linux-{arch}/{name}')
                        data = json.dumps({"source_revision": revision}).encode() if name == "manifest.json" else arch.encode()
                        info.size = len(data)
                        tar.addfile(info, io.BytesIO(data))
                archive.with_name(archive.name + '.sha256').write_text(
                    hashlib.sha256(archive.read_bytes()).hexdigest() + '  ' + archive.name + '\n')
            for arch in ('amd64', 'arm64'):
                package(arch, 'a' * 40)
            curl = root / 'curl'
            curl.write_text('''#!/usr/bin/env python3
import os, pathlib, sys, shutil
assert sys.stdin.read() == 'DEPLOY-TOKEN: test-secret\\n'
args=sys.argv[1:]
shutil.copyfile(pathlib.Path(os.environ['TEST_REGISTRY']) / args[-1].rsplit('/',1)[-1], args[args.index('--output')+1])
print('200',end='')
''')
            curl.chmod(0o755)
            env = dict(os.environ, PATH=str(root)+':'+os.environ['PATH'],
                       TEST_REGISTRY=str(registry), PIPELINE_JIT_ENABLED='1',
                       PP_JIT_READ_TOKEN='test-secret', PP_JIT_VERSION='v0.1.0',
                       PLATYPUS_JIT_RUNTIME_DIR=str(root/'output'))
            def run():
                return subprocess.run(['bash', str(SCRIPT)], env=env,
                                      capture_output=True, text=True)
            env.pop("PLATYPUS_JIT_EXPECTED_REVISION", None)
            result = run()
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertNotIn('test-secret', result.stdout + result.stderr)
            for arch in ('amd64', 'arm64'):
                self.assertEqual((root/'output'/f'linux-{arch}'/'libplatypus_jit.so').read_text(), arch)
            self.assertEqual((root/'output/source-revision').read_text().strip(), 'a'*40)
            for invalid in ('b'*40, 'bad-sha', None):
                package('arm64', invalid)
                result = run()
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual((root/'output/source-revision').read_text().strip(), 'a'*40)
            (registry/'platypus-jit-linux-arm64.tar.gz').write_bytes(b'corrupt')
            self.assertNotEqual(run().returncode, 0)
            self.assertEqual((root/'output/linux-arm64/libplatypus_jit.so').read_text(), 'arm64')
            env.pop('PP_JIT_READ_TOKEN')
            self.assertNotEqual(run().returncode, 0)
            env['PIPELINE_JIT_ENABLED'] = '0'
            self.assertEqual(run().returncode, 0)
            self.assertFalse(list((root/'output').glob('.download.*')))


if __name__ == '__main__':
    unittest.main()
