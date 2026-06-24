// @ts-nocheck
import React from 'react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';

const mockNavigate = jest.fn();
const mockAlertError = jest.fn();
const mockUpdateDatakits = jest.fn();
const mockQueryDatakitList = jest.fn();
const mockReloadDatakit = jest.fn();
const mockUpgradeDatakit = jest.fn();
const mockGetDatakitListByID = jest.fn();
const mockOperateDatakit = jest.fn();
const mockGetSearchValue = jest.fn();
const mockMessageSuccess = jest.fn();
const mockModalConfirm = jest.fn();
let mockDidInitColumns = false;

let mockDatakits = [
  {
    id: 'dk-1',
    host_name: 'host-1',
    ip: '10.0.0.1',
    os: 'linux',
    arch: 'amd64',
    status: 'running',
    version: '1.0.0',
    run_in_container: false,
    updated_at: Date.now(),
    global_host_tags_string: '{"env":"prod"}',
  },
];

let mockDatakitListResponse = {
  success: true,
  message: '',
  content: {
    data: mockDatakits,
    pageInfo: {
      count: 1,
      pageIndex: 1,
      pageSize: 20,
      totalCount: 3,
    },
  },
};

let mockSearchValueResponse = {
  success: true,
  content: {
    host_name: ['host-1'],
  },
};

jest.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

jest.mock('react-redux', () => ({
  connect: () => (Component) => Component,
}));

jest.mock('src/helper/helper', () => ({
  alertError: (...args) => mockAlertError(...args),
  isContainerMode: (dk) => !!dk?.run_in_container,
  isDatakitManagement: (dk) => dk?.status === 'running',
  isDatakitUpgradeable: (dk, latestVersion) => dk?.status === 'running' && dk?.version !== latestVersion && !dk?.run_in_container,
  isLoadingStatus: (dk) => ['upgrading', 'restarting'].includes(dk?.status),
  runJob: jest.fn((_limit, arr, fn) => Promise.all(arr.map(fn))),
}));

jest.mock('src/hooks', () => ({
  useAppSelector: (selector) => selector({ datakit: { value: mockDatakits } }),
}));

jest.mock('src/pages/Dashboard/Dashboard', () => {
  const React = require('react');
  return {
    DashboardContext: React.createContext({
      currentWorkspace: { uuid: 'ws-1' },
      latestDatakitVersion: '2.0.0',
    }),
    getOSIcon: () => 'linux.png',
  };
});

jest.mock('src/store/datakitApi', () => ({
  useLazyGetDatakitListQuery: () => [mockQueryDatakitList, { currentData: mockDatakitListResponse, isFetching: false, isError: false }],
  useLazyReloadDatakitQuery: () => [mockReloadDatakit],
  useLazyUpgradeDatakitQuery: () => [mockUpgradeDatakit],
  useLazyGetDatakitListByIDQuery: () => [mockGetDatakitListByID],
  useLazyOperateDatakitQuery: () => [mockOperateDatakit],
  useLazyGetSearchValueQuery: () => [mockGetSearchValue, { currentData: mockSearchValueResponse, isError: false }],
}));

jest.mock('../DatakitStatus/DatakitStatus', () => ({ datakit }) => <span>{datakit.status}</span>);
jest.mock('./AdditionalColumnOptions/AdditionColumnOptions', () => ({
  AdditionColumnOptions: ({ onValueChange }) => {
    const React = require('react');
    React.useEffect(() => {
      if (!mockDidInitColumns) {
        mockDidInitColumns = true;
        onValueChange?.({
          host_name: true,
          ip: true,
          os_arch: true,
          status_text: true,
          uptime: true,
          environment: true,
          last_update: true,
          is_container_running: true,
          datakit_version: true,
          operation: true,
        });
      }
    }, [onValueChange]);
    return <button>display_column</button>;
  },
}));

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key, params) => {
      if (key === 'total_datakit') return `total:${params?.count}`;
      if (key === 'selected_num') return `selected:${params?.num}`;
      if (key === 'reload_datakit_message') return `reload:${params?.count}`;
      if (key === 'datakit.operation_disabled.container_reload') return 'container reload disabled';
      if (key === 'datakit.operation_disabled.container_upgrade') return 'container upgrade disabled';
      if (key === 'datakit.operation_disabled.latest_version') return 'latest version disabled';
      if (key === 'datakit.operation_disabled.not_running') return 'not running disabled';
      return key;
    },
  }),
}));

