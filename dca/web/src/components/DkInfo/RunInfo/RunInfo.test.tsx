import React from 'react';
import { render, screen } from '@testing-library/react';

jest.mock('antd', () => ({
  Space: ({ children }: any) => <div>{children}</div>,
  Table: ({ dataSource = [] }: any) => (
    <div>
      {dataSource.map((item: any, index: number) => (
        <pre key={item.input || item.name || index}>{JSON.stringify(item)}</pre>
      ))}
    </div>
  ),
}));

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const translations: Record<string, string> = {
        no: 'No',
        yes: 'Yes',
        status_text: 'Status',
        'run_info.overview': 'Overview',
        'run_info.stats_realtime': 'Live DataKit stats',
        'run_info.basic_info': 'Basic information',
        'run_info.enabled_collectors': 'Enabled collectors',
        'run_info.collector_metrics': 'Collector runtime metrics',
        'run_info.inputs_count': '{{count}} inputs',
        'run_info.instances_count': '{{count}} instances',
        'run_info.crash_count': '{{count}} crashes in total',
        'run_info.crashed_count': '{{count}} crashed',
        'run_info.no_crash_record': 'No crash records',
        'run_info.no_enabled_collectors': 'No enabled collectors',
        'run_info.normal': 'Normal',
        'run_info.last_feed_zero_time_filtered': 'LastFeed hides invalid zero timestamps',
      };

      return (translations[key] || key).replace(/\{\{(\w+)\}\}/g, (_, name) => String(params?.[name] ?? ''));
    },
  }),
}));

