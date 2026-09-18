import { AxiosError } from 'axios';
import type { AxiosResponse } from 'axios';
import { ErrorCode } from '@/api/api';
import { cn, formatTime, isInvalidToken } from './utils';

describe('cn', () => {
  it('joins class names with a single space', () => {
    expect(cn('a', 'b', 'c')).toBe('a b c');
  });

  it('drops falsy class arguments', () => {
    expect(cn('a', false, '', null, undefined, 0, 'b')).toBe('a b');
  });

  it('supports clsx conditional objects', () => {
    expect(cn('a', { b: true, c: false })).toBe('a b');
  });

  it('deduplicates conflicting tailwind classes (last wins)', () => {
    expect(cn('text-red-500', 'text-blue-500')).toBe('text-blue-500');
  });
});

describe('formatTime', () => {
  it('returns an HH:mm:ss 24-hour string', () => {
    const result = formatTime('2024-01-15T10:30:45.000Z', 'en-GB');
    expect(result).toMatch(/^\d{2}:\d{2}:\d{2}$/);
  });

  it('never uses an AM/PM label (hour12: false)', () => {
    const result = formatTime('2024-01-15T22:15:05.000Z', 'en-US');
    expect(result).not.toMatch(/[AP]M/);
  });
});

describe('isInvalidToken', () => {
  const makeResponse = (status: number, code: ErrorCode): AxiosResponse =>
    ({ status, data: { code } }) as AxiosResponse;

  it('returns true for a 401 with INVALID_TOKEN', () => {
    const error = new AxiosError(
      'Unauthorized',
      ErrorCode.INVALID_TOKEN,
      undefined,
      undefined,
      makeResponse(401, ErrorCode.INVALID_TOKEN),
    );
    expect(isInvalidToken(error)).toBe(true);
  });

  it('returns false for a 401 with a different error code', () => {
    const error = new AxiosError(
      'Unauthorized',
      ErrorCode.INVALID_CREDENTIALS,
      undefined,
      undefined,
      makeResponse(401, ErrorCode.INVALID_CREDENTIALS),
    );
    expect(isInvalidToken(error)).toBe(false);
  });

  it('returns false for a 403 with INVALID_TOKEN', () => {
    const error = new AxiosError(
      'Forbidden',
      ErrorCode.INVALID_TOKEN,
      undefined,
      undefined,
      makeResponse(403, ErrorCode.INVALID_TOKEN),
    );
    expect(isInvalidToken(error)).toBe(false);
  });

  it('returns false for non-AxiosError values', () => {
    expect(isInvalidToken(new Error('boom'))).toBe(false);
    expect(isInvalidToken('string')).toBe(false);
    expect(isInvalidToken(null)).toBe(false);
  });

  it('returns false when the AxiosError has no response', () => {
    const error = new AxiosError('Network Error');
    expect(isInvalidToken(error)).toBe(false);
  });
});
