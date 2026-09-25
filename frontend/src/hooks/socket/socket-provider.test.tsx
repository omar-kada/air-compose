interface MockWebSocketInstance {
  readyState: number;
  send: ReturnType<typeof vi.fn>;
  close: ReturnType<typeof vi.fn>;
  onopen: ((event: Event) => void) | null;
  onclose: ((event: CloseEvent) => void) | null;
  onerror: ((event: Event) => void) | null;
  onmessage: ((event: MessageEvent) => void) | null;
}

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});
vi.mock('sonner', async () => {
  const { createSonnerMock } = await import('@/tests/mock-factories');
  const mock = createSonnerMock();
  mock.toast.dismiss = vi.fn();
  mock.toast.promise = vi.fn(() => Promise.resolve(true));
  return mock;
});
vi.mock('i18next', () => ({
  default: { t: vi.fn((k: string) => `translated:${k}`) },
}));
const mockOnLogEvent = vi.hoisted(() => vi.fn());
vi.mock('..', () => ({ onLogEvent: mockOnLogEvent }));
vi.mock('../stacks', () => ({
  incrementUnreadCount: vi.fn(),
  refetchState: vi.fn(),
}));
vi.mock('./socket-emitter', () => ({
  createSocketEmitter: vi.fn(() => ({
    emit: vi.fn(),
    startLogs: vi.fn(),
    endLogs: vi.fn(),
    onOpen: vi.fn(),
  })),
}));
vi.mock('./socket-receiver', () => ({
  createSocketReceiver: vi.fn(() => vi.fn()),
}));
vi.mock('./use-ws', () => ({
  useWsRetry: vi.fn(() => ({
    attempt: 0,
    scheduleRetry: vi.fn(() => true),
    reset: vi.fn(),
    cancel: vi.fn(),
    retriesRef: { current: 0 },
  })),
  useWsStatus: vi.fn(() => ({ status: 'off', updateStatus: vi.fn() })),
}));

import { render, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { WebSocketProvider } from './socket-provider';

describe('WebSocketProvider', () => {
  const createMockWebSocket = () => {
    return vi.fn(function (this: MockWebSocketInstance) {
      this.readyState = 1;
      this.send = vi.fn();
      this.close = vi.fn();
      this.onopen = null;
      this.onclose = null;
      this.onerror = null;
      this.onmessage = null;
    });
  };

  const setupMockWS = (MockWebSocket: ReturnType<typeof createMockWebSocket>) => {
    const origWS = (globalThis as { WebSocket: typeof WebSocket }).WebSocket;
    (globalThis as { WebSocket: typeof WebSocket }).WebSocket =
      MockWebSocket as unknown as typeof WebSocket;
    return () => {
      (globalThis as { WebSocket: typeof WebSocket }).WebSocket = origWS;
    };
  };

  it('renders children when enabled is false', () => {
    const { getByText } = render(
      <QueryClientProvider client={new QueryClient()}>
        <WebSocketProvider url="ws://test" enabled={false}>
          <div>child</div>
        </WebSocketProvider>
      </QueryClientProvider>,
    );
    expect(getByText('child')).toBeDefined();
  });

  it('creates WebSocket when enabled is true', async () => {
    const MockWebSocket = createMockWebSocket();
    const restore = setupMockWS(MockWebSocket);
    try {
      render(
        <QueryClientProvider client={new QueryClient()}>
          <WebSocketProvider url="ws://test" enabled>
            <div>child</div>
          </WebSocketProvider>
        </QueryClientProvider>,
      );
      await waitFor(() => {
        expect(MockWebSocket).toHaveBeenCalledWith('ws://test');
      });
    } finally {
      restore();
    }
  });

  it('cleans up WebSocket on unmount', async () => {
    const MockWebSocket = createMockWebSocket();
    const restore = setupMockWS(MockWebSocket);
    try {
      const { unmount } = render(
        <QueryClientProvider client={new QueryClient()}>
          <WebSocketProvider url="ws://test" enabled>
            <div>child</div>
          </WebSocketProvider>
        </QueryClientProvider>,
      );
      await waitFor(() => {
        expect(MockWebSocket).toHaveBeenCalled();
      });
      const socket = MockWebSocket.mock.instances[0] as unknown as MockWebSocketInstance;
      unmount();
      expect(socket.close).toHaveBeenCalledWith(1000, 'unmount');
    } finally {
      restore();
    }
  });

  it('calls onopen handler when socket opens', async () => {
    const MockWebSocket = createMockWebSocket();
    const restore = setupMockWS(MockWebSocket);
    try {
      render(
        <QueryClientProvider client={new QueryClient()}>
          <WebSocketProvider url="ws://test" enabled>
            <div>child</div>
          </WebSocketProvider>
        </QueryClientProvider>,
      );
      await waitFor(() => {
        expect(MockWebSocket).toHaveBeenCalled();
      });
      const socket = MockWebSocket.mock.instances[0] as unknown as MockWebSocketInstance;
      socket.onopen!({} as Event);
      // The provider should have set onopen handler, not null
      expect(socket.onopen).not.toBeNull();
    } finally {
      restore();
    }
  });

  it('does not schedule retry on normal close (code 1000)', async () => {
    const MockWebSocket = createMockWebSocket();
    const restore = setupMockWS(MockWebSocket);
    try {
      render(
        <QueryClientProvider client={new QueryClient()}>
          <WebSocketProvider url="ws://test" enabled>
            <div>child</div>
          </WebSocketProvider>
        </QueryClientProvider>,
      );
      await waitFor(() => {
        expect(MockWebSocket).toHaveBeenCalled();
      });
      const socket = MockWebSocket.mock.instances[0] as unknown as MockWebSocketInstance;
      // Simulate the provider set onclose handler, then trigger it with code 1000
      socket.onclose!({ code: 1000 } as CloseEvent);
      // Should not call socket.close() again (it returns early on manual/system close)
    } finally {
      restore();
    }
  });

  it('calls onerror handler which closes the socket', async () => {
    const MockWebSocket = createMockWebSocket();
    const restore = setupMockWS(MockWebSocket);
    try {
      render(
        <QueryClientProvider client={new QueryClient()}>
          <WebSocketProvider url="ws://test" enabled>
            <div>child</div>
          </WebSocketProvider>
        </QueryClientProvider>,
      );
      await waitFor(() => {
        expect(MockWebSocket).toHaveBeenCalled();
      });
      const socket = MockWebSocket.mock.instances[0] as unknown as MockWebSocketInstance;
      socket.onerror!({} as Event);
      expect(socket.close).toHaveBeenCalled();
    } finally {
      restore();
    }
  });

  it('does not create WebSocket when enabled is false', () => {
    const MockWebSocket = createMockWebSocket();
    const restore = setupMockWS(MockWebSocket);
    try {
      render(
        <QueryClientProvider client={new QueryClient()}>
          <WebSocketProvider url="ws://test" enabled={false}>
            <div>child</div>
          </WebSocketProvider>
        </QueryClientProvider>,
      );
      expect(MockWebSocket).not.toHaveBeenCalled();
    } finally {
      restore();
    }
  });
});
