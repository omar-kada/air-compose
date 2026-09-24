const mockFn = vi.hoisted(() => vi.fn());

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return { ...original, getDiffAPIGetQueryOptions: mockFn };
});

import { getDiffQueryOptions } from './use-diff';

describe('getDiffQueryOptions', () => {
  beforeEach(() => mockFn.mockClear());

  it('passes select and gcTime to API', () => {
    getDiffQueryOptions();
    expect(mockFn).toHaveBeenCalledWith({
      query: {
        select: expect.any(Function),
        gcTime: 600000,
      },
    });
  });

  it('select extracts data.data from response', () => {
    getDiffQueryOptions();
    const select = mockFn.mock.calls[0][0].query.select;
    expect(select({ data: { diff: 'abc' } })).toEqual({ diff: 'abc' });
  });
});
