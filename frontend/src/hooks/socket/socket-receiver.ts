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

function assertNever(x: never): never {
  throw new Error('Unhandled server message: ' + JSON.stringify(x));
}

export function createSocketReceiver(queryClient: QueryClient) {
  return (event: MessageEvent) => {
    const serverEvent: ServerMessage = JSON.parse(event.data);

    switch (serverEvent.kind) {
      case ServerMessageStateKind.state:
        queryClient.setQueryData(['state'], serverEvent.value);
        break;

      case ServerMessageLogKind.log:
      case ServerMessagePreviousLogsKind.previousLogs:
        onLogEvent(serverEvent, queryClient);
        break;

      case ServerMessageNewDeploymentKind.newDeployment:
        queryClient.setQueryData(['deployment'], serverEvent.value);
        queryClient.invalidateQueries({ queryKey: [getDeployementAPIListQueryKey()] });
        break;
      default:
        assertNever(serverEvent);
    }
  };
}
