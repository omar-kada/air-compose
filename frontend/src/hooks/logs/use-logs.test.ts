vi.mock('..', () => ({
  useWs: vi.fn(() => ({ startLogs: vi.fn(), endLogs: vi.fn() })),
  useWsStatusQuery: vi.fn(() => ({ data: 'connected' })),
}));

import { useLogs, onLogEvent, getLogsQueryKey } from './use-logs';
import { ServerMessageLogKind, ServerMessagePreviousLogsKind } from '@/api';
import { renderHookWithQuery } from '@/tests/test-utils';

describe('getLogsQueryKey', () => {
  it('returns ["logs"]', () => {
    expect(getLogsQueryKey()).toEqual(['logs']);
  });
});

describe('useLogs', () => {
  it('returns query result', () => {
    const { result } = renderHookWithQuery(() => useLogs(0));
    expect(result.current).toBeDefined();
  });
});

describe('onLogEvent', () => {
  it('appends log events via updater', () => {
    const setQueryData = vi.fn();
    const queryClient = { setQueryData } as any;
    const event = { kind: ServerMessageLogKind.log, value: 'test log' };
    onLogEvent(event as any, queryClient);
    expect(setQueryData).toHaveBeenCalledWith(['logs'], expect.any(Function));
    const updater = setQueryData.mock.calls[0][1];
    expect(updater(['prev'])).toEqual(['prev', 'test log']);
    expect(updater([])).toEqual(['test log']);
  });

  it('replaces data with previousLogs value', () => {
    const setQueryData = vi.fn();
    const queryClient = { setQueryData } as any;
    const event = { kind: ServerMessagePreviousLogsKind.previousLogs, value: ['log1', 'log2'] };
    onLogEvent(event as any, queryClient);
    expect(setQueryData).toHaveBeenCalledWith(['logs'], ['log1', 'log2']);
  });
});
