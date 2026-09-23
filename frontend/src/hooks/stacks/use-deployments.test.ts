const mockList = vi.hoisted(() => vi.fn(() => ({ data: { items: [], pageInfo: { endCursor: '' } } })));
const mockKey = vi.hoisted(() => vi.fn(() => ['deployments-key']));

vi.mock('@/api/api', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/api/api')>();
  return { ...original, deployementAPIList: mockList, getDeployementAPIListQueryKey: mockKey };
});

import { getDeploymentsQueryOptions, refetchDeployments } from './use-deployments';

describe('getDeploymentsQueryOptions', () => {
  it('returns infinite query options with correct structure', () => {
    const options = getDeploymentsQueryOptions();
    expect(options.queryKey).toEqual(['deployments-key']);
    expect(options.queryFn).toBeInstanceOf(Function);
    expect(options.select).toBeInstanceOf(Function);
    expect(options.getNextPageParam).toBeInstanceOf(Function);
    expect(options.initialPageParam).toEqual({ limit: 10, offset: '' });
    expect(options.gcTime).toBe(600000);
  });

  it('queryFn calls deployementAPIList with pageParam', async () => {
    const options = getDeploymentsQueryOptions();
    const result = await (options.queryFn! as any)({ pageParam: { limit: 10, offset: 'abc' } });
    expect(mockList).toHaveBeenCalledWith({ limit: 10, offset: 'abc' });
    expect(result).toEqual({ data: { items: [], pageInfo: { endCursor: '' } } });
  });

  it('select flattens pages into Deployment[]', () => {
    const options = getDeploymentsQueryOptions();
    const data = {
      pageParams: [],
      pages: [
        { data: { items: [{ id: 1 }, { id: 2 }] } },
        { data: { items: [{ id: 3 }] } },
      ],
    };
    expect((options.select! as any)(data)).toEqual([{ id: 1 }, { id: 2 }, { id: 3 }]);
  });

  it('getNextPageParam returns undefined when endCursor is empty', () => {
    const options = getDeploymentsQueryOptions();
    const lastPage = { data: { items: [], pageInfo: { endCursor: '' } } };
    expect((options.getNextPageParam! as any)(lastPage)).toBeUndefined();
  });

  it('getNextPageParam returns offset object when endCursor is present', () => {
    const options = getDeploymentsQueryOptions();
    const lastPage = { data: { items: [], pageInfo: { endCursor: 'xyz' } } };
    expect((options.getNextPageParam! as any)(lastPage)).toEqual({ limit: 10, offset: 'xyz' });
  });

  it('refetchDeployments calls refetchQueries with query options', () => {
    const mockClient = { refetchQueries: vi.fn() } as any;
    refetchDeployments(mockClient);
    expect(mockClient.refetchQueries).toHaveBeenCalled();
  });
});
