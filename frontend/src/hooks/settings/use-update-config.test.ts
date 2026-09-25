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
    getConfigAPISetMutationOptions: mockFn,
    getConfigAPIGetQueryKey: vi.fn(() => ['config-key']),
  };
});

import { getUpdateConfigOptions, useUpdateConfig } from './use-update-config';
import { renderHookWithQuery } from '@/tests/test-utils';
import type { AnyFunction } from '@/tests/test-utils';

describe('useUpdateConfig', () => {
  it('returns updateConfig function', () => {
    const { result } = renderHookWithQuery(() => useUpdateConfig());
    expect(result.current.updateConfig).toBeInstanceOf(Function);
  });

  it('updateConfig calls toast.promise', () => {
    const { result } = renderHookWithQuery(() => useUpdateConfig());
    (result.current.updateConfig as unknown as AnyFunction)({ repo: 'test' });
    expect(mockToastPromise).toHaveBeenCalled();
  });

  it('getUpdateConfigOptions onSuccess calls setQueryData with config key', () => {
    const onSuccess = (getUpdateConfigOptions() as any).onSuccess;
    const client = { setQueryData: vi.fn(), refetchQueries: vi.fn() };
    const context = { client };
    onSuccess({ data: { repo: 'test' } }, undefined, undefined, context);
    expect(client.setQueryData).toHaveBeenCalledWith(['config-key'], { data: { repo: 'test' } });
  });
});
