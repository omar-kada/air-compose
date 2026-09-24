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
    getUserAPIDeleteMutationOptions: mockFn,
    getUserAPIGetQueryOptions: vi.fn(() => ({ queryKey: ['user'] })),
    getAuthAPIRegisteredQueryOptions: vi.fn(() => ({
      queryKey: ['registered'],
    })),
  };
});

import { useDeleteAccount } from './use-delete-account';
import { renderHookWithQuery } from '@/tests/test-utils';
import type { AnyFunction } from '@/tests/test-utils';

describe('useDeleteAccount', () => {
  it('returns deleteAccount function', () => {
    const { result } = renderHookWithQuery(() => useDeleteAccount());
    expect(result.current.deleteAccount).toBeInstanceOf(Function);
  });

  it('deleteAccount calls toast.promise with unwrap', async () => {
    const { result } = renderHookWithQuery(() => useDeleteAccount());
    await (result.current.deleteAccount as unknown as AnyFunction)();
    expect(mockToastPromise).toHaveBeenCalled();
  });
});
