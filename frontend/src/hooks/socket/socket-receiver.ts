import {
  getDeployementAPIListQueryKey,
  ServerMessageLogKind,
  ServerMessageNewDeploymentKind,
  ServerMessagePreviousLogsKind,
  ServerMessageStateKind,
  type ServerMessage,
} from '@/api';
import { type QueryClient } from '@tanstack/react-query';
import { onLogEvent } from './use-logs';

export function createSocketReceiver(queryClient: QueryClient) {
  return (event: MessageEvent) => {
    const message: ServerMessage = JSON.parse(event.data);

    switch (message.kind) {
      case ServerMessageStateKind.state:
        queryClient.setQueryData(['state'], message.value);
        break;

      case ServerMessageLogKind.log:
      case ServerMessagePreviousLogsKind.previousLogs:
        onLogEvent(message, queryClient);
        break;

      case ServerMessageNewDeploymentKind.newDeployment:
        queryClient.setQueryData(['deployment'], message.value);
        queryClient.invalidateQueries({ queryKey: [getDeployementAPIListQueryKey()] });
        break;
      default:
        throw new Error(`Unhandled server message: ${JSON.stringify(message)}`);
    }
  };
}
