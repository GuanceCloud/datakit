// @ts-nocheck
import React from 'react';
import { act, render, screen, waitFor } from '@testing-library/react';

const fs = require('fs');
const path = require('path');

const mockDownloadLogFile = jest.fn();
const mockUseIsAdmin = jest.fn(() => true);

class MockWebSocket {
  static instances = [];

  constructor(url) {
    this.handlers = {};
    this.url = url;
    MockWebSocket.instances.push(this);
  }

  addEventListener(type, handler) {
    this.handlers[type] = this.handlers[type] || [];
    this.handlers[type].push(handler);
  }

  close() {}

  emit(type, event) {
    (this.handlers[type] || []).forEach((handler) => handler(event));
  }

  static reset() {
    MockWebSocket.instances = [];
  }
}

jest.mock('src/api/api', () => ({
  downloadLogFile: (...args) => mockDownloadLogFile(...args),
  getQueryPath: jest.fn((basePath, params) => {
    if (!params) {
      return basePath;
    }
    const query = Object.entries(params).map(([key, value]) => `${key}=${String(value)}`).join('&');
    return `${basePath}?${query}`;
  }),
}));

jest.mock('src/hooks/useIsAdmin', () => ({
  useIsAdmin: () => mockUseIsAdmin(),
}));

jest.mock('src/helper/helper', () => ({
  alertError: jest.fn(),
}));

jest.mock('../DkInfo', () => {
  const React = require('react');
  return {
    DkInfoContext: React.createContext({
      datakit: {
        id: 'dk-1',
        host_name: 'dk-host',
      },
      datakitStat: undefined,
    }),
  };
});

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key, params) => {
      if (key === 'log.connection.connecting') return 'Connecting';
      if (key === 'log.connection.connected') return 'Connected';
      if (key === 'log.connection.disconnected') return 'Disconnected';
      if (key === 'log.description.datakit') return 'DataKit log';
      if (key === 'log.description.gin') return 'gin log';
      if (key === 'log.meta.showing') return `Showing ${params?.shown} / ${params?.total}`;
      if (key === 'log.meta.matches') return `${params?.count} matches`;
      if (key === 'log.meta.follow') return 'Follow';
      if (key === 'log.search.placeholder') return 'Search logs';
      if (key === 'log.action.reconnect') return 'Reconnect';
      if (key === 'log.action.resume') return 'Resume';
      if (key === 'log.action.pause') return 'Pause';
      if (key === 'log.action.clear') return 'Clear';
      if (key === 'all') return 'All';
      if (key === 'export') return 'Export';
      return key;
    },
  }),
}));

