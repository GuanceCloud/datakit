import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';

const mockLocation = {
  pathname: '/dashboard/log',
  state: {
    datakit: {
      id: 'dk-1',
      host_name: 'dk-host',
      os: 'linux',
      status: 'running',
      version: '1.0.0',
      run_in_container: false,
    },
  },
};

const mockGetDatakitStat = jest.fn();
const mockUseLazyGetDatakitStatQuery = jest.fn();
const mockUseLazyReloadDatakitQuery = jest.fn();
const mockUseLazyUpgradeDatakitQuery = jest.fn();

jest.mock('react-redux', () => ({
  connect: () => (component: unknown) => component,
}));

jest.mock('react-router-dom', () => ({
  useLocation: () => mockLocation,
  Outlet: () => <div>outlet rendered</div>,
}));

jest.mock('src/hooks', () => ({
  useAppSelector: jest.fn(() => []),
}));

jest.mock('src/store/datakitApi', () => ({
  useLazyGetDatakitStatQuery: (...args: unknown[]) => mockUseLazyGetDatakitStatQuery(...args),
  useLazyReloadDatakitQuery: (...args: unknown[]) => mockUseLazyReloadDatakitQuery(...args),
  useLazyUpgradeDatakitQuery: (...args: unknown[]) => mockUseLazyUpgradeDatakitQuery(...args),
}));

jest.mock('src/helper/helper', () => ({
  alertError: jest.fn(),
  getLatestDatakitVersion: jest.fn((version, latestVersions) => latestVersions[`v${version.split('.')[0]}`]),
  isContainerMode: jest.fn((dk) => !!dk?.run_in_container),
  isDatakitManagement: jest.fn((dk) => dk?.status === 'running'),
  isNewerDatakitVersionAvailable: jest.fn((version, latestVersions) => version !== latestVersions[`v${version.split('.')[0]}`]),
}));

jest.mock('../DatakitInfoNav/DatakitInfoNav', () => ({
  DatakitInfoNav: () => <div>datakit nav</div>,
}));

jest.mock('src/pages/Dashboard/Dashboard', () => {
  const React = require('react');
  return {
    DashboardContext: React.createContext({ latestDatakitVersions: { v1: '1.0.0', v2: '2.0.0' } }),
    getOSIcon: jest.fn(() => ''),
  };
});

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      if (key === 'datakit.operation_disabled.container_reload') return 'container reload disabled';
      if (key === 'datakit.operation_disabled.container_upgrade') return 'container upgrade disabled';
      if (key === 'datakit.operation_disabled.latest_version') return 'latest version disabled';
      if (key === 'datakit.operation_disabled.not_running') return 'not running disabled';
      return key;
    },
  }),
}));

jest.mock('antd', () => {
  const actual = jest.requireActual('antd');
  return {
    ...actual,
    App: {
      useApp: () => ({
        modal: {
          confirm: jest.fn(),
        },
      }),
    },
    Tooltip: ({ children, title }: any) => <div title={title}>{children}</div>,
  };
});

const DkInfo = require('./DkInfo').default;

describe('DkInfo', () => {
  beforeEach(() => {
    mockGetDatakitStat.mockClear();
    mockLocation.pathname = '/dashboard/log';
    mockLocation.state.datakit = {
      id: 'dk-1',
      host_name: 'dk-host',
      os: 'linux',
      status: 'running',
      version: '1.0.0',
      run_in_container: false,
    };
    mockUseLazyGetDatakitStatQuery.mockReturnValue([mockGetDatakitStat, { data: undefined, isFetching: false, isError: true }]);
    mockUseLazyReloadDatakitQuery.mockReturnValue([jest.fn(), { isFetching: false, isError: false }]);
    mockUseLazyUpgradeDatakitQuery.mockReturnValue([jest.fn(), { isFetching: false, isError: false }]);
  });

  it('renders child pages that do not require stats even when stats loading failed', () => {
    render(<DkInfo />);

    expect(screen.getByText('outlet rendered')).toBeInTheDocument();
    expect(screen.queryByText('network_error')).not.toBeInTheDocument();
    expect(mockGetDatakitStat).toHaveBeenCalled();
  });

  it('keeps runinfo blocked when stats are unavailable', async () => {
    mockLocation.pathname = '/dashboard/runinfo';

    render(<DkInfo />);

    await waitFor(() => {
      expect(screen.getByText('network_error')).toBeInTheDocument();
    });
    expect(screen.queryByText('outlet rendered')).not.toBeInTheDocument();
  });

  it('keeps config blocked when stats are unavailable', async () => {
    mockLocation.pathname = '/dashboard/config';

    render(<DkInfo />);

    await waitFor(() => {
      expect(screen.getByText('network_error')).toBeInTheDocument();
    });
    expect(screen.queryByText('outlet rendered')).not.toBeInTheDocument();
  });

  it('explains disabled upgrade and reload actions in the management header', () => {
    mockLocation.pathname = '/dashboard/log';
    mockLocation.state.datakit = {
      id: 'dk-1',
      host_name: 'dk-host',
      os: 'linux',
      status: 'offline',
      version: '1.0.0',
      run_in_container: false,
    };

    render(<DkInfo />);

    expect(screen.getAllByTitle('not running disabled')).toHaveLength(2);
    expect(screen.getByRole('button', { name: /upgrade/i })).toBeDisabled();
    expect(screen.getByRole('button', { name: /reload/i })).toBeDisabled();
  });
});
