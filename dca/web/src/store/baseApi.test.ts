jest.mock('src/helper/helper', () => ({
  alertError: jest.fn(),
}));

jest.mock('.', () => ({
  clearStore: jest.fn(() => Promise.resolve()),
}));

jest.mock('@reduxjs/toolkit/query/react', () => ({
  fetchBaseQuery: jest.fn(),
  createApi: jest.fn(() => ({})),
}));

const fs = require('fs');
const path = require('path');
export {};

describe('getMsg', () => {
  beforeEach(() => {
    jest.resetModules();
    jest.clearAllMocks();
  });

  it('formats object error messages without rendering [object Object]', () => {
    const { getMsg } = require('./baseApi');
    const msg = getMsg({
      errorCode: 'network.error',
      message: { detail: 'request failed' } as any,
    });

    expect(msg).toContain('request failed');
    expect(msg).not.toContain('[object Object]');
  });
});

describe('fetchWithIntercept', () => {
  beforeEach(() => {
    jest.resetModules();
    jest.clearAllMocks();
  });

  it('shows formatted backend error details when fetchBaseQuery returns error.data', async () => {
    const { alertError } = require('src/helper/helper');
    const { fetchBaseQuery } = require('@reduxjs/toolkit/query/react');
    const backendError = {
      errorCode: 'auth.failed',
      message: { detail: 'token expired' },
    };
    fetchBaseQuery.mockImplementation(() => jest.fn().mockResolvedValue({
      data: undefined,
      error: {
        status: 401,
        data: backendError,
      },
    }));

    const { fetchWithIntercept } = require('./baseApi');

    await expect(fetchWithIntercept('/api/test', {} as any, {} as any)).rejects.toEqual(backendError);
    expect(alertError).toHaveBeenCalledWith(expect.stringContaining('token expired'));
    expect(alertError).not.toHaveBeenCalledWith('Unexpected server error');
  });

  it('falls back to a generic error when neither data nor structured error details exist', async () => {
    const { alertError } = require('src/helper/helper');
    const { fetchBaseQuery } = require('@reduxjs/toolkit/query/react');
    fetchBaseQuery.mockImplementation(() => jest.fn().mockResolvedValue({
      data: undefined,
      error: undefined,
    }));

    const { fetchWithIntercept } = require('./baseApi');

    await expect(fetchWithIntercept('/api/test', {} as any, {} as any)).rejects.toThrow('empty response data');
    expect(alertError).toHaveBeenCalledWith('Unexpected server error');
  });

  it('does not report aborted fetches as unexpected server errors', async () => {
    const { alertError } = require('src/helper/helper');
    const { fetchBaseQuery } = require('@reduxjs/toolkit/query/react');
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
    const abortError = {
      status: 'FETCH_ERROR',
      error: 'AbortError: The user aborted a request.',
    };
    fetchBaseQuery.mockImplementation(() => jest.fn().mockResolvedValue({
      data: undefined,
      error: abortError,
    }));

    const { fetchWithIntercept } = require('./baseApi');

    await expect(fetchWithIntercept('/api/test', {} as any, {} as any)).rejects.toEqual(abortError);
    expect(alertError).not.toHaveBeenCalled();
    expect(consoleSpy).not.toHaveBeenCalled();
    consoleSpy.mockRestore();
  });
});

describe('baseApi logging', () => {
  it('does not log every intercepted response to the browser console', () => {
    const source = fs.readFileSync(path.join(__dirname, 'baseApi.ts'), 'utf8');

    expect(source).not.toContain('console.log(');
  });
});
