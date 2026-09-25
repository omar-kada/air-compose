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
    getAuthAPIRegisterMutationOptions: mockFn,
    getAuthAPIRegisteredQueryOptions: vi.fn(() => ({
      queryKey: ['registered'],
    })),
    getUserAPIGetQueryOptions: vi.fn(() => ({ queryKey: ['user'] })),
  };
});

import { getRegisterOptions, useRegister } from './use-register';
import { renderHookWithQuery } from '@/tests/test-utils';
import type { AnyFunction } from '@/tests/test-utils';

describe('useRegister', () => {
  it('returns register function', () => {
    const { result } = renderHookWithQuery(() => useRegister());
    expect(result.current.register).toBeInstanceOf(Function);
  });

  it('register calls toast.promise', () => {
    const { result } = renderHookWithQuery(() => useRegister());
    (result.current.register as unknown as AnyFunction)({
      username: 'test',
      password: 'test',
    });
    expect(mockToastPromise).toHaveBeenCalled();
  });

  it('getRegisterOptions onSuccess refetches registered and user queries when registration succeeds', () => {
    const onSuccess = (getRegisterOptions() as any).onSuccess;
    const client = { refetchQueries: vi.fn() };
    const context = { client };
    onSuccess({ data: { success: true } }, undefined, undefined, context);
    expect(client.refetchQueries).toHaveBeenCalledWith({ queryKey: ['registered'] });
    expect(client.refetchQueries).toHaveBeenCalledWith({ queryKey: ['user'] });
  });

  it('getRegisterOptions onSuccess does not refetch when registration fails', () => {
    const onSuccess = (getRegisterOptions() as any).onSuccess;
    const client = { refetchQueries: vi.fn() };
    const context = { client };
    onSuccess({ data: { success: false } }, undefined, undefined, context);
    expect(client.refetchQueries).not.toHaveBeenCalled();
  });
});
