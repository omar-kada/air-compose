const mockFn = vi.hoisted(() => vi.fn((opts) => opts));

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return { ...original, getDeployementAPIReadQueryOptions: mockFn };
});

import { getDeploymentOptions } from './use-deployment';
import { DeploymentStatus } from '@/api/api';

type QueryOpts = {
  query: {
    select: (data: unknown) => unknown;
    staleTime: (state: unknown) => number;
  };
};

describe('getDeploymentOptions', () => {
  it('passes id and query options to getDeployementAPIReadQueryOptions', () => {
    getDeploymentOptions('test-id');
    expect(mockFn).toHaveBeenCalledWith('test-id', {
      query: {
        select: expect.any(Function),
        staleTime: expect.any(Function),
        gcTime: 600000,
      },
    });
  });

  it('select extracts data.data', () => {
    getDeploymentOptions('id');
    const select = (mockFn.mock.calls[0] as unknown as [unknown, QueryOpts])[1].query.select;
    expect(select({ data: { id: 1 } })).toEqual({ id: 1 });
    expect(select({ data: null })).toBeNull();
  });

  it('staleTime returns 500 for running status', () => {
    getDeploymentOptions('id');
    const staleTime = (mockFn.mock.calls[0] as unknown as [unknown, QueryOpts])[1].query.staleTime;
    expect(
      staleTime({
        state: { data: { data: { status: DeploymentStatus.running } } },
      }),
    ).toBe(500);
  });

  it('staleTime returns Infinity for error status', () => {
    getDeploymentOptions('id');
    const staleTime = (mockFn.mock.calls[0] as unknown as [unknown, QueryOpts])[1].query.staleTime;
    expect(
      staleTime({
        state: { data: { data: { status: DeploymentStatus.error } } },
      }),
    ).toBe(Infinity);
  });

  it('staleTime returns Infinity for success status', () => {
    getDeploymentOptions('id');
    const staleTime = (mockFn.mock.calls[0] as unknown as [unknown, QueryOpts])[1].query.staleTime;
    expect(
      staleTime({
        state: { data: { data: { status: DeploymentStatus.success } } },
      }),
    ).toBe(Infinity);
  });

  it('staleTime returns 10000 for default status', () => {
    getDeploymentOptions('id');
    const staleTime = (mockFn.mock.calls[0] as unknown as [unknown, QueryOpts])[1].query.staleTime;
    expect(
      staleTime({
        state: { data: { data: { status: DeploymentStatus.planned } } },
      }),
    ).toBe(10000);
  });
});
