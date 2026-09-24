const mockToastPromise = vi.hoisted(() =>
  vi.fn(() => ({ unwrap: vi.fn().mockResolvedValue(true) })),
);
const mockFn = vi.hoisted(() =>
  vi.fn((opts: { mutation?: Record<string, unknown> } = {}) => ({
    mutationFn: vi.fn().mockResolvedValue({ data: { success: true } }),
    ...(opts?.mutation ?? {}),
  })),
);

vi.mock('sonner', async () => {
  const { createSonnerMock } = await import('@/tests/mock-factories');
  const mock = createSonnerMock();
  mock.toast.promise = mockToastPromise;
  return mock;
});

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return {
    ...original,
    getAuthAPILogoutMutationOptions: mockFn,
    getUserAPIGetQueryOptions: vi.fn(() => ({ queryKey: ['user'] })),
  };
});

import { useLogout } from './use-logout';
import { renderHookWithQuery } from '@/tests/test-utils';
import type { AnyFunction } from '@/tests/test-utils';

describe('useLogout', () => {
  it('returns logout function', () => {
    const { result } = renderHookWithQuery(() => useLogout());
    expect(result.current.logout).toBeInstanceOf(Function);
  });

  it('logout calls toast.promise and unwrap', async () => {
    const { result } = renderHookWithQuery(() => useLogout());
    await (result.current.logout as unknown as AnyFunction)();
    expect(mockToastPromise).toHaveBeenCalled();
  });
});
