const mockToastPromise = vi.hoisted(() => vi.fn());
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
    getConfigAPISetMutationOptions: mockFn,
    getConfigAPIGetQueryKey: vi.fn(() => ['config-key']),
  };
});

import { useUpdateConfig } from './use-update-config';
import { renderHookWithQuery } from '@/tests/test-utils';

describe('useUpdateConfig', () => {
  it('returns updateConfig function', () => {
    const { result } = renderHookWithQuery(() => useUpdateConfig());
    expect(result.current.updateConfig).toBeInstanceOf(Function);
  });

  it('updateConfig calls toast.promise', () => {
    const { result } = renderHookWithQuery(() => useUpdateConfig());
    (result.current.updateConfig as any)({ repo: 'test' });
    expect(mockToastPromise).toHaveBeenCalled();
  });
});
