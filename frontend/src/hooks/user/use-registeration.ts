import { getAuthAPIRegisteredQueryOptions } from '@/api/api';
import { useQuery } from '@tanstack/react-query';

export function useRegisteration() {
  return useQuery(
    getAuthAPIRegisteredQueryOptions({
      query: {
        select: (data) => {
          return data.data;
        },
      },
    }),
  );
}
