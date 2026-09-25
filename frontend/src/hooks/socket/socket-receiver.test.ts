const mockOnLogEvent = vi.hoisted(() => vi.fn());
const mockIncrementUnreadCount = vi.hoisted(() => vi.fn());

vi.mock('..', () => ({
  onLogEvent: mockOnLogEvent,
}));
vi.mock('../stacks', () => ({
  incrementUnreadCount: mockIncrementUnreadCount,
}));
vi.mock('i18next', () => ({
  default: { t: vi.fn((k: string) => `t:${k}`) },
}));
vi.mock('sonner', async () => {
  const { createSonnerMock } = await import('@/tests/mock-factories');
  return createSonnerMock();
});

import { createSocketReceiver } from './socket-receiver';
import type { QueryClient } from '@tanstack/react-query';
import {
  ServerMessageLogKind,
  ServerMessagePreviousLogsKind,
  ServerMessageNewDeploymentKind,
  ServerMessageEventKind,
  EventType,
} from '@/api';
import { toast } from 'sonner';

const mockClient = () => ({
  setQueryData: vi.fn(),
  invalidateQueries: vi.fn(),
  refetchQueries: vi.fn(),
});

describe('createSocketReceiver', () => {
  it('routes log messages to onLogEvent', () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({ kind: ServerMessageLogKind.log, value: 'test' }),
    };
    onEvent(msg as unknown as MessageEvent);
    const parsed = JSON.parse(msg.data);
    expect(mockOnLogEvent).toHaveBeenCalledWith(parsed, client);
  });

  it('routes previousLogs to onLogEvent', () => {
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

  it('invalidates queries on newDeployment', () => {
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

  it('calls onLogEvent for event messages with deploymentId', () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessageEventKind.event,
        value: { deploymentId: '123', type: EventType.DEPLOYMENT_STARTED },
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    expect(client.invalidateQueries).toHaveBeenCalled();
  });

  it('calls toast.error on ERROR event', () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessageEventKind.event,
        value: { type: EventType.ERROR, msg: 'fail' },
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    expect(vi.mocked(toast.error)).toHaveBeenCalled();
  });

  it('handles event with isNotification: refetches deployment and increments unread', () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessageEventKind.event,
        value: {
          deploymentId: 123,
          isNotification: true,
          type: EventType.DEPLOYMENT_STARTED,
          msg: 'started',
        },
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    expect(client.refetchQueries).toHaveBeenCalled();
    expect(mockIncrementUnreadCount).toHaveBeenCalledWith(client);
    expect(client.invalidateQueries).toHaveBeenCalled();
  });

  it('handles DEPLOYMENT_ERROR event', () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessageEventKind.event,
        value: { type: EventType.DEPLOYMENT_ERROR, msg: 'error', isNotification: false },
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    expect(client.refetchQueries).toHaveBeenCalled();
  });

  it('handles DEPLOYMENT_SUCCESS event', () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessageEventKind.event,
        value: { type: EventType.DEPLOYMENT_SUCCESS, msg: 'done', isNotification: false },
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    expect(client.refetchQueries).toHaveBeenCalled();
  });

  it('handles HEALTH_CHANGE event', () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessageEventKind.event,
        value: { type: EventType.HEALTH_CHANGE, msg: 'health', isNotification: false },
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    expect(client.refetchQueries).toHaveBeenCalled();
  });

  it('handles CONFIGURATION_UPDATED event', () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessageEventKind.event,
        value: { type: EventType.CONFIGURATION_UPDATED, msg: 'config', isNotification: false },
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    expect(client.invalidateQueries).toHaveBeenCalled();
  });

  it('handles event without deploymentId or isNotification (no-op branches)', () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = {
      data: JSON.stringify({
        kind: ServerMessageEventKind.event,
        value: { type: EventType.MISC, msg: 'misc', isNotification: false },
      }),
    };
    onEvent(msg as unknown as MessageEvent);
    // Should not have called refetchQueries or invalidateQueries
    expect(client.refetchQueries).not.toHaveBeenCalled();
    expect(client.invalidateQueries).not.toHaveBeenCalled();
  });

  it('throws on unhandled message kind', () => {
    const client = mockClient();
    const onEvent = createSocketReceiver(client as unknown as QueryClient);
    const msg = { data: JSON.stringify({ kind: 'unknown' }) };
    expect(() => onEvent(msg as unknown as MessageEvent)).toThrow('Unhandled');
  });
});
