import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';

const mockNavigate = jest.fn();
const mockToggleLanguage = jest.fn();
const mockAlertError = jest.fn();
const mockClearStore = jest.fn(() => Promise.resolve());
const mockMessageSuccess = jest.fn();
const mockModalConfirm = jest.fn(({ onOk }) => onOk?.());
const mockSetUserInfo = jest.fn();
const mockGetWorkspaceList = jest.fn();
const mockGetCurrentWorkspace = jest.fn(() => ({
  unwrap: () => Promise.resolve(mockCurrentWorkspace),
}));
const mockGetDatakitVersion = jest.fn();
const mockGetCurrentAccount = jest.fn(() => ({
  unwrap: () => Promise.resolve({
    code: 200,
    content: {
      name: 'demo-user',
      email: 'demo@example.com',
    },
  }),
}));
const mockChangeWorkspace = jest.fn(() => ({
  unwrap: () => Promise.resolve({ code: 200 }),
}));
const mockUserLogout = jest.fn(() => ({
  unwrap: () => Promise.resolve({}),
}));

let mockWorkspaceListData: any = {
  code: 200,
  content: {
    data: [
      { uuid: 'ws-1', name: 'Workspace 1', wsName: 'Workspace 1' },
      { uuid: 'ws-2', name: 'Workspace 2', wsName: 'Workspace 2' },
    ],
  },
};

let mockCurrentWorkspace: any = {
  name: 'Workspace 1',
  wsName: 'Workspace 1',
};

let mockDatakitVersionData: any = {
  code: 200,
  content: {
    v1: { version: '1.94.1' },
    v2: { version: '2.9.0' },
  },
};

jest.mock('react-router-dom', () => ({
  Outlet: () => <div>outlet content</div>,
  useNavigate: () => mockNavigate,
}));

jest.mock('react-redux', () => ({
  connect: () => (Component: any) => Component,
}));

jest.mock('src/store', () => ({
  clearStore: () => mockClearStore(),
}));

jest.mock('src/helper/helper', () => ({
  alertError: (...args: any[]) => mockAlertError(...args),
}));

jest.mock('src/store/consoleApi', () => ({
  useLazyGetWorkspaceListQuery: () => [mockGetWorkspaceList, { data: mockWorkspaceListData }],
  useLazyGetCurrentWorkspaceQuery: () => [mockGetCurrentWorkspace, { data: mockCurrentWorkspace }],
  useLazyGetCurrentAccountQuery: () => [mockGetCurrentAccount],
  useChangeWorkspaceMutation: () => [mockChangeWorkspace],
  useLazyLogoutQuery: () => [mockUserLogout],
  useLazyGetDatakitVersionQuery: () => [mockGetDatakitVersion, { data: mockDatakitVersionData }],
}));

jest.mock('src/store/user/user', () => ({
  set: jest.fn((payload) => payload),
}));

jest.mock('src/i18n', () => ({
  toggleLanguage: () => mockToggleLanguage(),
}));

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: any) => {
      if (key === 'workspace_list') return 'Workspace list';
      if (key === 'search_workspace') return 'Search workspace';
      if (key === 'help') return 'Help';
      if (key === 'language') return 'Language';
      if (key === 'logout') return 'Logout';
      if (key === 'confirm') return 'Confirm';
      if (key === 'cancel') return 'Cancel';
      if (key === 'is_logout') return 'Confirm logout';
      if (key === 'logout_success') return 'Logout success';
      if (key === 'logout_fail') return 'Logout failed';
      if (key === 'no_data') return 'No data';
      return key;
    },
  }),
}));

jest.mock('antd', () => {
  const React = require('react');
  return {
    Divider: () => <span>|</span>,
    Dropdown: ({ children, menu, dropdownRender }: any) => {
      const originNode = (
        <div>
          {menu?.items?.map((item: any) => (
            <div
              key={item.key}
              onClick={() => {
                if (item.disabled) return;
                item.label?.props?.onClick?.() || menu?.onClick?.({ key: item.key });
              }}
            >
              {item.label}
            </div>
          ))}
        </div>
      );
      return (
      <div>
        {children}
        {dropdownRender ? dropdownRender(originNode) : originNode}
      </div>
      );
    },
    message: {
      success: (content: any) => mockMessageSuccess(content),
    },
    Modal: {
      confirm: (config: any) => mockModalConfirm(config),
    },
    Space: ({ children, onClick }: any) => <div onClick={onClick}>{children}</div>,
    Tooltip: ({ children }: any) => <div>{children}</div>,
    Typography: {
      Text: ({ children }: any) => <span>{children}</span>,
    },
  };
});

jest.mock('@ant-design/icons', () => ({
  CaretDownOutlined: () => <span>caret</span>,
  ExclamationCircleOutlined: () => <span>warning</span>,
  LogoutOutlined: () => <span>logout</span>,
  TranslationOutlined: () => <span>translate</span>,
  UserOutlined: () => <span>user</span>,
}));

