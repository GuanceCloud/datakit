import { message } from 'antd';
import {
  aesEncrypt,
  alertError,
  isContainerMode,
  isDatakitManagement,
  isDatakitUpgradeable,
  isLoadingStatus,
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
    expect(isLoadingStatus(restarting)).toBe(true);
    expect(isDatakitUpgradeable(running, '2.0.0')).toBe(true);
    expect(isDatakitUpgradeable(offline, '2.0.0')).toBe(false);
    expect(isDatakitUpgradeable(inContainer, '2.0.0')).toBe(false);
    expect(isContainerMode(inContainer)).toBe(true);
    expect(isContainerMode()).toBe(false);
  });

  it('runs jobs with concurrency control', async () => {
    const promise = runJob(2, [1, 2, 3], async (item: number) => item * 2);
    jest.advanceTimersByTime(1000);
    const result = await promise;

    expect(result).toHaveLength(3);
    expect(result.every((item) => item.status === 'fulfilled')).toBe(true);
  });
});
