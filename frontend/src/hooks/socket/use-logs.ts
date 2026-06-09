import {
  ServerMessageLogKind,
  ServerMessagePreviousLogsKind,
  type LogMessages,
  type ServerMessageLog,
  type ServerMessagePreviousLogs,
} from '@/api';
import { QueryClient, useQuery, type QueryKey, type UseQueryOptions } from '@tanstack/react-query';
import { useEffect } from 'react';
import { useWs } from '..';

export function getLogsQueryKey(): QueryKey {
  return ['logs'];
}

export function useLogs(previousLines = 0, query?: Partial<UseQueryOptions<LogMessages>>) {
  const { startLogs, endLogs } = useWs();

  useEffect(() => {
    startLogs(previousLines);
    return () => endLogs();
  }, [startLogs, endLogs]);

  return useQuery<LogMessages>({
    queryKey: getLogsQueryKey(),
    queryFn: async (): Promise<LogMessages> => [],
    enabled: true,
    ...query,
  });
}

export function onLogEvent(
  serverEvent: ServerMessageLog | ServerMessagePreviousLogs,
  queryClient: QueryClient,
) {
  switch (serverEvent.kind) {
    case ServerMessageLogKind.log:
      queryClient.setQueryData(getLogsQueryKey(), (prev: LogMessages = []) => [
        ...prev,
        serverEvent.value,
      ]);
      break;

    case ServerMessagePreviousLogsKind.previousLogs:
      queryClient.setQueryData(getLogsQueryKey(), serverEvent.value);
      break;
  }
}
