export {};

const mockInjectEndpoints = jest.fn();
const mockBuilder = {
  query: jest.fn((config) => config),
  mutation: jest.fn((config) => config),
};

jest.mock('./baseApi', () => ({
  baseApi: {
    injectEndpoints: (...args: unknown[]) => mockInjectEndpoints(...args),
  },
}));

describe('consoleApi endpoint definitions', () => {
  beforeEach(() => {
    jest.resetModules();
    mockInjectEndpoints.mockReset();
    mockBuilder.query.mockClear();
    mockBuilder.mutation.mockClear();
    mockInjectEndpoints.mockImplementation(({ endpoints }) => {
      const builtEndpoints = endpoints(mockBuilder);
      return {
        endpoints: builtEndpoints,
      };
    });
  });

  it('defines workspace, account, version and logout endpoints with expected queries', () => {
    const consoleApi = require('./consoleApi');

    expect(mockInjectEndpoints).toHaveBeenCalledWith(expect.objectContaining({
      overrideExisting: false,
      endpoints: expect.any(Function),
    }));

    expect(mockBuilder.query).toHaveBeenCalled();
    expect(mockBuilder.mutation).toHaveBeenCalled();
    expect(consoleApi.default.endpoints).toBeDefined();

    expect(mockBuilder.query.mock.calls[0][0].query()).toEqual({
      url: '/api/console/workspaceList',
    });
    expect(mockBuilder.query.mock.calls[0][0].keepUnusedDataFor).toBe(5);

    expect(mockBuilder.query.mock.calls[1][0].query()).toEqual({
      url: '/api/console/currentWorkspace',
    });
    expect(mockBuilder.query.mock.calls[1][0].transformResponse({
      content: { uuid: 'ws-1' },
    })).toEqual({ uuid: 'ws-1' });

    expect(mockBuilder.mutation.mock.calls[0][0].query('ws-2')).toEqual({
      url: '/api/console/changeWorkspace',
      method: 'post',
      body: {
        workspaceUUID: 'ws-2',
      },
      headers: {
        'X-Workspace-Uuid': 'ws-2',
      },
    });

    expect(mockBuilder.query.mock.calls[2][0].query()).toEqual({
      url: '/api/console/accountPermissions',
    });
    expect(mockBuilder.query.mock.calls[3][0].query()).toEqual({
      url: '/api/console/currentAccount',
    });
    expect(mockBuilder.query.mock.calls[4][0].query()).toEqual({
      url: '/api/lastDatakitVersion',
    });
    expect(mockBuilder.query.mock.calls[5][0].query()).toEqual({
      method: 'post',
      url: '/api/console/logout',
    });
  });
});
