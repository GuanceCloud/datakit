import React from 'react';
import { act, render } from '@testing-library/react';

const mockCodeMirror = jest.fn();

jest.mock('react-codemirror2', () => ({
  Controlled: (props: any) => {
    mockCodeMirror(props);
    return <div data-testid="codemirror" />;
  },
}));

jest.mock('codemirror/lib/codemirror.css', () => ({}));
jest.mock('codemirror/theme/base16-light.css', () => ({}));
jest.mock('codemirror/mode/toml/toml', () => ({}));
jest.mock('codemirror/mode/javascript/javascript', () => ({}));
jest.mock('codemirror/addon/selection/active-line', () => ({}));
jest.mock('codemirror/addon/scroll/scrollpastend', () => ({}));

const DCAEditor = require('./DCAEditor').default;

describe('DCAEditor', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    mockCodeMirror.mockClear();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  it('merges default options with editable mode and forwards value changes', () => {
    const setValue = jest.fn();
    render(
      <DCAEditor
        value="hostname = 'demo'"
        setValue={setValue}
        editorOptions={{ readOnly: false, mode: 'javascript' }}
      />
    );

    const props = mockCodeMirror.mock.calls.at(-1)?.[0];
    expect(props.value).toBe("hostname = 'demo'");
    expect(props.options).toEqual(expect.objectContaining({
      mode: 'javascript',
      lineNumbers: true,
      lineWrapping: true,
      dragDrop: false,
      readOnly: false,
      cursorBlinkRate: 500,
    }));

    props.onBeforeChange({}, {}, 'new value');
    expect(setValue).toHaveBeenCalledWith('new value');
  });

  it('configures the editor on mount, adds overlay, and forwards editorDidMount', () => {
    const focus = jest.fn();
    const setCursor = jest.fn();
    const addOverlay = jest.fn();
    const editorDidMount = jest.fn();

    render(
      <DCAEditor
        value="source"
        editorDidMount={editorDidMount}
      />
    );

    const props = mockCodeMirror.mock.calls.at(-1)?.[0];
    const editor = {
      focus,
      setCursor,
      addOverlay,
    };
    const callback = jest.fn();

    act(() => {
      props.editorDidMount(editor, 'source', callback);
      jest.runAllTimers();
    });

    expect(setCursor).toHaveBeenCalledWith(0);
    expect(addOverlay).toHaveBeenCalledWith(expect.objectContaining({
      token: expect.any(Function),
    }));
    expect(focus).toHaveBeenCalled();
    expect(editorDidMount).toHaveBeenCalledWith(editor, 'source', callback);

    const stream = {
      pos: 0,
      string: '{{ value }} tail',
      skipToEnd: jest.fn(),
    };
    const tokenFn = addOverlay.mock.calls[0][0].token;
    expect(tokenFn(stream)).toBe('notelink');
    expect(stream.pos).toBe('{{ value }}'.length);
  });

  it('refocuses on empty change and adjusts cursor after paste selection', () => {
    render(<DCAEditor value="demo" />);

    const props = mockCodeMirror.mock.calls.at(-1)?.[0];
    const focus = jest.fn();
    const setCursor = jest.fn();
    const getSelection = jest.fn(() => 'pasted-value');
    const listSelections = jest.fn(() => [{
      anchor: { line: 1, ch: 3 },
      head: { line: 2, ch: 5 },
    }]);

    const editor = {
      focus,
      setCursor,
      getSelection,
      listSelections,
    };

    props.onChange(editor, {}, '');
    expect(focus).toHaveBeenCalled();

    const event = {
      clipboardData: {
        items: [{
          getAsString: (cb: (value: string) => void) => cb('pasted-value'),
        }],
      },
    };

    props.onPaste(editor, event);
    expect(setCursor).toHaveBeenCalledWith({ line: 2, ch: 5 });
  });

  it('ignores incomplete paste events and mismatched pasted content', () => {
    render(<DCAEditor value="demo" />);

    const props = mockCodeMirror.mock.calls.at(-1)?.[0];
    const setCursor = jest.fn();
    const editor = {
      getSelection: jest.fn(() => 'selection'),
      listSelections: jest.fn(() => [{ anchor: { line: 0, ch: 0 }, head: { line: 0, ch: 1 } }]),
      setCursor,
    };

    expect(() => props.onPaste(editor, {})).not.toThrow();

    const event = {
      clipboardData: {
        items: [{
          getAsString: (cb: (value: string) => void) => cb('different'),
        }],
      },
    };

    props.onPaste(editor, event);
    expect(setCursor).not.toHaveBeenCalled();
  });
});
