import { useQueryClient } from '@tanstack/react-query';
import { createContext, useContext, useEffect, useMemo, useRef } from 'react';
import {
  createSocketEmitter,
  type SocketBusinessEmitter,
  type SocketEmitterActions,
} from './socket-emitter';
import { createSocketReceiver } from './socket-receiver';

const WsContext = createContext<SocketBusinessEmitter | null>(null);

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

  const emitter = useMemo(() => createSocketEmitter(socketRef), [socketRef]);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    const socket = new WebSocket(url);
    socketRef.current = socket;

    // Flush any queued events once connection opens
    socket.onopen = emitter.onOpen;
    socket.onmessage = createSocketReceiver(queryClient);

    return () => {
      socket.close();
      socketRef.current = null;
    };
  }, [enabled, queryClient, url]);

  return <WsContext.Provider value={emitter}>{children}</WsContext.Provider>;
}

export function useWs(): SocketEmitterActions {
  const ctx = useContext(WsContext);
  if (!ctx) throw new Error('useWs must be used inside <WebSocketProvider>');
  return ctx;
}
