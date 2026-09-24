const mockFn = vi.hoisted(() => vi.fn());

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return { ...original, getSettingsAPIGetQueryOptions: mockFn };
});

import { getSettingsQueryOptions } from './use-settings';

describe('getSettingsQueryOptions', () => {
  beforeEach(() => mockFn.mockClear());

  it('passes select and gcTime to API', () => {
    getSettingsQueryOptions();
    expect(mockFn).toHaveBeenCalledWith({
      query: {
        select: expect.any(Function),
        gcTime: 600000,
      },
    });
  });

  it('select extracts data.data from response', () => {
    getSettingsQueryOptions();
    const select = mockFn.mock.calls[0][0].query.select;
    expect(select({ data: { repo: 'test' } })).toEqual({ repo: 'test' });
  });
});
