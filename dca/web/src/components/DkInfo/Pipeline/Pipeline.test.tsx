import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';

const mockAlertError = jest.fn();
const mockMessageSuccess = jest.fn();
const mockModalConfirm = jest.fn(({ onOk }) => onOk?.());
const mockValidateFields = jest.fn();
const mockIsAdmin = jest.fn(() => true);
const mockGetPipelineList = jest.fn();
const mockGetPipelineDetail = jest.fn();
const mockCreatePipeline = jest.fn();
const mockDeletePipeline = jest.fn();
const mockUpdatePipeline = jest.fn();

jest.mock('react-router-dom', () => ({
  useBlocker: () => ({ state: 'unblocked' }),
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

jest.mock('src/api/api', () => ({
  getPipelineList: (...args: unknown[]) => mockGetPipelineList(...args),
  getPipelineDetail: (...args: unknown[]) => mockGetPipelineDetail(...args),
  createPipeline: (...args: unknown[]) => mockCreatePipeline(...args),
  deletePipeline: (...args: unknown[]) => mockDeletePipeline(...args),
  updatePipeline: (...args: unknown[]) => mockUpdatePipeline(...args),
}));

jest.mock('./PipelineTest/PipelineTest', () => () => <div>pipeline test panel</div>);

jest.mock('src/components/DCAEditor/DCAEditor', () => ({
  __esModule: true,
  default: ({ value }: { value: string }) => <pre>{value}</pre>,
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
    }),
  };
});

jest.mock('antd', () => {
  const React = require('react');
  return {
    Button: ({ children, onClick, disabled }: any) => <button disabled={disabled} onClick={onClick}>{children}</button>,
    Dropdown: ({ children, menu }: any) => (
      <div>
        {children}
        <div>
          {(menu?.items || []).map((item: any) => (
            <button key={item.key} onClick={() => menu.onClick?.({ key: item.key })}>
              {item.label}
            </button>
          ))}
        </div>
      </div>
    ),
    Form: Object.assign(({ children }: any) => <div>{children}</div>, {
      useForm: () => [
        {
          validateFields: (...args: unknown[]) => mockValidateFields(...args),
        },
      ],
      Item: ({ children, help }: any) => <div>{children}{help ? <span>{help}</span> : null}</div>,
    }),
    Input: ({ value, onChange }: any) => <input value={value} onChange={onChange} />,
    Modal: Object.assign(({ open, onOk, onCancel, children }: any) => open ? (
      <div>
        {children}
        <button onClick={onOk}>modal_ok</button>
        <button onClick={onCancel}>modal_cancel</button>
      </div>
    ) : null, {
      confirm: (config: any) => mockModalConfirm(config),
    }),
    Space: ({ children }: any) => <div>{children}</div>,
    Tooltip: ({ children }: any) => <div>{children}</div>,
    Typography: {
      Text: ({ children, onClick }: any) => <span onClick={onClick}>{children}</span>,
    },
    message: {
      success: (content: any) => mockMessageSuccess(content),
    },
  };
});

jest.mock('@ant-design/icons', () => ({
  CopyOutlined: () => <span>copy</span>,
  DownOutlined: () => <span>down</span>,
  EditOutlined: () => <span>edit</span>,
  ExclamationCircleOutlined: () => <span>confirm</span>,
  PlusCircleOutlined: () => <span>plus</span>,
  QuestionCircleOutlined: () => <span>question</span>,
  SaveOutlined: () => <span>save</span>,
  SnippetsOutlined: () => <span>snippets</span>,
  ToolOutlined: () => <span>tool</span>,
}));

const Pipeline = require('./Pipeline').default;
const { DkInfoContext } = require('../DkInfo');

function renderPipeline(datakit?: any) {
  return render(
    <DkInfoContext.Provider value={{ datakit: datakit || { id: 'dk-1', os: 'linux' } }}>
      <Pipeline />
    </DkInfoContext.Provider>
  );
}

describe('Pipeline', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockGetPipelineList.mockResolvedValue([null, [
      { fileName: 'default.p', category: '', content: 'default-content', fileDir: '/usr/local/datakit/pipeline' },
      { fileName: 'metric.p', category: 'metric', content: 'metric-content', fileDir: '/usr/local/datakit/pipeline' },
    ]]);
    mockGetPipelineDetail.mockResolvedValue([null, {
      content: 'selected pipeline content',
      path: '/usr/local/datakit/pipeline/default.p',
    }]);
    mockCreatePipeline.mockResolvedValue([null, {
      fileName: 'new-file.p',
      fileDir: '/usr/local/datakit/pipeline',
      category: 'default',
    }]);
    mockDeletePipeline.mockResolvedValue([null]);
    mockUpdatePipeline.mockResolvedValue([null]);
    mockValidateFields.mockResolvedValue({ fileName: 'new-file.p' });
  });

  it('loads pipeline files and shows selected file details', async () => {
    renderPipeline();

    expect(mockGetPipelineList).toHaveBeenCalledWith(expect.objectContaining({ id: 'dk-1' }));
    expect(await screen.findByText('default.p')).toBeInTheDocument();

    fireEvent.click(screen.getByText('default.p'));

    await waitFor(() => {
      expect(mockGetPipelineDetail).toHaveBeenCalledWith(expect.objectContaining({ id: 'dk-1' }), 'default.p', '');
    });
    expect(await screen.findByText('selected pipeline content')).toBeInTheDocument();
    expect(screen.getByText(/file_path： \/usr\/local\/datakit\/pipeline\/default.p/)).toBeInTheDocument();
  });

  it('switches category and shows files from the selected category', async () => {
    renderPipeline();

    expect(await screen.findByText('default.p')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'metric' }));

    expect(await screen.findByText('metric.p')).toBeInTheDocument();
  });

  it('reports pipeline detail loading errors', async () => {
    mockGetPipelineDetail.mockResolvedValue(['detail failed', null]);

    renderPipeline();

    fireEvent.click(await screen.findByText('default.p'));

    await waitFor(() => {
      expect(mockAlertError).toHaveBeenCalledWith('detail failed');
    });
    expect(screen.queryByText('selected pipeline content')).not.toBeInTheDocument();
  });
});
