const fs = require('fs');
const path = require('path');

export {};

describe('DCA frontend debug logging', () => {
  it('does not leave console.log in core DCA components', () => {
    const files = [
      'DkInfo/DkInfo.tsx',
      'DatakitList/DkList.tsx',
      'DatakitList/AdditionalColumnOptions/useAdditionalColumnOptions.ts',
    ];

    for (const file of files) {
      const source = fs.readFileSync(path.join(__dirname, file), 'utf8');
      expect(source).not.toContain('console.log(');
    }
  });
});
