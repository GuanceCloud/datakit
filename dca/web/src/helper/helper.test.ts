import { message } from 'antd';
import {
  aesEncrypt,
  alertError,
  isContainerMode,
  compareDatakitVersions,
  getDatakitVersionLine,
  getLatestDatakitVersion,
  isDatakitManagement,
  isDatakitUpgradeable,
  isLoadingStatus,
  isNewerDatakitVersionAvailable,
  isPhoneNumber,
  isValidIP,
  runJob,
  showDuration,
  sleep,
} from './helper';
import { DCA_STATUS } from 'src/store/type';
import { getMsg } from 'src/store/baseApi';

jest.mock('antd', () => ({
  message: {
    error: jest.fn(),
  },
}));

jest.mock('src/store/baseApi', () => ({
  getMsg: jest.fn(() => 'formatted error'),
}));

describe('helper utilities', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.clearAllMocks();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('debounces alertError and formats object errors', () => {
    alertError({ errorCode: 'x', message: 'bad' });

    jest.advanceTimersByTime(1000);

    expect(getMsg).toHaveBeenCalledWith({ errorCode: 'x', message: 'bad' });
    expect(message.error).toHaveBeenCalled();
  });

  it('resolves sleep after the requested time', async () => {
    const promise = sleep(50);

    jest.advanceTimersByTime(50);

    await expect(promise).resolves.toBeUndefined();
  });

  it('encrypts deterministically with aesEncrypt', () => {
    expect(aesEncrypt('abc')).toBe(aesEncrypt('abc'));
    expect(aesEncrypt('abc')).not.toBe('abc');
  });

  it('validates phone numbers and ips', () => {
    expect(isPhoneNumber('13800138000')).toBe(true);
    expect(isPhoneNumber('1234')).toBe(false);
    expect(isValidIP('192.168.1.1')).toBe(true);
    expect(isValidIP('999.1.1.1')).toBe(false);
  });

  it('formats nanosecond durations to the right unit', () => {
    expect(showDuration(0)).toBe('');
    expect(showDuration(1500)).toBe('1.50µs');
    expect(showDuration(1500000)).toBe('1.50ms');
    expect(showDuration(1500000000)).toBe('1.50s');
  });

  it('evaluates datakit status helpers', () => {
    const running = { status: DCA_STATUS.RUNNING, version: '1.0.0', run_in_container: false } as any;
    const restarting = { status: DCA_STATUS.RESTARTING } as any;
    const offline = { status: DCA_STATUS.OFFLINE, version: '1.0.0', run_in_container: false } as any;
    const inContainer = { status: DCA_STATUS.RUNNING, version: '1.0.0', run_in_container: true } as any;

    expect(isDatakitManagement(running)).toBe(true);
    // a running row without a live session (reported by the DCA backend) must not
    // expose the management actions; older backends omit the field
    expect(isDatakitManagement({ status: DCA_STATUS.RUNNING, alive: false } as any)).toBe(false);
    expect(isDatakitManagement({ status: DCA_STATUS.RUNNING, alive: true } as any)).toBe(true);
    expect(isLoadingStatus(restarting)).toBe(true);
    const latestVersions = { v1: '1.1.0', v2: '2.0.0' };
    expect(isDatakitUpgradeable(running, latestVersions)).toBe(true);
    expect(isDatakitUpgradeable(offline, latestVersions)).toBe(false);
    expect(isDatakitUpgradeable(inContainer, latestVersions)).toBe(false);
    expect(isContainerMode(inContainer)).toBe(true);
    expect(isContainerMode()).toBe(false);
  });

  it('selects and compares versions within the same release line', () => {
    const latestVersions = { v1: '1.94.1', v2: '2.9.0' };

    expect(getDatakitVersionLine('v1.93.0')).toBe('v1');
    expect(getDatakitVersionLine('2.8.0')).toBe('v2');
    expect(getDatakitVersionLine('3.0.0')).toBeUndefined();
    expect(getLatestDatakitVersion('1.93.0', latestVersions)).toBe('1.94.1');
    expect(getLatestDatakitVersion('2.8.0', latestVersions)).toBe('2.9.0');
    expect(isNewerDatakitVersionAvailable('1.93.0', latestVersions)).toBe(true);
    expect(isNewerDatakitVersionAvailable('2.9.0', latestVersions)).toBe(false);
    expect(isNewerDatakitVersionAvailable('2.10.0', latestVersions)).toBe(false);
    expect(compareDatakitVersions('2.9.0-rc1', '2.9.0')).toBeLessThan(0);
    expect(compareDatakitVersions('invalid', '2.9.0')).toBeUndefined();
  });

  it('runs jobs with concurrency control', async () => {
    const promise = runJob(2, [1, 2, 3], async (item: number) => item * 2);
    jest.advanceTimersByTime(1000);
    const result = await promise;

    expect(result).toHaveLength(3);
    expect(result.every((item) => item.status === 'fulfilled')).toBe(true);
  });
});
