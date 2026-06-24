jest.mock('src/store/baseApi', () => ({
  getMsg: jest.fn((err) => err.message || err.errorCode),
}));

import { downloadLogFile, getDatakitConfig, getLogTail, saveDatakitConfig } from './api';
import { type IDatakit } from '../store/type';

describe('datakit api query params', () => {
  const datakit = { id: 'dk-1' } as IDatakit;

  afterEach(() => {
    jest.restoreAllMocks();
  });

  it('encodes query parameter values', async () => {
    const fetchMock = jest.spyOn(global, 'fetch').mockResolvedValue({
      status: 200,
      json: async () => ({
        code: 200,
        content: 'ok',
        success: true,
      }),
    } as Response);

    await getDatakitConfig(datakit, '/usr/local/datakit/conf.d/a&b #1.conf');

    const url = fetchMock.mock.calls[0][0] as string;
    expect(url).toContain('datakit_id=dk-1');
    expect(url).toContain('path=%2Fusr%2Flocal%2Fdatakit%2Fconf.d%2Fa%26b%20%231.conf');
    expect(url).not.toContain('/usr/local/datakit/conf.d/a&b #1.conf');
  });

  it('does not log request bodies on successful requests', async () => {
    jest.spyOn(global, 'fetch').mockResolvedValue({
      status: 200,
      json: async () => ({
        code: 200,
        content: {},
        success: true,
      }),
    } as Response);
    const debugSpy = jest.spyOn(console, 'debug').mockImplementation(() => undefined);

    await saveDatakitConfig(datakit, {
      path: '/usr/local/datakit/conf.d/secret.conf',
      config: 'token = "secret-token"',
      isNew: false,
      inputName: 'cpu',
      isForce: false,
    });

    expect(debugSpy).not.toHaveBeenCalled();
  });

  it('encodes datakit id on non-GET requests', async () => {
    const fetchMock = jest.spyOn(global, 'fetch').mockResolvedValue({
      status: 200,
      json: async () => ({
        code: 200,
        content: {},
        success: true,
      }),
    } as Response);

    await saveDatakitConfig({ ...datakit, id: 'dk&bad=1' }, {
      path: '/usr/local/datakit/conf.d/cpu.conf',
      config: 'interval = "10s"',
      isNew: false,
      inputName: 'cpu',
      isForce: false,
    });

    const url = fetchMock.mock.calls[0][0] as string;
    expect(url).toContain('datakit_id=dk%26bad%3D1');
    expect(url).not.toContain('datakit_id=dk&bad=1');
  });

  it('encodes log tail query values', async () => {
    const fetchMock = jest.spyOn(global, 'fetch').mockResolvedValue({
      status: 200,
      body: {
        getReader: jest.fn(),
      },
    } as unknown as Response);

    await getLogTail({ ...datakit, id: 'dk&bad=1' }, 'gin.log&debug=true');

    const url = fetchMock.mock.calls[0][0] as string;
    expect(url).toBe('/api/datakit/log/tail?datakit_id=dk%26bad%3D1&type=gin.log%26debug%3Dtrue');
    expect(url).not.toContain('dk&bad=1');
    expect(url).not.toContain('gin.log&debug=true');
  });

  it('encodes log download query values', async () => {
    const openSpy = jest.spyOn(window, 'open').mockImplementation(() => null);

    await downloadLogFile({ ...datakit, id: 'dk&bad=1' }, 'gin.log&debug=true');

    expect(openSpy).toHaveBeenCalledWith('/api/datakit/log/download?datakit_id=dk%26bad%3D1&type=gin.log%26debug%3Dtrue');
  });

  it('returns an error when log download window is blocked', async () => {
    jest.spyOn(window, 'open').mockImplementation(() => null);

    const err = await downloadLogFile(datakit);

    expect(err).toBe('download log failed');
  });
});
