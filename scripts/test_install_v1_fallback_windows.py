"""Offline routing tests. Set POWERSHELL to pwsh or powershell.exe."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


TEMPLATE = Path(__file__).resolve().parents[1] / 'templates/install.template.ps1'
POWERSHELL = os.environ.get('POWERSHELL') or shutil.which('pwsh') or shutil.which('powershell')


@unittest.skipUnless(POWERSHELL, 'PowerShell is required')
class WindowsFallbackTest(unittest.TestCase):
    def run_script(self, **variables):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            script = TEMPLATE.read_text().split('$x = [Environment]::GetEnvironmentVariable("DK_HOSTNAME")', 1)[0]
            script = script.replace('$domain = @(', '''
function remove-host { param($filename, $hostname) }
function Add-Content { param($Path, $Value, [switch]$Force) Write-Output "HOSTMAP:$Value" }
$domain = @(''', 1)
            script = script.replace('{{.InstallBaseURL}}', 'release.example/datakit-v2')
            script = script.replace(
                '$windowsMajorVersion = [Environment]::OSVersion.Version.Major',
                '$windowsMajorVersion = [int]$env:TEST_OS_MAJOR')
            mocks = '''
function Set-ExecutionPolicy {}
function Import-Module {}
function Start-BitsTransfer {
    param($Source, $Destination, $ErrorAction, $ProxyUsage, $ProxyList)
    Write-Output "DOWNLOAD_PROXY:$ProxyUsage|$ProxyList"
    Set-Content -Path $env:TEST_DOWNLOAD_URL -Value $Source
    if ($env:TEST_DOWNLOAD_FAIL -eq '1') { throw 'download failed' }
    if ($env:TEST_EMPTY -eq '1') {
        Set-Content -Path $Destination -Value '' -NoNewline
    } else {
        Copy-Item $env:TEST_CHILD_SCRIPT $Destination
    }
}
'''
            child = '''
Write-Output "CHILD:$env:DK_INSTALLER_BASE_URL|$env:DK_DATAWAY|$env:DK_UPGRADE"
if ($env:TEST_THROW -eq '1') { throw 'installation failed' }
if ($env:TEST_WRITE_ERROR -eq '1') { Write-Error 'installation failed'; Write-Output 'must not reach cleanup' }
if ($env:TEST_NATIVE_FAIL -eq '1') {
    & (Get-Process -Id $PID).Path -NoProfile -Command 'exit 9'
    Write-Output 'cleanup after native installer'
} else {
    exit ([int]$env:TEST_CHILD_EXIT)
}
'''
            (root / 'child.ps1').write_text(child)
            (root / 'test.ps1').write_text(mocks + script + '\nWrite-Output "CONTINUE"\n')
            env = {k: v for k, v in os.environ.items()
                   if not k.startswith('DK_') and not k.lower().endswith('_proxy')}
            env.update(TEMP=str(root), TEST_OS_MAJOR='6', TEST_CHILD_EXIT='0',
                       TEST_DOWNLOAD_URL=str(root / 'url'), TEST_CHILD_SCRIPT=str(root / 'child.ps1'))
            env.update(variables)
            result = subprocess.run([POWERSHELL, '-NoProfile', '-File', str(root / 'test.ps1')],
                                    env=env, capture_output=True, text=True)
            url = (root / 'url').read_text(encoding='utf-8-sig').strip() if (root / 'url').exists() else ''
            self.assertEqual(list(root.glob('dk-v1-*')), [], result.stderr)
            return result, url

    def test_disabled(self):
        for value in ('', '0', 'true'):
            result, url = self.run_script(DK_ALLOW_V1_FALLBACK=value)
            self.assertEqual(result.returncode, 1, result.stderr)
            self.assertEqual(url, '')

    def test_supported_os(self):
        result, url = self.run_script(TEST_OS_MAJOR='10', DK_ALLOW_V1_FALLBACK='1')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('CONTINUE', result.stdout)
        self.assertEqual(url, '')

    def test_source_environment_and_upgrade(self):
        for upgrade in ('', '0', '1'):
            result, url = self.run_script(DK_ALLOW_V1_FALLBACK='1', DK_UPGRADE=upgrade,
                DK_INSTALLER_BASE_URL='https://mirror.example/nested/datakit-v2/',
                DK_DATAWAY='https://dataway.example?token=test')
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(url, 'https://mirror.example/nested/datakit/install.ps1')
            self.assertIn('CHILD:https://mirror.example/nested/datakit|https://dataway.example?token=test|' + upgrade, result.stdout)
            self.assertNotIn('CONTINUE', result.stdout)

    def test_default_source_and_child_exit(self):
        for status in ('0', '7'):
            result, url = self.run_script(DK_ALLOW_V1_FALLBACK='1', TEST_CHILD_EXIT=status)
            self.assertEqual(result.returncode, int(status), result.stderr)
            self.assertEqual(url, 'https://release.example/datakit/install.ps1')
            self.assertNotIn('CONTINUE', result.stdout)

    def test_native_failure(self):
        result, _ = self.run_script(DK_ALLOW_V1_FALLBACK='1', TEST_NATIVE_FAIL='1')
        self.assertEqual(result.returncode, 9, result.stderr)

    def test_failures(self):
        for flag in ('TEST_DOWNLOAD_FAIL', 'TEST_EMPTY', 'TEST_THROW', 'TEST_WRITE_ERROR'):
            result, _ = self.run_script(DK_ALLOW_V1_FALLBACK='1', **{flag: '1'})
            self.assertEqual(result.returncode, 1, result.stderr)
            self.assertNotIn('CONTINUE', result.stdout)

    def test_custom_path(self):
        result, url = self.run_script(DK_ALLOW_V1_FALLBACK='1',
                                      DK_INSTALLER_BASE_URL='https://mirror.example/custom')
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertEqual(url, '')

    def test_proxy(self):
        for variables, expected in (
            ({}, 'DOWNLOAD_PROXY:|'),
            ({'HTTP_PROXY': 'http://http.example:8080'}, 'DOWNLOAD_PROXY:Override|http://http.example:8080'),
            ({'HTTPS_PROXY': 'http://https.example:8080'}, 'DOWNLOAD_PROXY:Override|http://https.example:8080'),
            ({'HTTP_PROXY': 'http://http.example:8080', 'HTTPS_PROXY': 'http://https.example:8080'}, 'DOWNLOAD_PROXY:Override|http://https.example:8080'),
        ):
            result, _ = self.run_script(DK_ALLOW_V1_FALLBACK='1', **variables)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn(expected, result.stdout)

    def test_nginx_proxy(self):
        result, _ = self.run_script(DK_ALLOW_V1_FALLBACK='1',
            DK_PROXY_TYPE='nginx', DK_NGINX_IP='192.0.2.10',
            HTTP_PROXY='http://http.example:8080', HTTPS_PROXY='http://https.example:8080')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('192.0.2.10\tstatic.guance.com', result.stdout)
        self.assertLess(result.stdout.index('HOSTMAP:'), result.stdout.index('DOWNLOAD_PROXY:'))
        self.assertIn('DOWNLOAD_PROXY:NoProxy|', result.stdout)

    def test_disabled_fallback_does_not_configure_proxy(self):
        result, url = self.run_script(DK_PROXY_TYPE='nginx', DK_NGINX_IP='192.0.2.10')
        self.assertEqual(result.returncode, 1)
        self.assertNotIn('HOSTMAP:', result.stdout)
        self.assertNotIn('Set nginx proxy', result.stdout)
        self.assertEqual(url, '')


if __name__ == '__main__':
    unittest.main()
