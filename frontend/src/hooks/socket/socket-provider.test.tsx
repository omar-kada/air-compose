interface MockWebSocketInstance {
  readyState: number;
  send: ReturnType<typeof vi.fn>;
  close: ReturnType<typeof vi.fn>;
  onopen: ((event: Event) => void) | null;
  onclose: ((event: CloseEvent) => void) | null;
  onerror: ((event: Event) => void) | null;
  onmessage: ((event: MessageEvent) => void) | null;
}

const mockRetryState = vi.hoisted(() => ({
  attempt: 0,
  scheduleRetry: vi.fn(() => false),
  reset: vi.fn(),
  cancel: vi.fn(),
  retriesRef: { current: 0 },
}));
const mockStatusState = vi.hoisted(() => ({
  status: 'off',
  updateStatus: vi.fn(),
}));
const mockSocketRefs = vi.hoisted(() => ({
  onLogEvent: vi.fn(),
  incrementUnreadCount: vi.fn(),
  refetchState: vi.fn(),
}));
const mockEmitter = vi.hoisted(() => ({
  emit: vi.fn(),
  startLogs: vi.fn(),
  endLogs: vi.fn(),
  onOpen: vi.fn(),
}));

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
vi.mock('i18next', async () => {
  const { createI18nextMock } = await import('@/tests/mock-factories');
  return createI18nextMock();
});
vi.mock('../stacks', () => ({
  onLogEvent: mockSocketRefs.onLogEvent,
  incrementUnreadCount: mockSocketRefs.incrementUnreadCount,
  refetchState: mockSocketRefs.refetchState,
}));
vi.mock('./socket-emitter', () => ({
  createSocketEmitter: vi.fn(() => mockEmitter),
}));
vi.mock('./socket-receiver', () => ({
  createSocketReceiver: vi.fn(() => vi.fn()),
}));
vi.mock('./use-ws', () => ({
  useWsRetry: vi.fn(() => mockRetryState),
  useWsStatus: vi.fn(() => mockStatusState),
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

  beforeEach(() => {
    vi.clearAllMocks();
    mockRetryState.scheduleRetry.mockReturnValue(false);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

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
    vi.stubGlobal('WebSocket', MockWebSocket as unknown as typeof WebSocket);
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
  });

  it('cleans up WebSocket on unmount', async () => {
    const MockWebSocket = createMockWebSocket();
    vi.stubGlobal('WebSocket', MockWebSocket as unknown as typeof WebSocket);
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
  });

  it('calls onopen handler when socket opens', async () => {
    const MockWebSocket = createMockWebSocket();
    vi.stubGlobal('WebSocket', MockWebSocket as unknown as typeof WebSocket);
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
    expect(socket.onopen).not.toBeNull();
    if (socket.onopen) {
      socket.onopen({} as Event);
    }
    expect(mockRetryState.reset).toHaveBeenCalled();
    expect(mockStatusState.updateStatus).toHaveBeenCalledWith('connected');
    expect(mockEmitter.onOpen).toHaveBeenCalled();
  });

  it('does not schedule retry on normal close (code 1000)', async () => {
    const MockWebSocket = createMockWebSocket();
    vi.stubGlobal('WebSocket', MockWebSocket as unknown as typeof WebSocket);
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
    expect(socket.onclose).not.toBeNull();
    if (socket.onclose) {
      socket.onclose({ code: 1000 } as CloseEvent);
    }
    expect(mockRetryState.scheduleRetry).not.toHaveBeenCalled();
    expect(mockStatusState.updateStatus).not.toHaveBeenCalled();
    expect(mockSocketRefs.refetchState).not.toHaveBeenCalled();
  });

  it('calls onerror handler which closes the socket', async () => {
    const MockWebSocket = createMockWebSocket();
    vi.stubGlobal('WebSocket', MockWebSocket as unknown as typeof WebSocket);
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
    expect(socket.onerror).not.toBeNull();
    if (socket.onerror) {
      socket.onerror({} as Event);
    }
    expect(socket.close).toHaveBeenCalled();
  });

  it('does not create WebSocket when enabled is false', () => {
    const MockWebSocket = createMockWebSocket();
    vi.stubGlobal('WebSocket', MockWebSocket as unknown as typeof WebSocket);
    render(
      <QueryClientProvider client={new QueryClient()}>
        <WebSocketProvider url="ws://test" enabled={false}>
          <div>child</div>
        </WebSocketProvider>
      </QueryClientProvider>,
    );
    expect(MockWebSocket).not.toHaveBeenCalled();
  });
});
