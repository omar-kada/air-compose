const mockToastPromise = vi.hoisted(() =>
  vi.fn(() => ({ unwrap: vi.fn().mockResolvedValue(true) })),
);
const mockFn = vi.hoisted(() =>
  vi.fn((opts: any) => ({
    mutationFn: vi.fn().mockResolvedValue({ data: { success: true } }),
    ...(opts?.mutation ?? {}),
  })),
);

vi.mock('sonner', async () => {
  const { createSonnerMock } = await import('@/tests/mock-factories');
  const m = createSonnerMock();
  (m.toast as any).promise = mockToastPromise;
  return m;
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
    getAuthAPIRegisteredQueryOptions: vi.fn(() => ({ queryKey: ['registered'] })),
  };
});

import { useDeleteAccount } from './use-delete-account';
import { renderHookWithQuery } from '@/tests/test-utils';

describe('useDeleteAccount', () => {
  it('returns deleteAccount function', () => {
    const { result } = renderHookWithQuery(() => useDeleteAccount());
    expect(result.current.deleteAccount).toBeInstanceOf(Function);
  });

  it('deleteAccount calls toast.promise with unwrap', async () => {
    const { result } = renderHookWithQuery(() => useDeleteAccount());
    await (result.current.deleteAccount as any)();
    expect(mockToastPromise).toHaveBeenCalled();
  });
});
