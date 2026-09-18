import { renderHook, act } from '@testing-library/react';
import { Level } from '@/api';
import { LEVELS } from './log-levels';
import { useLogFilter } from './use-filter-logs';

describe('useLogFilter', () => {
  const debug = { level: Level.DEBUG, msg: 'debug msg', time: '1', meta: {} };
  const info = { level: Level.INFO, msg: 'info msg', time: '2', meta: {} };
  const warn = { level: Level.WARN, msg: 'warn msg', time: '3', meta: {} };
  const error = { level: Level.ERROR, msg: 'error msg', time: '4', meta: {} };
  const logs = [debug, info, warn, error];

  it('starts with all levels active and no text filter', () => {
    const { result } = renderHook(() => useLogFilter(logs));
    expect(result.current.text).toBe('');
    expect(result.current.activeLevels).toEqual(new Set(LEVELS));
    expect(result.current.filtered).toHaveLength(4);
  });

  it('excludes levels that are deactivated', () => {
    const { result } = renderHook(() => useLogFilter(logs));
    act(() => result.current.setActiveLevels(new Set([Level.ERROR])));
    expect(result.current.filtered).toHaveLength(1);
    expect(result.current.filtered[0]).toBe(error);
  });

  it('keeps an entry whose message matches the text filter case-insensitively', () => {
    const { result } = renderHook(() => useLogFilter(logs));
    act(() => result.current.setText('INFO'));
    expect(result.current.filtered).toHaveLength(1);
    expect(result.current.filtered[0]).toBe(info);
  });

  it('returns empty when the text filter matches no message', () => {
    const { result } = renderHook(() => useLogFilter(logs));
    act(() => result.current.setText('nope'));
    expect(result.current.filtered).toHaveLength(0);
  });

  it('composes text and level filters together', () => {
    const { result } = renderHook(() => useLogFilter(logs));
    act(() => {
      result.current.setText('error');
      result.current.setActiveLevels(new Set([Level.DEBUG, Level.ERROR]));
    });
    expect(result.current.filtered).toEqual([error]);
  });

  it('tracks the provided text across multiple changes', () => {
    const { result } = renderHook(() => useLogFilter(logs));
    act(() => result.current.setText('a'));
    expect(result.current.text).toBe('a');
    act(() => result.current.setText(''));
    expect(result.current.filtered).toHaveLength(4);
  });
});
