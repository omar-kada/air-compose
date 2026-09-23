const mockFn = vi.hoisted(() =>
  vi.fn((opts: any) => ({
    mutationFn: vi.fn().mockResolvedValue({ data: { success: true } }),
    ...(opts?.mutation ?? {}),
  })),
);

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return {
    ...original,
    getAuthAPIRefreshMutationOptions: mockFn,
    getUserAPIGetQueryOptions: vi.fn(() => ({ queryKey: ['user'] })),
  };
});

import { useRefresh } from './use-refresh';
import { renderHookWithQuery } from '@/tests/test-utils';

describe('useRefresh', () => {
  it('returns refresh function', () => {
    const { result } = renderHookWithQuery(() => useRefresh());
    expect(result.current.refresh).toBeInstanceOf(Function);
  });

  it('refresh calls mutationFn', async () => {
    const { result } = renderHookWithQuery(() => useRefresh());
    const mutationFn = mockFn.mock.results[0].value.mutationFn;
    await (result.current.refresh as any)();
    expect(mutationFn).toHaveBeenCalled();
  });
});
