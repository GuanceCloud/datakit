import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';

const mockAlertError = jest.fn();
const mockMessageSuccess = jest.fn();
const mockModalConfirm = jest.fn(({ onOk }) => onOk?.());
const mockModalInfo = jest.fn();
const mockIsAdmin = jest.fn(() => true);
const mockGetDatakitConfig = jest.fn();
const mockSaveDatakitConfig = jest.fn();
const mockDeleteDatakitConfig = jest.fn();
const mockWindowOpen = jest.fn();

jest.mock('react-router-dom', () => ({
  useBlocker: () => ({ state: 'unblocked', proceed: jest.fn(), reset: jest.fn() }),
}));

jest.mock('src/hooks/useIsAdmin', () => ({
  __esModule: true,
  useIsAdmin: () => mockIsAdmin(),
}));

jest.mock('src/helper/helper', () => ({
  __esModule: true,
  alertError: (...args: unknown[]) => mockAlertError(...args),
  isContainerMode: jest.fn(() => false),
}));

jest.mock('../../../api/api', () => ({
  getDatakitConfig: (...args: unknown[]) => mockGetDatakitConfig(...args),
  saveDatakitConfig: (...args: unknown[]) => mockSaveDatakitConfig(...args),
  deleteDatakitConfig: (...args: unknown[]) => mockDeleteDatakitConfig(...args),
}));

jest.mock('src/components/Common/ResizeBar/ResizeBar', () => () => <div>resize-bar</div>);

jest.mock('src/components/DCAEditor/DCAEditor', () => ({
  __esModule: true,
  default: ({ value }: { value: string }) => <pre>{value}</pre>,
}));

jest.mock('src/config', () => ({
  __esModule: true,
  default: {
    docURL: 'https://docs.example.com',
  },
}));

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

jest.mock('../DkInfo', () => {
  const React = require('react');
  return {
    DkInfoContext: React.createContext({
      datakit: undefined,
      datakitStat: undefined,
    }),
  };
});

jest.mock('antd', () => {
  const React = require('react');
  return {
    Button: ({ children, onClick, disabled }: any) => <button disabled={disabled} onClick={onClick}>{children}</button>,
    Input: ({ value, onChange }: any) => <input value={value} onChange={onChange} />,
    message: {
      success: (content: any) => mockMessageSuccess(content),
    },
    Modal: Object.assign(({ open, onOk, onCancel, children }: any) => open ? (
      <div>
        {children}
        <button onClick={onOk}>modal_ok</button>
        <button onClick={onCancel}>modal_cancel</button>
      </div>
    ) : null, {
      confirm: (config: any) => mockModalConfirm(config),
      info: (config: any) => mockModalInfo(config),
    }),
    Space: ({ children }: any) => <div>{children}</div>,
    Tooltip: ({ children }: any) => <div>{children}</div>,
    Typography: {
      Text: ({ children, onClick }: any) => <span onClick={onClick}>{children}</span>,
    },
  };
});

jest.mock('@ant-design/icons', () => ({
  CaretRightOutlined: () => <span>right</span>,
  CaretDownOutlined: () => <span>down</span>,
  ExclamationCircleOutlined: () => <span>confirm</span>,
  EditOutlined: () => <span>edit</span>,
  QuestionCircleFilled: () => <span>question</span>,
}));

const Config = require('./Config').default;
const { DkInfoContext } = require('../DkInfo');

function renderConfig(overrides?: any) {
  const contextValue = {
    datakit: {
      id: 'dk-1',
      os: 'linux',
    },
    datakitStat: {
      os_arch: 'linux/amd64',
      config_info: {
        datakit: {
          sample_config: 'hostname = "demo"',
          config_paths: [{ path: '/usr/local/datakit/datakit.conf', loaded: 1 }],
        },
        inputs: {
          cpu: {
            sample_config: 'interval = "10s"',
            config_dir: '/usr/local/datakit/conf.d',
            catalog: '',
            config_paths: [{ path: '/usr/local/datakit/conf.d/cpu.conf', loaded: 1 }],
          },
          disk: {
            sample_config: 'mount_points = ["/"]',
            config_dir: '/usr/local/datakit/conf.d',
            catalog: 'host',
            config_paths: [],
          },
          self: {
            sample_config: '',
            config_dir: '',
            catalog: '',
            config_paths: [],
          },
        },
      },
      ...(overrides?.datakitStat || {}),
    },
    ...(overrides || {}),
  };

  return render(
    <DkInfoContext.Provider value={contextValue}>
      <Config />
    </DkInfoContext.Provider>
  );
}

describe('Config', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockGetDatakitConfig.mockResolvedValue([null, 'interval = "10s"\n']);
    mockSaveDatakitConfig.mockResolvedValue([null, null]);
    mockDeleteDatakitConfig.mockResolvedValue([null]);
    mockWindowOpen.mockReset();
    window.open = mockWindowOpen as any;
  });

  it('renders configured files and sample list from datakit stats', async () => {
    renderConfig();

    expect(await screen.findByText('datakit.conf')).toBeInTheDocument();
    expect(screen.getAllByText('cpu').length).toBeGreaterThan(0);
    expect(screen.getByText('- cpu.conf')).toBeInTheDocument();
    expect(screen.getByText('disk')).toBeInTheDocument();
    expect(screen.queryByText('self')).not.toBeInTheDocument();
  });

  it('loads config content when selecting an existing config file', async () => {
    renderConfig();

    fireEvent.click(await screen.findAllByText('cpu').then((els) => els[0]));
    fireEvent.click(await screen.findByText('- cpu.conf'));

    await waitFor(() => {
      expect(mockGetDatakitConfig).toHaveBeenCalledWith(
        expect.objectContaining({ id: 'dk-1' }),
        '/usr/local/datakit/conf.d/cpu.conf'
      );
    });
    expect(await screen.findByText(/interval = "10s"/)).toBeInTheDocument();
    expect(screen.getByText('file_path： /usr/local/datakit/conf.d/cpu.conf')).toBeInTheDocument();
  });

  it('switches to new config mode when selecting a sample config', async () => {
    renderConfig();

    fireEvent.click(await screen.findByText('disk'));

    expect(screen.getByText('config.new_config_message')).toBeInTheDocument();
    expect(screen.getByText('mount_points = ["/"]')).toBeInTheDocument();
    expect(screen.getByText('config.configure_sample')).toBeInTheDocument();
  });
});
