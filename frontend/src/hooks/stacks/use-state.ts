import { getStateAPIGetQueryOptions, type Error as ApiError, type State } from '@/api/api';
import type { UseQueryOptions } from '@tanstack/react-query';
import type { AxiosError, AxiosResponse } from 'axios';

export const getStateQueryOptions = (
  queryOptions?: Partial<
    UseQueryOptions<AxiosResponse<State, unknown>, AxiosError<ApiError>, State>
  >,
) => {
  return getStateAPIGetQueryOptions({
    query: {
      select: (data) => data?.data,
      refetchInterval: (query) => {
        const nextDeploy = query.state.data?.data.nextDeploy;
        if (nextDeploy) {
          const now = new Date();
          const nextDeployDate = new Date(nextDeploy);
          const diffMs = nextDeployDate.getTime() - now.getTime();
          return diffMs > 0 ? diffMs : 60 * 1000;
        }
        return 60 * 1000;
      },
      refetchIntervalInBackground: false,
      staleTime: 0,
      gcTime: 10 * 60 * 1000,
      ...queryOptions,
    },
  });
};
