vi.mock("react-i18next", async () => {
  const { createI18nMock } = await import("@/tests/mock-factories");
  return createI18nMock();
});
vi.mock("sonner", async () => {
  const { createSonnerMock } = await import("@/tests/mock-factories");
  const m = createSonnerMock();
  m.toast.dismiss = vi.fn();
  m.toast.promise = vi.fn(() => Promise.resolve(true));
  return m;
});
vi.mock("i18next", () => ({
  default: { t: vi.fn((k: string) => `t:${k}`) },
}));
const mockOnLogEvent = vi.hoisted(() => vi.fn());
vi.mock("..", () => ({ onLogEvent: mockOnLogEvent }));
vi.mock("../stacks", () => ({
  incrementUnreadCount: vi.fn(),
  refetchState: vi.fn(),
}));
vi.mock("./socket-emitter", () => ({
  createSocketEmitter: vi.fn(() => ({
    emit: vi.fn(),
    startLogs: vi.fn(),
    endLogs: vi.fn(),
    onOpen: vi.fn(),
  })),
}));
vi.mock("./socket-receiver", () => ({
  createSocketReceiver: vi.fn(() => vi.fn()),
}));
vi.mock("./use-ws", () => ({
  useWsRetry: vi.fn(() => ({
    attempt: 0,
    scheduleRetry: vi.fn(() => true),
    reset: vi.fn(),
    cancel: vi.fn(),
    retriesRef: { current: 0 },
  })),
  useWsStatus: vi.fn(() => ({ status: "off", updateStatus: vi.fn() })),
}));

import { render, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { WebSocketProvider } from "./socket-provider";

describe("WebSocketProvider", () => {
  it("renders children when enabled is false", () => {
    const { getByText } = render(
      <QueryClientProvider client={new QueryClient()}>
        <WebSocketProvider url="ws://test" enabled={false}>
          <div>child</div>
        </WebSocketProvider>
      </QueryClientProvider>,
    );
    expect(getByText("child")).toBeDefined();
  });

  it("creates WebSocket when enabled is true", async () => {
    const origWS = (globalThis as { WebSocket: typeof WebSocket }).WebSocket;
    const MockWebSocket = vi.fn(function (this: any, _url: string) {
      this.readyState = 1;
      this.send = vi.fn();
      this.close = vi.fn();
    });
    (globalThis as { WebSocket: typeof WebSocket }).WebSocket =
      MockWebSocket as unknown as typeof WebSocket;
    try {
      render(
        <QueryClientProvider client={new QueryClient()}>
          <WebSocketProvider url="ws://test" enabled={true}>
            <div>child</div>
          </WebSocketProvider>
        </QueryClientProvider>,
      );
      await waitFor(() => {
        expect(MockWebSocket).toHaveBeenCalledWith("ws://test");
      });
    } finally {
      (globalThis as { WebSocket: typeof WebSocket }).WebSocket = origWS;
    }
  });

  it("cleans up WebSocket on unmount", async () => {
    const origWS = (globalThis as { WebSocket: typeof WebSocket }).WebSocket;
    const MockWebSocket = vi.fn(function (this: any) {
      this.readyState = 1;
      this.send = vi.fn();
      this.close = vi.fn();
    });
    (globalThis as { WebSocket: typeof WebSocket }).WebSocket =
      MockWebSocket as unknown as typeof WebSocket;
    try {
      const { unmount } = render(
        <QueryClientProvider client={new QueryClient()}>
          <WebSocketProvider url="ws://test" enabled={true}>
            <div>child</div>
          </WebSocketProvider>
        </QueryClientProvider>,
      );
      await waitFor(() => {
        expect(MockWebSocket).toHaveBeenCalled();
      });
      const socket = MockWebSocket.mock.instances[0] as unknown as WebSocket;
      unmount();
      expect(socket.close).toHaveBeenCalled();
    } finally {
      (globalThis as { WebSocket: typeof WebSocket }).WebSocket = origWS;
    }
  });
});
