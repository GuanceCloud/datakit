import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';

const mockAlertError = jest.fn();
const mockTestPipeline = jest.fn();

jest.mock('src/api/api', () => ({
  testPipeline: (...args: unknown[]) => mockTestPipeline(...args),
}));

jest.mock('src/helper/helper', () => ({
  __esModule: true,
  alertError: (...args: unknown[]) => mockAlertError(...args),
}));

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

jest.mock('react-json-view-lite', () => ({
  JsonView: ({ data }: { data: unknown }) => <pre>{JSON.stringify(data)}</pre>,
  allExpanded: jest.fn(),
  defaultStyles: {},
}));

jest.mock('antd', () => ({
  Button: ({ children, onClick }: any) => <button onClick={onClick}>{children}</button>,
}));

jest.mock('antd/lib/input/TextArea', () => ({
  __esModule: true,
  default: ({ value, onChange }: any) => (
    <textarea value={value} onChange={onChange} />
  ),
}));

jest.mock('@ant-design/icons', () => ({
  CloseOutlined: () => <span>close</span>,
  PlusOutlined: () => <span>plus</span>,
}));

const PipelineTest = require('./PipelineTest').default;

function renderPipelineTest() {
  return render(
    <PipelineTest
      datakit={{ id: 'dk-1' }}
      fileName="test.p"
      category="logging"
      pipeline='grok(_, "%{GREEDYDATA:msg}")'
    />
  );
}

describe('PipelineTest', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockTestPipeline.mockResolvedValue([null, { logging: [{ msg: 'ok' }] }]);
  });

  it('runs pipeline test with current sample text and renders parsed result', async () => {
    renderPipelineTest();

    fireEvent.change(screen.getByRole('textbox'), {
      target: { value: 'hello world' },
    });
    fireEvent.click(screen.getByRole('button', { name: /pipeline.start_test/i }));

    await waitFor(() => {
      expect(mockTestPipeline).toHaveBeenCalledWith(
        { id: 'dk-1' },
        expect.objectContaining({
          category: 'logging',
          script_name: 'test.p',
          data: ['hello world'],
          pipeline: {
            logging: {
              'test.p': 'grok(_, "%{GREEDYDATA:msg}")',
            },
          },
        })
      );
    });
    await waitFor(() => {
      expect(screen.getByText((content) => content.includes('"msg":"ok"'))).toBeInTheDocument();
    });
  });

  it('adds and removes samples while enforcing the three-sample limit', () => {
    renderPipelineTest();

    fireEvent.click(screen.getByText('pipeline.add_sample'));
    fireEvent.click(screen.getByText('pipeline.add_sample'));
    fireEvent.click(screen.getByText('pipeline.add_sample'));

    expect(screen.getAllByRole('textbox')).toHaveLength(3);
    expect(screen.getByText('sample1')).toBeInTheDocument();
    expect(screen.getByText('sample2')).toBeInTheDocument();
    expect(screen.getByText('sample3')).toBeInTheDocument();

    fireEvent.click(screen.getAllByText('close')[1]);
    expect(screen.getAllByRole('textbox')).toHaveLength(2);
  });

  it('switches samples and reruns the test for the selected tab', async () => {
    renderPipelineTest();

    fireEvent.click(screen.getByText('pipeline.add_sample'));
    fireEvent.change(screen.getAllByRole('textbox')[0], {
      target: { value: 'first sample' },
    });
    fireEvent.change(screen.getAllByRole('textbox')[1], {
      target: { value: 'second sample' },
    });

    mockTestPipeline.mockClear();
    fireEvent.click(screen.getByText('sample2'));

    await waitFor(() => {
      expect(mockTestPipeline).toHaveBeenCalledWith(
        { id: 'dk-1' },
        expect.objectContaining({
          data: ['second sample'],
        })
      );
    });
  });

  it('reports pipeline execution errors', async () => {
    mockTestPipeline.mockResolvedValue(['pipeline failed', null]);

    renderPipelineTest();

    fireEvent.change(screen.getByRole('textbox'), {
      target: { value: 'broken sample' },
    });
    fireEvent.click(screen.getByRole('button', { name: /pipeline.start_test/i }));

    await waitFor(() => {
      expect(mockAlertError).toHaveBeenCalledWith('pipeline failed');
    });
  });
});
