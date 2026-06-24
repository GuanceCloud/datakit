import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import DcaHelp from './DcaHelp';

const mockSuccess = jest.fn();

jest.mock('antd', () => ({
  Button: ({ children, onClick }: any) => <button onClick={onClick}>{children}</button>,
  Spin: () => <div>loading</div>,
  message: {
    success: (...args: any[]) => mockSuccess(...args),
  },
}));

jest.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      if (key === 'no_data') return '```toml\n[a]\n```';
      if (key === 'copy_success') return 'copied';
      if (key === 'copy_code') return 'copy';
      return key;
    },
  }),
}));

jest.mock('react-copy-to-clipboard', () => ({
  CopyToClipboard: ({ children, onCopy }: any) => <div onClick={onCopy}>{children}</div>,
}));

jest.mock('react-markdown', () => ({
  __esModule: true,
  default: ({ components, children }: any) => {
    const code = components.code;
    return <div>{code({ inline: false, className: 'language-toml', children: ['[a]\n'] })}</div>;
  },
}));

jest.mock('react-syntax-highlighter', () => ({
  Prism: ({ children }: any) => <pre>{children}</pre>,
}));

jest.mock('react-syntax-highlighter/dist/esm/styles/prism', () => ({
  xonokai: {},
}));

jest.mock('remark-gfm', () => jest.fn());

describe('DcaHelp', () => {
  it('renders markdown help content and supports copy feedback', async () => {
    render(<DcaHelp />);

    await waitFor(() => {
      expect(screen.getByText('copy')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('copy'));
    expect(mockSuccess).toHaveBeenCalledWith('copied');
  });
});