jest.mock('antd', () => {
  const React = require('react');
  return {
    App: {
      useApp: () => ({
        modal: {
          confirm: (...args) => mockModalConfirm(...args),
        },
      }),
    },
    Avatar: ({ src }) => <img alt="avatar" src={src} />,
    Button: ({ children, onClick, disabled }) => <button disabled={disabled} onClick={onClick}>{children}</button>,
    Checkbox: ({ checked, onChange }) => <input aria-label="checkbox" type="checkbox" checked={checked} onChange={(e) => onChange?.(e)} />,
    Input: ({ value, onChange, onPressEnter, placeholder }) => (
      <input
        placeholder={placeholder}
        value={value}
        onChange={onChange}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            onPressEnter?.(e);
          }
        }}
      />
    ),
    Modal: ({ open, children }) => open ? <div>{children}</div> : null,
    Select: ({ onChange, options = [], value }) => (
      <select value={value} onChange={(e) => onChange?.(e.target.value)}>
        {options.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
      </select>
    ),
    Space: ({ children }) => <div>{children}</div>,
    Spin: () => <div>loading</div>,
    Table: ({ dataSource, rowSelection, columns = [] }) => (
      <div>
        {rowSelection?.columnTitle}
        <div>
          {columns.map((column) => (
            <span key={String(column.key || column.dataIndex || column.title)}>{column.title}</span>
          ))}
        </div>
        {dataSource.map((item) => (
          <div key={item.id}>
            {columns.map((column) => {
              const rawValue = column.dataIndex ? item[column.dataIndex] : undefined;
              const content = column.render ? column.render(rawValue, item, 0) : rawValue;
              return (
                <div key={String(column.key || column.dataIndex || Math.random())}>
                  {content}
                </div>
              );
            })}
          </div>
        ))}
      </div>
    ),
    Tooltip: ({ children, title }) => <div title={title}>{children}</div>,
    Typography: {
      Text: ({ children, onClick, disabled }) => <span aria-disabled={disabled} onClick={onClick}>{children}</span>,
    },
    message: {
      success: (...args) => mockMessageSuccess(...args),
    },
  };
});

jest.mock('@ant-design/icons', () => ({
  CloseCircleOutlined: ({ onClick }) => <button onClick={onClick}>remove</button>,
  CopyOutlined: () => <span>copy</span>,
  FilterFilled: () => <span>filter</span>,
  LoadingOutlined: () => <span>loading</span>,
  PlusOutlined: () => <span>plus</span>,
  SearchOutlined: () => <span>search</span>,
  SelectOutlined: () => <span>select</span>,
}));

import DkList from './DkList';

