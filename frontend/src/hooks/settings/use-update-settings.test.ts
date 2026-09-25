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
    getSettingsAPISetMutationOptions: mockFn,
    getSettingsAPIGetQueryKey: vi.fn(() => ['settings-key']),
  };
});

import { getUpdateSettingsOptions, useUpdateSettings } from './use-update-settings';
import { renderHookWithQuery } from '@/tests/test-utils';
import type { AnyFunction } from '@/tests/test-utils';

describe('useUpdateSettings', () => {
  it('returns updateSettings function', () => {
    const { result } = renderHookWithQuery(() => useUpdateSettings());
    expect(result.current.updateSettings).toBeInstanceOf(Function);
  });

  it('updateSettings calls toast.promise', () => {
    const { result } = renderHookWithQuery(() => useUpdateSettings());
    (result.current.updateSettings as unknown as AnyFunction)({ theme: 'dark' });
    expect(mockToastPromise).toHaveBeenCalled();
  });

  it('getUpdateSettingsOptions onSuccess calls setQueryData and refetchQueries', () => {
    const onSuccess = (getUpdateSettingsOptions() as any).onSuccess;
    const client = { setQueryData: vi.fn(), refetchQueries: vi.fn() };
    const context = { client };
    onSuccess({ data: { theme: 'dark' } }, undefined, undefined, context);
    expect(client.setQueryData).toHaveBeenCalledWith(['settings-key'], { data: { theme: 'dark' } });
    expect(client.refetchQueries).toHaveBeenCalled();
  });
});
