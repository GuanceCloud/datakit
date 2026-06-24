import { fireEvent, render } from '@testing-library/react';
import ResizeBar from './ResizeBar';

describe('ResizeBar', () => {
  it('updates width within bounds and clears document handlers on mouseup', () => {
    const setWidth = jest.fn();
    const { container } = render(<ResizeBar oriWidth={200} setWidth={setWidth} className="bar" maxWidth={260} minWidth={150} />);
    const event = new MouseEvent('mousedown', { bubbles: true });
    Object.defineProperty(event, 'pageX', { value: 100 });

    (container.firstChild as Element).dispatchEvent(event);
    document.onmousemove?.({ pageX: 180 } as any);
    document.onmousemove?.({ pageX: 400 } as any);
    document.onmousemove?.({ pageX: -1000 } as any);

    expect(setWidth).toHaveBeenCalledWith(260);
    expect(setWidth).toHaveBeenCalledWith(150);

    document.onmouseup?.({} as any);
    expect(document.onmousemove).toBeNull();
    expect(document.onmouseup).toBeNull();
  });
});