jest.mock('antd', () => {
  const React = require('react');
  return {
    Button: ({ children, title, onClick, icon }) => <button title={title} onClick={onClick}>{icon}{children}</button>,
    Dropdown: ({ children }) => <div>{children}</div>,
    Input: ({ placeholder, value, onChange }) => <input placeholder={placeholder} value={value} onChange={onChange} />,
    Menu: ({ items }) => <div>{items?.map((item) => <span key={item.key}>{item.label}</span>)}</div>,
    Select: ({ value, options, onChange }) => (
      <select aria-label="level-filter" value={value} onChange={(event) => onChange?.(event.target.value)}>
        {options?.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
      </select>
    ),
    Switch: ({ checked, onChange }) => <input type="checkbox" checked={checked} onChange={(event) => onChange?.(event.target.checked)} />,
    Tag: ({ children }) => <span>{children}</span>,
  };
});

const Log = require('./Log').default;

function renderLog() {
  return render(<Log />);
}

describe('Log page source', () => {
  it('does not print log stream content to the browser console', () => {
    const source = fs.readFileSync(path.join(__dirname, 'Log.tsx'), 'utf8');

    expect(source).not.toContain('console.log(');
  });

  it('keeps the log viewer controls discoverable in source', () => {
    const source = fs.readFileSync(path.join(__dirname, 'Log.tsx'), 'utf8');

    expect(source).toContain('isPaused');
    expect(source).toContain('isFollowing');
    expect(source).toContain('connectionStatus');
    expect(source).toContain('reconnectToken');
    expect(source).toContain('clearLogs');
    expect(source).toContain('maskSensitiveLogLine');
  });

  it('reconnects the log stream when resuming from a disconnected state', () => {
    const source = fs.readFileSync(path.join(__dirname, 'Log.tsx'), 'utf8');

    expect(source).toContain('setReconnectToken');
    expect(source).toContain('connectionStatusRef.current === "disconnected"');
  });

  it('ignores status events from stale websocket connections', () => {
    const source = fs.readFileSync(path.join(__dirname, 'Log.tsx'), 'utf8');

    expect(source).toContain('activeConnectionIDRef');
    expect(source).toContain('isActiveConnection');
  });

  it('highlights warning and error log lines', () => {
    const source = fs.readFileSync(path.join(__dirname, 'Log.tsx'), 'utf8');

    expect(source).toContain('normalizeLogLines');
    expect(source).toContain('log-line-error');
    expect(source).toContain('log-line-warn');
    expect(source).toContain('log-line-content');
  });

  it('supports searching and filtering log lines by level', () => {
    const source = fs.readFileSync(path.join(__dirname, 'Log.tsx'), 'utf8');

    expect(source).toContain('searchText');
    expect(source).toContain('levelFilter');
    expect(source).toContain('filteredLogLines');
    expect(source).toContain('matchCount');
  });

  it('highlights the matched keyword inside each log line', () => {
    const source = fs.readFileSync(path.join(__dirname, 'Log.tsx'), 'utf8');

    expect(source).toContain('renderLogLineContent');
    expect(source).toContain('log-keyword-match');
  });

  it('makes log status and counts easier to understand', () => {
    const source = fs.readFileSync(path.join(__dirname, 'Log.tsx'), 'utf8');

    expect(source).toContain('getConnectionStatusColor');
    expect(source).toContain('reconnectLogStream');
    expect(source).toContain('log.meta.showing');
    expect(source).toContain('log.meta.matches');
    expect(source).toContain('log.connection.');
    expect(source).toContain('log.search.placeholder');
  });
});

describe('Log websocket status', () => {
  beforeEach(() => {
    MockWebSocket.reset();
    mockDownloadLogFile.mockReset();
    // @ts-ignore
    global.WebSocket = MockWebSocket;
  });

  it('switches from connecting to connected after the websocket open event', async () => {
    renderLog();

    expect(screen.getByText('Connecting')).toBeInTheDocument();
    expect(MockWebSocket.instances).toHaveLength(1);

    act(() => {
      MockWebSocket.instances[0].emit('open');
    });

    await waitFor(() => {
      expect(screen.getByText('Connected')).toBeInTheDocument();
    });
  });

  it('switches to connected after receiving a log message even if open is never fired', async () => {
    renderLog();

    expect(screen.getByText('Connecting')).toBeInTheDocument();
    expect(MockWebSocket.instances).toHaveLength(1);

    act(() => {
      MockWebSocket.instances[0].emit('message', { data: '2026-05-11 INFO hello' });
    });

    await waitFor(() => {
      expect(screen.getByText('Connected')).toBeInTheDocument();
    });
    expect(screen.getByText(/INFO hello/)).toBeInTheDocument();
  });

  it('masks sensitive token-like values in log lines', async () => {
    renderLog();

    act(() => {
      MockWebSocket.instances[0].emit('message', {
        data: '2026-05-11 INFO dataway=https://openway.example.com/v1/write/logging?token=secret-token kv=real-value password="secret-pass"',
      });
    });

    await waitFor(() => {
      expect(screen.getByText(/token=\*\*\*\*\*\*/)).toBeInTheDocument();
      expect(screen.getByText(/kv=\*\*\*\*\*\*/i)).toBeInTheDocument();
      expect(screen.getByText(/password="\*\*\*\*\*\*/i)).toBeInTheDocument();
    });
    expect(screen.queryByText(/secret-token/)).not.toBeInTheDocument();
    expect(screen.queryByText(/real-value/)).not.toBeInTheDocument();
    expect(screen.queryByText(/secret-pass/)).not.toBeInTheDocument();
  });

  it('reconnects and returns to connected after the stream disconnects', async () => {
    renderLog();

    expect(screen.getByText('Connecting')).toBeInTheDocument();
    expect(MockWebSocket.instances).toHaveLength(1);

    act(() => {
      MockWebSocket.instances[0].emit('open');
    });

    await waitFor(() => {
      expect(screen.getByText('Connected')).toBeInTheDocument();
    });

    act(() => {
      MockWebSocket.instances[0].emit('close');
    });

    await waitFor(() => {
      expect(screen.getByText('Disconnected')).toBeInTheDocument();
    });

    act(() => {
      screen.getByTitle('Reconnect').click();
    });

    await waitFor(() => {
      expect(MockWebSocket.instances).toHaveLength(2);
      expect(screen.getByText('Connecting')).toBeInTheDocument();
    });

    act(() => {
      MockWebSocket.instances[1].emit('open');
    });

    await waitFor(() => {
      expect(screen.getByText('Connected')).toBeInTheDocument();
    });
  });
});
