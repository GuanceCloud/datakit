const mockMetrics = {
  getCLS: jest.fn(),
  getFID: jest.fn(),
  getFCP: jest.fn(),
  getLCP: jest.fn(),
  getTTFB: jest.fn(),
};

jest.mock('web-vitals', () => mockMetrics);

import reportWebVitals from './reportWebVitals';
import fs from 'fs';
import path from 'path';

describe('reportWebVitals', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('does nothing when no callback is provided', async () => {
    reportWebVitals();
    await Promise.resolve();

    expect(mockMetrics.getCLS).not.toHaveBeenCalled();
  });

  it('keeps the web-vitals dynamic import wiring in source', () => {
    const source = fs.readFileSync(path.join(__dirname, 'reportWebVitals.ts'), 'utf8');

    expect(source).toContain("import('web-vitals')");
    expect(source).toContain('getCLS');
    expect(source).toContain('getFID');
    expect(source).toContain('getFCP');
    expect(source).toContain('getLCP');
    expect(source).toContain('getTTFB');
  });
});
