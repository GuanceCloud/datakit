export {};

const mockBuilder = {
  query: jest.fn((config) => config),
};

jest.mock('./baseApi', () => ({
  baseApi: {
    injectEndpoints: jest.fn(() => ({
      useLazyGetDatakitStatQuery: jest.fn(),
      useLazyGetFilterQuery: jest.fn(),
      useLazyReloadDatakitQuery: jest.fn(),
      useLazyUpgradeDatakitQuery: jest.fn(),
      useLazyGetDatakitListQuery: jest.fn(),
      useLazyGetDatakitListByIDQuery: jest.fn(),
      useLazyOperateDatakitQuery: jest.fn(),
      useLazyGetSearchValueQuery: jest.fn(),
    })),
  },
}));

let mockInjectEndpoints: jest.Mock;

describe('datakitApi url builder', () => {
  beforeEach(() => {
    jest.resetModules();
    mockInjectEndpoints = require('./baseApi').baseApi.injectEndpoints;
    mockInjectEndpoints.mockReset();
    mockBuilder.query.mockClear();
    mockInjectEndpoints.mockImplementation(({ endpoints }) => {
      const builtEndpoints = endpoints(mockBuilder);
      return {
        endpoints: builtEndpoints,
      };
    });
  });

  it('encodes datakit id query value', () => {
    const { buildDatakitURL } = require('./datakitApi');
    const url = buildDatakitURL('/api/datakit/stats', {
      datakit_id: 'dk&bad=1',
    });

    expect(url).toBe('/api/datakit/stats?datakit_id=dk%26bad%3D1');
    expect(url).not.toContain('dk&bad=1');
  });

  it('encodes list query values with special characters', () => {
    const { buildDatakitURL } = require('./datakitApi');
    const url = buildDatakitURL('/api/datakit/list', {
      pageIndex: 1,
      pageSize: 10,
      search: 'prod&debug=true',
      filter: 'region=cn #1',
      minLastUpdateTime: 1710000000000,
    });

    expect(url).toBe('/api/datakit/list?pageIndex=1&pageSize=10&search=prod%26debug%3Dtrue&filter=region%3Dcn%20%231&minLastUpdateTime=1710000000000');
    expect(url).not.toContain('search=prod&debug=true');
  });

  it('encodes raw filter JSON exactly once', () => {
    const { buildDatakitURL } = require('./datakitApi');
    const filter = JSON.stringify({
      relation: 'and',
      items: [{ field: 'env', operator: 'in', value: ['prod'] }],
    });

    const url = buildDatakitURL('/api/datakit/list', {
      pageIndex: 1,
      pageSize: 10,
      filter,
    });

    expect(url).toContain('filter=%7B%22relation%22%3A%22and%22');
    expect(url).not.toContain('%257B');
    expect(decodeURIComponent(new URLSearchParams(url.split('?')[1]).get('filter') || '')).toBe(filter);
  });

  it('skips empty optional query values', () => {
    const { buildDatakitURL } = require('./datakitApi');
    const url = buildDatakitURL('/api/datakit/list', {
      pageIndex: 1,
      pageSize: 10,
      search: '',
      filter: undefined,
    });

    expect(url).toBe('/api/datakit/list?pageIndex=1&pageSize=10');
  });

  it('defines datakit endpoints with encoded query builders', () => {
    require('./datakitApi');

    expect(mockInjectEndpoints).toHaveBeenCalledWith(expect.objectContaining({
      overrideExisting: false,
      endpoints: expect.any(Function),
    }));

    expect(mockBuilder.query.mock.calls[0][0].query({ id: 'dk&1' })).toEqual({
      url: '/api/datakit/stats?datakit_id=dk%261',
    });

    expect(mockBuilder.query.mock.calls[1][0].query({
      pageIndex: 2,
      pageSize: 20,
      search: 'prod&1',
      filter: 'region=cn #1',
      minLastUpdateTime: 123,
    })).toEqual({
      url: '/api/datakit/list?pageIndex=2&pageSize=20&search=prod%261&filter=region%3Dcn%20%231&minLastUpdateTime=123',
      method: 'get',
    });
    expect(mockBuilder.query.mock.calls[1][0].keepUnusedDataFor).toBe(0);

    expect(mockBuilder.query.mock.calls[2][0].query({ ids: 'dk-1,dk&2' })).toEqual({
      url: '/api/datakit/listByID?ids=dk-1%2Cdk%262',
      method: 'get',
    });

    expect(mockBuilder.query.mock.calls[3][0].query()).toEqual({
      url: '/api/datakit/searchValue',
      method: 'get',
    });

    expect(mockBuilder.query.mock.calls[5][0].query({ id: 'dk&1' })).toEqual({
      url: '/api/datakit/filter?datakit_id=dk%261',
    });

    expect(mockBuilder.query.mock.calls[6][0].query({ id: 'dk&1' })).toEqual({
      url: '/api/datakit/restart?datakit_id=dk%261',
      method: 'PUT',
    });

    expect(mockBuilder.query.mock.calls[7][0].query({ id: 'dk&1' })).toEqual({
      url: '/api/datakit/upgrade?datakit_id=dk%261',
      method: 'POST',
    });

    expect(mockBuilder.query.mock.calls[8][0].query({ ids: 'dk-1,dk&2', type: 'upgrade/reload' })).toEqual({
      url: '/api/datakit/operation/upgrade%2Freload?ids=dk-1%2Cdk%262',
      method: 'POST',
    });
  });
});
