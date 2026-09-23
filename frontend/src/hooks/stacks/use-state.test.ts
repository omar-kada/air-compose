const mockGetKey = vi.hoisted(() => vi.fn(() => ['state-key']));
const mockGetOpts = vi.hoisted(() => vi.fn((opts) => opts));

vi.mock('lodash.debounce', () => ({ default: (fn: Function) => fn }));

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return { ...original, getStateAPIGetQueryKey: mockGetKey, getStateAPIGetQueryOptions: mockGetOpts };
});

import { getStateQueryOptions, refetchState } from './use-state';

describe('getStateQueryOptions', () => {
  it('passes select, refetchInterval, staleTime, gcTime to API', () => {
    getStateQueryOptions();
    expect(mockGetOpts).toHaveBeenCalledTimes(1);
    const queryOpts = (mockGetOpts as any).mock.calls[0][0];
    const query = queryOpts.query;
    expect(query.select).toBeInstanceOf(Function);
    expect(query.refetchInterval).toBeInstanceOf(Function);
    expect(query.staleTime).toBe(0);
    expect(query.gcTime).toBe(600000);
    expect(query.refetchIntervalInBackground).toBe(false);
  });

  it('select extracts data.data', () => {
    getStateQueryOptions();
    const select = (mockGetOpts as any).mock.calls[0][0].query.select;
    expect(select({ data: { status: 'running' } })).toEqual({ status: 'running' });
  });

  it('refetchInterval returns diffMs for future nextDeploy', () => {
    getStateQueryOptions();
    const refetchInterval = (mockGetOpts as any).mock.calls[0][0].query.refetchInterval;
    const future = new Date(Date.now() + 50_000).toISOString();
    const result = refetchInterval({ state: { data: { data: { nextDeploy: future } } } });
    expect(result).toBeGreaterThan(0);
    expect(result).toBeLessThanOrEqual(50_000);
  });

  it('refetchInterval returns 60000 for past nextDeploy', () => {
    getStateQueryOptions();
    const refetchInterval = (mockGetOpts as any).mock.calls[0][0].query.refetchInterval;
    const past = new Date(Date.now() - 10_000).toISOString();
    expect(refetchInterval({ state: { data: { data: { nextDeploy: past } } } })).toBe(60000);
  });

  it('refetchInterval returns 60000 when nextDeploy is absent', () => {
    getStateQueryOptions();
    const refetchInterval = (mockGetOpts as any).mock.calls[0][0].query.refetchInterval;
    expect(refetchInterval({ state: { data: { data: {} } } })).toBe(60000);
  });

  it('refetchState calls refetchQueries with getStateAPIGetQueryKey', () => {
    const mockClient = { refetchQueries: vi.fn() } as any;
    refetchState(mockClient);
    expect(mockGetKey).toHaveBeenCalled();
    expect(mockClient.refetchQueries).toHaveBeenCalled();
  });
});
