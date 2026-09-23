vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('sonner', async () => {
  const { createSonnerMock } = await import('@/tests/mock-factories');
  const m = createSonnerMock();
  (m.toast as any).dismiss = vi.fn();
  (m.toast as any).promise = vi.fn(() => Promise.resolve(true));
  return m;
});

vi.mock('i18next', () => ({
  default: { t: vi.fn((key: string) => `t:${key}`) },
}));

const mockOnLogEvent = vi.hoisted(() => vi.fn());
const mockIncrement = vi.hoisted(() => vi.fn());

vi.mock('..', () => ({ onLogEvent: mockOnLogEvent }));
vi.mock('../stacks', () => ({
  incrementUnreadCount: mockIncrement,
  refetchState: vi.fn(),
}));

import { useWs, getWsStatusKey, useWsStatusQuery, useWsStatus, useWsRetry } from './use-ws';
import { MAX_RETRIES } from './socket-provider';
import { renderHook } from '@testing-library/react';
import { renderHookWithQuery } from '@/tests/test-utils';

describe('getWsStatusKey', () => {
  it('returns ws-status query key', () => {
    expect(getWsStatusKey()).toEqual(['ws-status']);
  });
});

describe('useWs', () => {
  it('throws outside WebSocketProvider', () => {
    expect(() => renderHook(() => useWs())).toThrow('useWs must be used inside');
  });
});

describe('useWsStatusQuery', () => {
  it('returns off as initial data', () => {
    const { result } = renderHookWithQuery(() => useWsStatusQuery());
    expect(result.current.data).toBe('off');
  });
});

describe('useWsStatus', () => {
  it('returns status and updateStatus', () => {
    const { result } = renderHookWithQuery(() => useWsStatus(true));
    expect(result.current.status).toBe('off');
    expect(result.current.updateStatus).toBeInstanceOf(Function);
  });
});

describe('useWsRetry', () => {
  beforeEach(() => { vi.useFakeTimers(); });
  afterEach(() => { vi.useRealTimers(); });

  it('scheduleRetry returns true', () => {
    const { result } = renderHook(() => useWsRetry());
    expect(result.current.scheduleRetry()).toBe(true);
    result.current.cancel();
  });

  it('returns false after MAX_RETRIES', () => {
    const { result } = renderHook(() => useWsRetry());
    for (let i = 0; i < MAX_RETRIES; i++) {
      expect(result.current.scheduleRetry()).toBe(true);
    }
    expect(result.current.scheduleRetry()).toBe(false);
  });

  it('reset allows retrying', () => {
    const { result } = renderHook(() => useWsRetry());
    for (let i = 0; i < MAX_RETRIES; i++) {
      result.current.scheduleRetry();
    }
    result.current.reset();
    expect(result.current.scheduleRetry()).toBe(true);
    result.current.cancel();
  });
});
