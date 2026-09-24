const mockToastPromise = vi.hoisted(() => vi.fn());
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
    getAuthAPILoginMutationOptions: mockFn,
    getUserAPIGetQueryOptions: vi.fn(() => ({ queryKey: ['user'] })),
  };
});

import { useLogin } from './use-login';
import { renderHookWithQuery } from '@/tests/test-utils';
import type { AnyFunction } from '@/tests/test-utils';

describe('useLogin', () => {
  it('returns login function', () => {
    const { result } = renderHookWithQuery(() => useLogin());
    expect(result.current.login).toBeInstanceOf(Function);
  });

  it('login calls toast.promise', () => {
    const { result } = renderHookWithQuery(() => useLogin());
    (result.current.login as unknown as AnyFunction)({
      username: 'test',
      password: 'test',
    });
    expect(mockToastPromise).toHaveBeenCalled();
  });
});
