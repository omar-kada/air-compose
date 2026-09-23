import { createSocketEmitter } from "./socket-emitter";
import type { Mock } from "vitest";

describe("createSocketEmitter", () => {
  const createMockSocketRef = (readyState: number) => ({
    current: {
      readyState,
      send: vi.fn(),
    } as Pick<WebSocket, "readyState" | "send">,
  });

  it("emit sends message when socket is OPEN", () => {
    const socketRef = createMockSocketRef(WebSocket.OPEN);
    const emitter = createSocketEmitter(
      socketRef as unknown as { current: { readyState: number; send: Mock } },
    );
    emitter.emit({ kind: "test" } as unknown as Parameters<typeof emitter.emit>[0]);
    expect(socketRef.current.send).toHaveBeenCalledWith(
      JSON.stringify({ kind: "test" }),
    );
  });

  it("emit queues message when socket is not OPEN", () => {
    const socketRef = createMockSocketRef(WebSocket.CLOSED);
    const emitter = createSocketEmitter(
      socketRef as unknown as { current: { readyState: number; send: Mock } },
    );
    emitter.emit({ kind: "queued" } as unknown as Parameters<typeof emitter.emit>[0]);
    expect(socketRef.current.send).not.toHaveBeenCalled();
  });

  it("onOpen flushes queued events when socket opens", () => {
    const socketRef = createMockSocketRef(WebSocket.CLOSED);
    const emitter = createSocketEmitter(
      socketRef as unknown as { current: { readyState: number; send: Mock } },
    );
    emitter.startLogs(5);
    emitter.endLogs();
    expect(socketRef.current.send).not.toHaveBeenCalled();
    socketRef.current.readyState = WebSocket.OPEN;
    emitter.onOpen();
    expect(socketRef.current.send).toHaveBeenCalledTimes(2);
  });

  it("startLogs emits startLogs kind with previousLines", () => {
    const socketRef = createMockSocketRef(WebSocket.OPEN);
    const emitter = createSocketEmitter(
      socketRef as unknown as { current: { readyState: number; send: Mock } },
    );
    emitter.startLogs(10);
    const sent = JSON.parse(socketRef.current.send!.mock.calls[0][0] as string);
    expect(sent.value.previousLines).toBe(10);
  });

  it("endLogs emits endLogs kind with empty value", () => {
    const socketRef = createMockSocketRef(WebSocket.OPEN);
    const emitter = createSocketEmitter(
      socketRef as unknown as { current: { readyState: number; send: Mock } },
    );
    emitter.endLogs();
    const sent = JSON.parse(socketRef.current.send!.mock.calls[0][0] as string);
    expect(sent.value).toEqual({});
  });

  it("onOpen does nothing when socket remains closed", () => {
    const socketRef = createMockSocketRef(WebSocket.CLOSED);
    const emitter = createSocketEmitter(
      socketRef as unknown as { current: { readyState: number; send: Mock } },
    );
    emitter.emit({ kind: "queued" } as unknown as Parameters<typeof emitter.emit>[0]);
    emitter.onOpen();
    expect(socketRef.current.send).not.toHaveBeenCalled();
  });
});
