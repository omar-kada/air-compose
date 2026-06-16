import { useQueryClient } from '@tanstack/react-query';
import { createContext, useEffect, useMemo, useRef } from 'react';
import { createSocketEmitter, type SocketBusinessEmitter } from './socket-emitter';
import { createSocketReceiver } from './socket-receiver';
import { useWsRetry, useWsStatus } from './use-ws';

export type WsStatus = 'off' | 'connected' | 'reconnecting' | 'failed';

export const WsContext = createContext<SocketBusinessEmitter | null>(null);

export const BASE_DELAY = 500;
export const MAX_DELAY = 30_000;
export const MAX_RETRIES = 10;

// WebSocketProvider.tsx
export function WebSocketProvider({
  url,
  enabled = true,
  children,
}: {
  url: string;
  enabled?: boolean;
  children: React.ReactNode;
}) {
  const queryClient = useQueryClient();
  const socketRef = useRef<WebSocket | null>(null);
  const isManuallyClosed = useRef(false);
  const emitter = useMemo(() => createSocketEmitter(socketRef), []);

  const { updateStatus } = useWsStatus(enabled);
  const { attempt, scheduleRetry, reset, cancel, retriesRef } = useWsRetry();

  useEffect(() => {
    if (!enabled) {
      updateStatus('off');
      return () => {};
    }

    isManuallyClosed.current = false;
    const socket = new WebSocket(url);
    socketRef.current = socket;

    socket.onclose = (event) => {
      socketRef.current = null;
      if (isManuallyClosed.current || event.code === 1000) return;
      const retried = scheduleRetry();
      updateStatus(retried ? 'reconnecting' : 'failed', retriesRef.current);
    };

    socket.onopen = () => {
      reset();
      updateStatus('connected');
      emitter.onOpen();
    };

    socket.onmessage = createSocketReceiver(queryClient);

    socket.onerror = () => socket.close();

    return () => {
      isManuallyClosed.current = true;
      cancel();
      socketRef.current?.close(1000, 'unmount');
      socketRef.current = null;
    };
  }, [enabled, url, queryClient, emitter, attempt]);

  return (
    <WsContext.Provider value={useMemo(() => ({ ...emitter }), [emitter])}>
      {children}
    </WsContext.Provider>
  );
}

// --- Hooks ---
