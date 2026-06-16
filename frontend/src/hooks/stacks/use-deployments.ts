import {
  deployementAPIList,
  getDeployementAPIListQueryKey,
  type DeployementAPIList200,
  type DeployementAPIListParams,
  type Deployment,
  type Error,
} from '@/api/api';
import {
  type InfiniteData,
  type QueryClient,
  type UseInfiniteQueryOptions,
} from '@tanstack/react-query';
import type { AxiosError, AxiosResponse } from 'axios';

const initialParams = { limit: 10, offset: '' } as DeployementAPIListParams;

export function getDeploymentsQueryOptions(): UseInfiniteQueryOptions<
  AxiosResponse<DeployementAPIList200>,
  AxiosError<Error>,
  Deployment[],
  readonly ['/api/deployment', ...DeployementAPIListParams[]],
  DeployementAPIListParams
> {
  return {
    queryKey: getDeployementAPIListQueryKey(initialParams),
    queryFn: ({ pageParam = initialParams }: { pageParam: DeployementAPIListParams }) =>
      deployementAPIList(pageParam),
    initialPageParam: initialParams,
    select: (
      data: InfiniteData<AxiosResponse<DeployementAPIList200>, DeployementAPIListParams>,
    ): Deployment[] => {
      return data.pages.flatMap((page) => page.data.items ?? []);
    },
    getNextPageParam: (lastPage: AxiosResponse<DeployementAPIList200>) => {
      if (lastPage.data.pageInfo.endCursor === '') {
        return undefined;
      }
      return { limit: initialParams.limit, offset: lastPage.data.pageInfo.endCursor };
    },
    gcTime: 10 * 60 * 1000,
  };
}

export function refetchDeployments(queryClient: QueryClient) {
  queryClient.refetchQueries(getDeploymentsQueryOptions());
}
