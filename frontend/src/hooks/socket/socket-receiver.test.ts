const mockOnLogEvent = vi.hoisted(() => vi.fn());
const mockIncrementUnreadCount = vi.hoisted(() => vi.fn());

vi.mock("..", () => ({
  onLogEvent: mockOnLogEvent,
}));
vi.mock("../stacks", () => ({
  incrementUnreadCount: mockIncrementUnreadCount,
}));
vi.mock("i18next", () => ({
  default: { t: vi.fn((k: string) => `t:${k}`) },
}));
vi.mock("sonner", async () => {
  const { createSonnerMock } = await import("@/tests/mock-factories");
  return createSonnerMock();
});

import { createSocketReceiver } from "./socket-receiver";
import type { QueryClient } from "@tanstack/react-query";
import {
  ServerMessageLogKind,
  ServerMessagePreviousLogsKind,
  ServerMessageNewDeploymentKind,
  ServerMessageEventKind,
  EventType,
} from "@/api";
import { toast } from "sonner";

const mockClient = () => ({
  setQueryData: vi.fn(),
  invalidateQueries: vi.fn(),
  refetchQueries: vi.fn(),
});

describe("createSocketReceiver", () => {
  it("routes log messages to onLogEvent", () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({ kind: ServerMessageLogKind.log, value: "test" }),
    };
    onEvent(msg as unknown as MessageEvent);
    const parsed = JSON.parse(msg.data);
    expect(mockOnLogEvent).toHaveBeenCalledWith(parsed, client);
  });

  it("routes previousLogs to onLogEvent", () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessagePreviousLogsKind.previousLogs,
        value: [],
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    const parsed = JSON.parse(msg.data);
    expect(mockOnLogEvent).toHaveBeenCalledWith(parsed, client);
  });

  it("invalidates queries on newDeployment", () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessageNewDeploymentKind.newDeployment,
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    expect(client.invalidateQueries).toHaveBeenCalled();
  });

  it("calls onLogEvent for event messages with deploymentId", () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessageEventKind.event,
        value: { deploymentId: "123", type: EventType.DEPLOYMENT_STARTED },
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    expect(client.invalidateQueries).toHaveBeenCalled();
  });

  it("calls toast.error on ERROR event", () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessageEventKind.event,
        value: { type: EventType.ERROR, msg: "fail" },
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    expect(vi.mocked(toast.error)).toHaveBeenCalled();
  });

  it("throws on unhandled message kind", () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = { data: JSON.stringify({ kind: "unknown" }) };
    expect(() => onEvent(msg as unknown as MessageEvent)).toThrow("Unhandled");
  });
});