import Dashboard, { getOSIcon } from './Dashboard';

const DashboardForTest = Dashboard as any;

describe('Dashboard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockGetCurrentWorkspace.mockImplementation(() => ({
      unwrap: () => Promise.resolve(mockCurrentWorkspace),
    }));
    mockGetCurrentAccount.mockImplementation(() => ({
      unwrap: () => Promise.resolve({
        code: 200,
        content: {
          name: 'demo-user',
          email: 'demo@example.com',
        },
      }),
    }));
    mockChangeWorkspace.mockImplementation(() => ({
      unwrap: () => Promise.resolve({ code: 200 }),
    }));
    mockUserLogout.mockImplementation(() => ({
      unwrap: () => Promise.resolve({}),
    }));
    mockWorkspaceListData = {
      code: 200,
      content: {
        data: [
          { uuid: 'ws-1', name: 'Workspace 1', wsName: 'Workspace 1' },
          { uuid: 'ws-2', name: 'Workspace 2', wsName: 'Workspace 2' },
        ],
      },
    };
    mockCurrentWorkspace = {
      name: 'Workspace 1',
      wsName: 'Workspace 1',
    };
    mockDatakitVersionData = {
      code: 200,
      content: {
        v1: { version: '1.94.1' },
        v2: { version: '2.9.0' },
      },
    };
  });

  it('initializes dashboard data and renders workspace, user and outlet content', async () => {
    render(<DashboardForTest user={{ name: '', email: '' } as any} setUserInfo={mockSetUserInfo} />);

    expect(mockGetWorkspaceList).toHaveBeenCalledTimes(1);
    expect(mockGetDatakitVersion).toHaveBeenCalledWith('');
    expect(mockGetCurrentWorkspace).toHaveBeenCalledTimes(1);
    expect(screen.getAllByText('Workspace 1')[0]).toBeInTheDocument();
    expect(screen.getByText('outlet content')).toBeInTheDocument();
    expect(screen.getByText('Help')).toBeInTheDocument();

    await waitFor(() => {
      expect(mockSetUserInfo).toHaveBeenCalledWith({
        name: 'demo-user',
        email: 'demo@example.com',
      });
    });

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
  });

  it('changes language and renders logout entry', () => {
    render(<DashboardForTest user={{ name: 'tester', email: 't@example.com' } as any} setUserInfo={mockSetUserInfo} />);

    fireEvent.click(screen.getByText('Language'));
    expect(mockToggleLanguage).toHaveBeenCalledTimes(1);
    expect(screen.getAllByText('Logout')[0]).toBeInTheDocument();
  });

  it('changes workspace when a workspace menu item is clicked', async () => {
    render(<DashboardForTest user={{ name: 'tester', email: 't@example.com' } as any} setUserInfo={mockSetUserInfo} />);

    fireEvent.click(screen.getAllByText('Workspace 2')[0].closest('div') as Element);

    await waitFor(() => {
      expect(mockChangeWorkspace).toHaveBeenCalledWith('ws-2');
    });
    expect(mockGetWorkspaceList).toHaveBeenCalledWith(undefined, false);
    expect(mockGetCurrentWorkspace).toHaveBeenCalledWith(undefined, false);
  });

  it('filters workspace menu items by keyword', async () => {
    mockWorkspaceListData = {
      code: 200,
      content: {
        data: [
          { uuid: 'ws-1', name: 'Alpha Workspace', wsName: 'Alpha Workspace' },
          { uuid: 'ws-2', name: 'Beta Workspace', wsName: 'Beta Workspace' },
        ],
      },
    };
    render(<DashboardForTest user={{ name: 'tester', email: 't@example.com' } as any} setUserInfo={mockSetUserInfo} />);

    fireEvent.change(screen.getByPlaceholderText('Search workspace'), { target: { value: 'beta' } });

    expect(screen.getByText('Beta Workspace')).toBeInTheDocument();
    expect(screen.queryByText('Alpha Workspace')).not.toBeInTheDocument();
  });

  it('does not change workspace when filtered workspace list is empty', async () => {
    render(<DashboardForTest user={{ name: 'tester', email: 't@example.com' } as any} setUserInfo={mockSetUserInfo} />);

    fireEvent.change(screen.getByPlaceholderText('Search workspace'), { target: { value: 'missing' } });
    fireEvent.click(screen.getByText('No data'));

    expect(mockChangeWorkspace).not.toHaveBeenCalled();
  });

  it('maps os names to icons', () => {
    expect(getOSIcon('linux')).toContain('linux.png');
    expect(getOSIcon('windows')).toContain('windows.png');
    expect(getOSIcon('mac')).toContain('mac.png');
  });
});
