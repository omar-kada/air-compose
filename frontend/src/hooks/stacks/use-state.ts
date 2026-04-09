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
      refetchInterval: 20 * 1000,
      refetchIntervalInBackground: false,
      staleTime: 0,
      gcTime: 10 * 60 * 1000,
      ...queryOptions,
    },
  });
};