describe('DkList', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockDidInitColumns = false;
    mockDatakits = [
      {
        id: 'dk-1',
        host_name: 'host-1',
        ip: '10.0.0.1',
        os: 'linux',
        arch: 'amd64',
        status: 'running',
        version: '1.0.0',
        run_in_container: false,
        updated_at: Date.now(),
        global_host_tags_string: '{"env":"prod"}',
      },
    ];
    mockDatakitListResponse = {
      success: true,
      message: '',
      content: {
        data: mockDatakits,
        pageInfo: {
          count: 1,
          pageIndex: 1,
          pageSize: 20,
          totalCount: 3,
        },
      },
    };
    mockSearchValueResponse = {
      success: true,
      content: {
        host_name: ['host-1'],
      },
    };
    mockOperateDatakit.mockImplementation(() => ({
      unwrap: () => Promise.resolve({ success: true }),
    }));
    mockReloadDatakit.mockImplementation(() => ({
      unwrap: () => Promise.resolve({ success: true }),
    }));
    mockUpgradeDatakit.mockImplementation(() => ({
      unwrap: () => Promise.resolve({ success: true }),
    }));
    mockGetDatakitListByID.mockImplementation(() => ({
      unwrap: () => Promise.resolve({ success: true, content: [] }),
    }));
  });

  it('initializes list queries and updates datakits from successful responses', async () => {
    render(<DkList updateDatakits={mockUpdateDatakits} />);

    expect(mockQueryDatakitList).toHaveBeenCalled();
    expect(mockGetSearchValue).toHaveBeenCalled();
    expect(screen.getByText('total:3')).toBeInTheDocument();
    expect(screen.getByText('host-1')).toBeInTheDocument();

    await waitFor(() => {
      expect(mockUpdateDatakits).toHaveBeenCalledWith([
        expect.objectContaining({
          id: 'dk-1',
          global_host_tags: { env: 'prod' },
        }),
      ]);
    });
  });

  it('clears the list and reports backend errors when list loading fails', async () => {
    mockDatakitListResponse = {
      success: false,
      message: 'load failed',
      content: {
        data: [],
      },
    };

    render(<DkList updateDatakits={mockUpdateDatakits} />);

    await waitFor(() => {
      expect(mockAlertError).toHaveBeenCalledWith('load failed');
    });
    expect(mockUpdateDatakits).toHaveBeenCalledWith([]);
  });

  it('supports batch selection state and refresh actions', async () => {
    render(<DkList updateDatakits={mockUpdateDatakits} />);

    const initialListCalls = mockQueryDatakitList.mock.calls.length;
    const initialSearchCalls = mockGetSearchValue.mock.calls.length;

    fireEvent.click(screen.getByText('batch_operation'));
    fireEvent.click(screen.getByText('select_all'));
    expect(screen.getByText('selected:3')).toBeInTheDocument();
    expect(screen.getByText('deselect_all')).toBeInTheDocument();

    await waitFor(() => {
      fireEvent.click(screen.getByText('refresh'));
    });

    expect(mockQueryDatakitList.mock.calls.length).toBeGreaterThan(initialListCalls);
    expect(mockGetSearchValue.mock.calls.length).toBeGreaterThan(initialSearchCalls);
  });

  it('searches again when pressing enter in the search input', async () => {
    render(<DkList updateDatakits={mockUpdateDatakits} />);

    const initialListCalls = mockQueryDatakitList.mock.calls.length;
    const searchInput = screen.getByPlaceholderText('search_host_ip');

    fireEvent.change(searchInput, { target: { value: 'host-1' } });
    fireEvent.keyDown(searchInput, { key: 'Enter', code: 'Enter' });

    await waitFor(() => {
      expect(mockQueryDatakitList.mock.calls.length).toBeGreaterThan(initialListCalls);
    });
  });

  it('navigates to the runinfo page from the management action', async () => {
    render(<DkList updateDatakits={mockUpdateDatakits} />);

    fireEvent.click(screen.getByRole('button', { name: 'management' }));

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard/runinfo', {
      state: {
        datakit: expect.objectContaining({ id: 'dk-1', host_name: 'host-1' }),
      },
    });
  });

  it('reloads a single datakit and updates its status on success', async () => {
    render(<DkList updateDatakits={mockUpdateDatakits} />);

    fireEvent.click(screen.getByRole('button', { name: 'reload' }));
    expect(mockModalConfirm).toHaveBeenCalled();

    await act(async () => {
      await mockModalConfirm.mock.calls[0][0].onOk();
    });

    await waitFor(() => {
      expect(mockReloadDatakit).toHaveBeenCalledWith(expect.objectContaining({ id: 'dk-1' }));
    });
    expect(mockMessageSuccess).toHaveBeenCalledWith('reload_datakit_success');
    expect(mockUpdateDatakits).toHaveBeenCalledWith([
      expect.objectContaining({
        id: 'dk-1',
        status: 'restarting',
      }),
    ]);
  });

  it('upgrades a single datakit and updates its status on success', async () => {
    render(<DkList updateDatakits={mockUpdateDatakits} />);

    fireEvent.click(screen.getByRole('button', { name: 'upgrade' }));
    expect(mockModalConfirm).toHaveBeenCalled();

    await act(async () => {
      await mockModalConfirm.mock.calls[0][0].onOk();
    });

    await waitFor(() => {
      expect(mockUpgradeDatakit).toHaveBeenCalledWith(expect.objectContaining({ id: 'dk-1' }));
    });
    expect(mockMessageSuccess).toHaveBeenCalledWith('upgrade_datakit_success');
    expect(mockUpdateDatakits).toHaveBeenCalledWith([
      expect.objectContaining({
        id: 'dk-1',
        status: 'upgrading',
      }),
    ]);
  });

  it('explains why reload and upgrade actions are disabled', async () => {
    mockDatakits = [
      {
        id: 'dk-1',
        host_name: 'host-1',
        ip: '10.0.0.1',
        os: 'linux',
        arch: 'amd64',
        status: 'running',
        version: '1.0.0',
        run_in_container: true,
        updated_at: Date.now(),
        global_host_tags_string: '{}',
      },
    ];
    mockDatakitListResponse = {
      success: true,
      message: '',
      content: {
        data: mockDatakits,
        pageInfo: {
          count: 1,
          pageIndex: 1,
          pageSize: 20,
          totalCount: 1,
        },
      },
    };

    render(<DkList updateDatakits={mockUpdateDatakits} />);

    expect(screen.getByTitle('container reload disabled')).toBeInTheDocument();
    expect(screen.getByTitle('container upgrade disabled')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'reload' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'upgrade' })).toBeDisabled();
  });
});
