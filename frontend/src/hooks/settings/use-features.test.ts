const mockFn = vi.hoisted(() => vi.fn());

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return { ...original, getFeaturesAPIGetQueryOptions: mockFn };
});

import { getFeaturesQueryOptions } from './use-features';

describe('getFeaturesQueryOptions', () => {
  beforeEach(() => mockFn.mockClear());

  it('passes select, staleTime=Infinity, gcTime to API', () => {
    getFeaturesQueryOptions();
    expect(mockFn).toHaveBeenCalledWith({
      query: {
        select: expect.any(Function),
        staleTime: Infinity,
        gcTime: 600000,
      },
    });
  });

  it('select returns data.data or empty object', () => {
    getFeaturesQueryOptions();
    const select = mockFn.mock.calls[0][0].query.select;
    expect(select({ data: { feature: true } })).toEqual({ feature: true });
  });
});
