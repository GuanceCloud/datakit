import React from 'react';

const mockCreateBrowserRouter = jest.fn((routes) => ({ routes }));
const mockUseRoutes = jest.fn((_routes) => <div>mock routes</div>);

jest.mock('react-router-dom', () => ({
  createBrowserRouter: (routes: any) => mockCreateBrowserRouter(routes),
  Navigate: ({ to }: any) => <div>{to}</div>,
  useRoutes: (routes: any) => mockUseRoutes(routes),
}));

jest.mock('../pages/Dashboard/Dashboard', () => () => <div>dashboard</div>);
jest.mock('src/components/DkInfo/DkInfo', () => () => <div>dkinfo</div>);
jest.mock('src/components/DkInfo/RunInfo/RunInfo', () => () => <div>runinfo</div>);
jest.mock('src/components/DatakitList/DkList', () => () => <div>list</div>);
jest.mock('src/components/DkInfo/Pipeline/Pipeline', () => () => <div>pipeline</div>);
jest.mock('src/components/DkInfo/Config/Config', () => () => <div>config</div>);
jest.mock('src/components/DkInfo/Log/Log', () => () => <div>log</div>);
jest.mock('src/components/DkInfo/BlackList/BlackList', () => () => <div>blacklist</div>);
jest.mock('src/components/Login/Login', () => () => <div>login</div>);

describe('router', () => {
  beforeEach(() => {
    jest.resetModules();
    mockCreateBrowserRouter.mockClear();
    mockUseRoutes.mockClear();
  });

  it('registers dashboard and fallback routes', () => {
    require('./index');

    expect(mockCreateBrowserRouter).toHaveBeenCalledTimes(1);
    const root = mockCreateBrowserRouter.mock.calls[0][0][0];
    expect(root.path).toBe('*');
  });
});
