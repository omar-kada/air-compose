import { ClientMessageEndLogsKind, ClientMessageStartLogsKind, type ClientMessage } from '@/api';
import { type RefObject } from 'react';

export interface SocketEmitter extends SocketBusinessEmitter {
  emit: (event: ClientMessage) => void;
  onOpen: () => void;
}

export interface SocketBusinessEmitter {
  startLogs: (previousLines?: number) => void;
  endLogs: () => void;
}

export type SocketEmitterActions = Omit<SocketEmitter, 'emit' | 'onOpen'>;

export function createSocketEmitter(socketRef: RefObject<WebSocket | null>): SocketEmitter {
  const pendingEvents: ClientMessage[] = [];

  const flushQueuedEvents = () => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }

    const queued = pendingEvents.splice(0);
    queued.forEach((event) => socket.send(JSON.stringify(event)));
  };

  const emit = (event: ClientMessage) => {
    const socket = socketRef.current;

    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(event));
      return;
    }

    pendingEvents.push(event);
  };

  return {
    emit,
    startLogs: (previousLines = 0) => {
      emit({ kind: ClientMessageStartLogsKind.startLogs, value: { previousLines } });
    },
    endLogs: () => {
      emit({ kind: ClientMessageEndLogsKind.endLogs, value: {} });
    },
    onOpen: flushQueuedEvents,
  };
}
