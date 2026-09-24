import { waitFor } from '@testing-library/react';
import { useFilteredQuery } from './use-filtered-error';

const mockIsInvalidToken = vi.hoisted(() => vi.fn());

vi.mock('@/lib', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/lib')>();
  return {
    ...original,
    isInvalidToken: mockIsInvalidToken,
  };
});

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return original;
});

import { renderHookWithQuery } from '@/tests/test-utils';

describe('useFilteredQuery', () => {
  beforeEach(() => {
    mockIsInvalidToken.mockReset();
  });

  it('passes through non-invalid-token errors', async () => {
    mockIsInvalidToken.mockReturnValue(false);
    const { result } = renderHookWithQuery(() =>
      useFilteredQuery<{ message: string }>({
        queryKey: ['test'],
        queryFn: () => Promise.resolve({ message: 'hello' }),
      }),
    );
    await waitFor(() => {
      expect(result.current.isPending).toBe(false);
    });
    expect(result.current.data).toEqual({ message: 'hello' });
  });

  it('transforms 401 INVALID_TOKEN errors to null error + isPending', async () => {
    mockIsInvalidToken.mockReturnValue(true);
    const axiosError = new Error('401') as unknown;
    const { result } = renderHookWithQuery(() =>
      useFilteredQuery<string>({
        queryKey: ['test'],
        queryFn: () => Promise.reject(axiosError),
      }),
    );
    await waitFor(() => {
      expect(result.current.isPending).toBe(true);
    });
    expect(result.current.error).toBeNull();
  });
});
