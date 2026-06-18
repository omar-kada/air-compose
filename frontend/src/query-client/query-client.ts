import { authAPIRefresh } from '@/api/api';
import { isInvalidToken, ROUTES } from '@/lib';
import { QueryCache, QueryClient } from '@tanstack/react-query';
import { AxiosError } from 'axios';
import debounce from 'lodash.debounce';

export const debouncedRefreshToken = debounce(refreshToken, 5000, { leading: true }); // 5 seconds debounce

export const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (error, _) => {
      if (isInvalidToken(error)) {
        debouncedRefreshToken();
      }
    },
  }),
  defaultOptions: {
    queries: {
      retry: 0,
      refetchOnWindowFocus: false,
      staleTime: 1000 * 60, // 1 min
      gcTime: 10 * 60 * 1000, // 10 min
    },
  },
});

function refreshToken() {
  authAPIRefresh()
    .then((success) => success && queryClient.refetchQueries())
    .catch((err: AxiosError) => {
      if (err.status === 401) {
        window.location.href = ROUTES.LOGIN;
      }
    });
}
