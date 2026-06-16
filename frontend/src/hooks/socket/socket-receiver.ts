import {
  EventType,
  getConfigAPIGetQueryKey,
  getDeployementAPIListQueryKey,
  getDeployementAPIReadQueryKey,
  getDiffAPIGetQueryKey,
  getNotificationsAPIListQueryKey,
  getStateAPIGetQueryKey,
  getStatusAPIGetQueryKey,
  ServerMessageEventKind,
  ServerMessageLogKind,
  ServerMessageNewDeploymentKind,
  ServerMessagePreviousLogsKind,
  type EventMessage,
  type ServerMessage,
} from '@/api';
import { type QueryClient } from '@tanstack/react-query';
import i18next from 'i18next';
import { toast } from 'sonner';
import { onLogEvent } from './use-logs';

export function createSocketReceiver(queryClient: QueryClient) {
  return (event: MessageEvent) => {
    const message: ServerMessage = JSON.parse(event.data);

    switch (message.kind) {
      case ServerMessageLogKind.log:
      case ServerMessagePreviousLogsKind.previousLogs:
        onLogEvent(message, queryClient);
        break;

      case ServerMessageNewDeploymentKind.newDeployment:
        queryClient.invalidateQueries({ queryKey: [getDeployementAPIListQueryKey()] });
        break;
      case ServerMessageEventKind.event:
        onEvent(queryClient, message.value);
        break;
      default:
        throw new Error(`Unhandled server message: ${JSON.stringify(message)}`);
    }
  };
}

function onEvent(queryClient: QueryClient, event: EventMessage) {
  if (event.deploymentId) {
    queryClient.refetchQueries({
      queryKey: getDeployementAPIReadQueryKey(`${event.deploymentId}`),
    });
  }
  queryClient.invalidateQueries({ queryKey: getNotificationsAPIListQueryKey() });
  queryClient.invalidateQueries({ queryKey: getNotificationsAPIListQueryKey() });
  switch (event.type) {
    case EventType.DEPLOYMENT_STARTED:
      queryClient.invalidateQueries({ queryKey: getDiffAPIGetQueryKey() });
      queryClient.refetchQueries({ queryKey: getDeployementAPIListQueryKey() });
      break;
    case EventType.DEPLOYMENT_ERROR:
    case EventType.DEPLOYMENT_SUCCESS:
      queryClient.refetchQueries({ queryKey: getDeployementAPIListQueryKey() });
      queryClient.refetchQueries({ queryKey: getStateAPIGetQueryKey() });
      break;
    case EventType.STACKS_HEALTHY:
    case EventType.STACKS_UNHEALTHY:
      queryClient.refetchQueries({ queryKey: getStateAPIGetQueryKey() });
      queryClient.refetchQueries({ queryKey: getStatusAPIGetQueryKey() });
      break;
    case EventType.CONFIGURATION_UPDATED:
      queryClient.invalidateQueries({ queryKey: getConfigAPIGetQueryKey() });
      queryClient.invalidateQueries({ queryKey: getDiffAPIGetQueryKey() });
      break;
    case EventType.ERROR:
      toast.error(i18next.t('ALERT.SERVER_ERROR'), {
        description: event.msg,
      });
  }
}
