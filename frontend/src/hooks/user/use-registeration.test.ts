const mockFn = vi.hoisted(() =>
  vi.fn((opts: any) => ({
    queryKey: ['registeration'],
    queryFn: vi.fn().mockResolvedValue({ data: { id: 1 } }),
    ...(opts?.query ?? {}),
  })),
);

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return { ...original, getAuthAPIRegisteredQueryOptions: mockFn };
});

import { useRegisteration } from './use-registeration';
import { renderHookWithQuery } from '@/tests/test-utils';

describe('useRegisteration', () => {
  it('passes select that returns data.data', () => {
    renderHookWithQuery(() => useRegisteration());
    const select = mockFn.mock.calls[0][0].query.select;
    expect(select({ data: { id: 1 } })).toEqual({ id: 1 });
  });

  it('select returns undefined when data is absent', () => {
    renderHookWithQuery(() => useRegisteration());
    const select = mockFn.mock.calls[0][0].query.select;
    expect(select({ data: null })).toBeNull();
  });
});