jest.mock('src/helper/helper', () => ({
  showDuration: (value: string) => `duration:${value}`,
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

const RunInfo = require('./RunInfo').default;
const { DkInfoContext } = require('../DkInfo');

function renderRunInfo(contextValue: any) {
  return render(
    <DkInfoContext.Provider value={contextValue}>
      <RunInfo />
    </DkInfoContext.Provider>
  );
}

describe('RunInfo', () => {
  it('shows no data when datakit stat is unavailable', () => {
    renderRunInfo({
      datakit: {
        version: '1.0.0',
      },
      datakitStat: undefined,
    });

    expect(screen.getByText('no data')).toBeInTheDocument();
  });

  it('renders transformed datakit runtime information', () => {
    renderRunInfo({
      datakit: {
        version: '1.2.3',
        ip: '10.0.0.1',
        runtime_id: 'rt-1',
        run_in_container: false,
        status: 'running',
      },
      datakitStat: {
        hostname: 'dk-host',
        os_arch: 'linux/amd64',
        uptime: '1.23456789s',
        resource_limit: 8,
        open_files: 12,
        elected: true,
        usage_trace: {
          run_mode: 'standalone',
          usage_cores: 2,
        },
        golang_runtime: {
          goroutines: 5,
          total_sys: 2048,
          heap_alloc: 1024,
        },
        datakit_runtime_info: {
          cpu_usage: 3.5,
        },
        enabled_input_list: {
          cpu: {
            input: 'cpu',
            instances: 2,
            panic: 1,
          },
        },
        inputs_status: {
          cpu: {
            avg_collect_cost: '10ms',
            avg_size: 10,
            category: 'metric',
            pts_total: 50,
            first: '2026-05-14T00:00:00Z',
            frequency: '10s',
            last: '2026-05-14T00:00:10Z',
            last_error: '',
            last_error_ts: '',
            max_collect_cost: '20ms',
            feed_total: 1000,
            p90_lat: '15ms',
            p90_pts: '30',
          },
        },
        goroutine_stats: {
          Items: {
            worker: {
              finished_goroutines: 1,
              running_goroutines: 2,
              total_cost_time: '3s',
              min_cost_time: '1s',
              max_cost_time: '5s',
              err_count: 0,
            },
          },
        },
        http_metrics: {},
        filter_stats: { rule_stats: {} },
        pl_stats: [],
        io_stats: {
          drop_pts: 0,
          chan_usage: {
            metric: [3, 2048],
          },
          M_fail_pts: 10,
          M_send_pts: 100,
        },
      },
    });

    expect(screen.getByText('Overview')).toBeInTheDocument();
    expect(screen.getByText('Basic information')).toBeInTheDocument();
    expect(screen.getByText('dk-host')).toBeInTheDocument();
    expect(screen.getByText('linux/amd64')).toBeInTheDocument();
    expect(screen.getByText('1.2.3')).toBeInTheDocument();
    expect(screen.getByText('10.0.0.1')).toBeInTheDocument();
    expect(screen.getByText('rt-1')).toBeInTheDocument();
    expect(screen.getByText('standalone')).toBeInTheDocument();
    expect(screen.getByText('No')).toBeInTheDocument();
    expect(screen.getByText('2.00KB')).toBeInTheDocument();
    expect(screen.getByText('1024.00B')).toBeInTheDocument();
    expect(screen.getByText('1.23s')).toBeInTheDocument();
    expect(screen.getByText('true')).toBeInTheDocument();
    expect(screen.getByText('Enabled collectors')).toBeInTheDocument();
    expect(screen.getByText('Collector runtime metrics')).toBeInTheDocument();
    expect(screen.getByText('1 crashes in total')).toBeInTheDocument();

    expect(screen.getByText('cpu')).toBeInTheDocument();
    expect(screen.getByText('2 instances')).toBeInTheDocument();
    expect(screen.getByText('1 crashed')).toBeInTheDocument();

    expect(screen.getByText(/"name":"cpu"/)).toBeInTheDocument();
    expect(screen.getByText(/"dataType":"M"/)).toBeInTheDocument();
    expect(screen.getByText(/"instanceCount":2/)).toBeInTheDocument();
    expect(screen.getByText(/"crashCount":1/)).toBeInTheDocument();
    expect(screen.getByText(/"feed_total":1000/)).toBeInTheDocument();
  });

  it('supports legacy enabled_inputs data and unknown categories', () => {
    renderRunInfo({
      datakit: {
        version: '2.0.0',
        run_in_container: true,
        status: 'running',
      },
      datakitStat: {
        hostname: 'legacy-host',
        os_arch: 'darwin/amd64',
        uptime: '5ms',
        enabled_inputs: [
          {
            input: 'custom',
            instances: 4,
            panic: 2,
          },
        ],
        enabled_input_list: {},
        inputs_status: {
          custom: {
            avg_collect_cost: '1ms',
            avg_size: 1,
            category: 'mystery',
            pts_total: 1,
            first: '2026-05-14T00:00:00Z',
            frequency: '',
            last: '0001-01-01T00:00:00Z',
            last_error: 'boom',
            last_error_ts: 'now',
            max_collect_cost: '2ms',
            feed_total: 2,
            p90_lat: '1ms',
            p90_pts: '1',
          },
        },
      },
    });

    expect(screen.getByText('legacy-host')).toBeInTheDocument();
    expect(screen.getByText('Yes')).toBeInTheDocument();
    expect(screen.getByText(/"name":"custom"/)).toBeInTheDocument();
    expect(screen.getByText(/"instanceCount":4/)).toBeInTheDocument();
    expect(screen.getByText(/"crashCount":2/)).toBeInTheDocument();
    expect(screen.getByText(/"dataType":"-"/)).toBeInTheDocument();
    expect(screen.getByText(/"frequency":"-"/)).toBeInTheDocument();
    expect(screen.getByText(/"last":"-"/)).toBeInTheDocument();
    expect(screen.getByText('0 inputs')).toBeInTheDocument();
    expect(screen.getByText('No enabled collectors')).toBeInTheDocument();
  });

  it('hides future LastFeed and NaN P90 values', () => {
    renderRunInfo({
      datakit: {
        version: '2.0.0',
        run_in_container: false,
        status: 'running',
      },
      datakitStat: {
        hostname: 'future-host',
        os_arch: 'linux/amd64',
        uptime: '5ms',
        enabled_input_list: {},
        inputs_status: {
          cpu: {
            avg_collect_cost: 0,
            avg_size: 1,
            category: 'metric',
            pts_total: 1,
            first: '2999-01-01T00:00:00Z',
            frequency: '',
            last: '2999-01-01T00:00:00Z',
            last_error: '',
            last_error_ts: '',
            max_collect_cost: 0,
            feed_total: 2,
            p90_lat: 'NaN',
            p90_pts: Number.NaN,
          },
        },
      },
    });

    expect(screen.getByText(/"name":"cpu"/)).toBeInTheDocument();
    expect(screen.getByText(/"last":"-"/)).toBeInTheDocument();
    expect(screen.getByText(/"p90_lat":"-"/)).toBeInTheDocument();
    expect(screen.getByText(/"p90_pts":"-"/)).toBeInTheDocument();
  });
});
