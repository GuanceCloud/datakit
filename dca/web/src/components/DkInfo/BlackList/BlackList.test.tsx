import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';

const mockAlertError = jest.fn();
const mockGetFilter = jest.fn();
const mockWindowOpen = jest.fn();

let mockFilterState: any = {
  data: undefined,
  isLoading: false,
  isError: false,
};
let mockConsoleError: jest.SpyInstance;

jest.mock('src/helper/helper', () => ({
  __esModule: true,
  alertError: (...args: unknown[]) => mockAlertError(...args),
}));

jest.mock('src/store/datakitApi', () => ({
  useLazyGetFilterQuery: () => [mockGetFilter, mockFilterState],
}));

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
    }),
    Nodata: ({ loading, isError, refresh }: any) => (
      <div>
        <span>{loading ? 'loading' : isError ? 'error' : 'empty'}</span>
        <button onClick={refresh}>refresh</button>
      </div>
    ),
  };
});

jest.mock('antd', () => ({
  Button: ({ children, onClick }: any) => <button onClick={onClick}>{children}</button>,
}));

const BlackList = require('./BlackList').default;
const { DkInfoContext } = require('../DkInfo');

function renderBlackList(datakit?: any) {
  return render(
    <DkInfoContext.Provider value={{ datakit: datakit || { id: 'dk-1' } }}>
      <BlackList />
    </DkInfoContext.Provider>
  );
}

describe('BlackList', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockConsoleError = jest.spyOn(console, 'error').mockImplementation(() => undefined);
    mockFilterState = {
      data: undefined,
      isLoading: false,
      isError: false,
    };
    window.open = mockWindowOpen as any;
  });

  afterEach(() => {
    mockConsoleError.mockRestore();
  });

  it('requests blacklist data on mount and shows formatted content', async () => {
    mockFilterState = {
      data: {
        code: 200,
        content: {
          content: '{"rules":["cpu","disk"]}',
          filePath: '/usr/local/datakit/blacklist.json',
        },
      },
      isLoading: false,
      isError: false,
    };

    renderBlackList();

    expect(mockGetFilter).toHaveBeenCalledWith(expect.objectContaining({ id: 'dk-1' }));
    expect(await screen.findByText('file_path： /usr/local/datakit/blacklist.json')).toBeInTheDocument();
    expect(screen.getByText((content) => content.includes('"rules"') && content.includes('"cpu"') && content.includes('"disk"'))).toBeInTheDocument();
  });

  it('keeps raw content when blacklist response is not valid json', async () => {
    mockFilterState = {
      data: {
        code: 200,
        content: {
          content: 'not-json-content',
          filePath: '/usr/local/datakit/blacklist.json',
        },
      },
      isLoading: false,
      isError: false,
    };

    renderBlackList();

    expect(await screen.findByText('not-json-content')).toBeInTheDocument();
  });

  it('shows nodata state, supports refresh, and reports error responses', async () => {
    mockFilterState = {
      data: {
        code: 500,
        message: 'bad response',
      },
      isLoading: false,
      isError: true,
    };

    renderBlackList();

    await waitFor(() => {
      expect(mockAlertError).toHaveBeenCalledWith(mockFilterState.data);
    });
    expect(screen.getByText('error')).toBeInTheDocument();

    fireEvent.click(screen.getByText('refresh'));
    expect(mockGetFilter).toHaveBeenCalledTimes(2);
  });

  it('opens help documentation', async () => {
    mockFilterState = {
      data: {
        code: 200,
        content: {
          content: '{"rules":[]}',
          filePath: '/usr/local/datakit/blacklist.json',
        },
      },
      isLoading: false,
      isError: false,
    };

    renderBlackList();

    fireEvent.click(await screen.findByText('help'));
    expect(mockWindowOpen).toHaveBeenCalledWith('https://docs.example.com/management/overall-blacklist/');
  });
});
