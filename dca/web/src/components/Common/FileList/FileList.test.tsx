import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';

jest.mock('antd', () => ({
  Tooltip: ({ children, title }: any) => <div data-title={title}>{children}</div>,
  Typography: {
    Text: ({ children, onClick, style }: any) => <span onClick={onClick} style={style}>{children}</span>,
  },
}));

jest.mock('@ant-design/icons', () => ({
  CaretDownOutlined: () => <span>down</span>,
  CaretRightOutlined: () => <span>right</span>,
  QuestionCircleOutlined: () => <span>help</span>,
}));

const FileList = require('./FileList').default;

describe('FileList', () => {
  it('renders the title and expands folders to show child items', () => {
    render(
      <FileList
        title="config files"
        list={[
          {
            name: 'inputs',
            expand: false,
            children: [{ name: 'cpu.conf', content: 'cpu' }],
          },
        ]}
        selected={null}
        setSelected={jest.fn()}
      />
    );

    expect(screen.getByText('config files')).toBeInTheDocument();
    expect(screen.getByText('- cpu.conf').closest('.content')).toHaveClass('hidden');

    fireEvent.click(screen.getByText('inputs'));
    expect(screen.getByText('- cpu.conf').closest('.content')).not.toHaveClass('hidden');
  });

  it('waits for onBeforeSelected and then calls setSelected/onAfterSelected', async () => {
    const setSelected = jest.fn();
    const onBeforeSelected = jest.fn().mockResolvedValue(true);
    const onAfterSelected = jest.fn().mockResolvedValue(undefined);

    render(
      <FileList
        title="pipelines"
        list={[
          {
            name: 'default',
            expand: true,
            children: [{ name: 'default.p', content: 'content' }],
          },
        ]}
        selected={null}
        setSelected={setSelected}
        onBeforeSelected={onBeforeSelected}
        onAfterSelected={onAfterSelected}
      />
    );

    fireEvent.click(screen.getByText('- default.p'));

    await waitFor(() => {
      expect(onBeforeSelected).toHaveBeenCalled();
      expect(setSelected).toHaveBeenCalledWith(
        expect.objectContaining({ name: 'default.p' })
      );
      expect(onAfterSelected).toHaveBeenCalledWith(
        expect.objectContaining({ name: 'default.p' })
      );
    });
  });

  it('does not select an item when onBeforeSelected rejects selection', async () => {
    const setSelected = jest.fn();
    const onBeforeSelected = jest.fn().mockResolvedValue(false);

    render(
      <FileList
        title="pipelines"
        list={[
          {
            name: 'default',
            expand: true,
            children: [{ name: 'default.p', content: 'content' }],
          },
        ]}
        selected={null}
        setSelected={setSelected}
        onBeforeSelected={onBeforeSelected}
      />
    );

    fireEvent.click(screen.getByText('- default.p'));

    await waitFor(() => {
      expect(onBeforeSelected).toHaveBeenCalled();
    });
    expect(setSelected).not.toHaveBeenCalled();
  });

  it('highlights the selected item and respects hidden child icons', () => {
    render(
      <FileList
        title="config files"
        list={[
          {
            name: 'inputs',
            expand: true,
            hiddenChildListIcon: true,
            children: [{ name: 'cpu.conf', content: 'cpu', tooltip: 'cpu file' }],
          },
        ]}
        selected={{ name: 'cpu.conf', content: 'cpu' }}
        setSelected={jest.fn()}
      />
    );

    const item = screen.getByText('cpu.conf').closest('.content-item');
    expect(item).toHaveClass('selected');
    expect(screen.queryByText('- cpu.conf')).not.toBeInTheDocument();
  });
});
