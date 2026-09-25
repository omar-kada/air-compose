const mockToastPromise = vi.hoisted(() => vi.fn());
const mockFn = vi.hoisted(() =>
  vi.fn((opts: { mutation?: Record<string, unknown> } = {}) => ({
    mutationFn: vi.fn().mockResolvedValue({ data: { id: '1' } }),
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

vi.mock('@/lib', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/lib')>();
  return { ...original, useDeploymentNavigate: vi.fn(() => vi.fn()) };
});

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return {
    ...original,
    getDeployementAPISyncMutationOptions: mockFn,
  };
});

import { getSyncOptions, useSync } from './use-sync';
import { renderHookWithQuery } from '@/tests/test-utils';
import type { AnyFunction } from '@/tests/test-utils';

describe('useSync', () => {
  it('returns sync function', () => {
    const { result } = renderHookWithQuery(() => useSync());
    expect(result.current.sync).toBeInstanceOf(Function);
  });

  it('sync calls toast.promise', () => {
    const { result } = renderHookWithQuery(() => useSync());
    (result.current.sync as unknown as AnyFunction)();
    expect(mockToastPromise).toHaveBeenCalled();
  });

  describe('getSyncOptions', () => {
    it('onSuccess refetches deployments when id is valid', () => {
      const onSuccess = (getSyncOptions() as any).onSuccess;
      const client = { refetchQueries: vi.fn() };
      const context = { client };
      onSuccess({ data: { id: '123' } }, undefined, undefined, context);
      expect(client.refetchQueries).toHaveBeenCalled();
    });

    it('onSuccess does not refetch when id is "0"', () => {
      const onSuccess = (getSyncOptions() as any).onSuccess;
      const client = { refetchQueries: vi.fn() };
      const context = { client };
      onSuccess({ data: { id: '0' } }, undefined, undefined, context);
      expect(client.refetchQueries).not.toHaveBeenCalled();
    });

    it('onSuccess does not refetch when id is undefined', () => {
      const onSuccess = (getSyncOptions() as any).onSuccess;
      const client = { refetchQueries: vi.fn() };
      const context = { client };
      onSuccess({ data: { id: undefined } }, undefined, undefined, context);
      expect(client.refetchQueries).not.toHaveBeenCalled();
    });
  });
});
