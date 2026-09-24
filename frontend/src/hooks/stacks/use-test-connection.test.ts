const mockToastPromise = vi.hoisted(() => vi.fn(() => Promise.resolve(true)));
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
    getSettingsAPITestGitConnectionMutationOptions: mockFn,
  };
});

import { useTestConnection } from './use-test-connection';
import { renderHookWithQuery } from '@/tests/test-utils';
import type { AnyFunction } from '@/tests/test-utils';

describe('useTestConnection', () => {
  it('returns testConnection function', () => {
    const { result } = renderHookWithQuery(() => useTestConnection());
    expect(result.current.testConnection).toBeInstanceOf(Function);
  });

  it('testConnection calls toast.promise', () => {
    const { result } = renderHookWithQuery(() => useTestConnection());
    (result.current.testConnection as unknown as AnyFunction)({
      token: 'abc',
      url: 'https://git.example.com',
    });
    expect(mockToastPromise).toHaveBeenCalled();
  });
});
