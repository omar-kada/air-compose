vi.mock('sonner', async () => {
  const { createSonnerMock } = await import('@/tests/mock-factories');
  const mock = createSonnerMock();
  mock.toast.success = vi.fn(() => 'toast-id');
  mock.toast.error = vi.fn(() => 'toast-id');
  mock.toast.loading = vi.fn(() => 'toast-id');
  mock.toast.dismiss = vi.fn();
  return mock;
});

import { toast } from 'sonner';
import { wsToast, wsToastOnStatus } from './socket-toast';
import type { TFunction } from 'i18next';

describe('wsToast', () => {
  beforeEach(() => {
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
    vi.mocked(toast.loading).mockClear();
    vi.mocked(toast.dismiss).mockClear?.();
  });

  it('connected calls toast.success', () => {
    wsToast.connected(vi.fn() as unknown as TFunction);
    expect(vi.mocked(toast.success)).toHaveBeenCalled();
  });

  it('reconnecting calls toast.loading', () => {
    wsToast.reconnecting(vi.fn() as unknown as TFunction, 1, 10);
    expect(vi.mocked(toast.loading)).toHaveBeenCalled();
  });

  it('failed calls toast.error', () => {
    wsToast.failed(vi.fn() as unknown as TFunction);
    expect(vi.mocked(toast.error)).toHaveBeenCalled();
  });

  it('dismiss calls toast.dismiss', () => {
    wsToast.connected(vi.fn() as unknown as TFunction);
    wsToast.dismiss();
    expect(vi.mocked(toast.dismiss)).toHaveBeenCalled();
  });
});

describe('wsToastOnStatus', () => {
  beforeEach(() => {
    wsToast.connected(vi.fn() as unknown as TFunction);
    wsToast.dismiss();
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
    vi.mocked(toast.loading).mockClear();
    vi.mocked(toast.dismiss).mockClear?.();
  });

  it('dismisses on off status', () => {
    wsToast.connected(vi.fn() as unknown as TFunction);
    wsToastOnStatus(vi.fn() as unknown as TFunction, 'connected', 'off');
    expect(vi.mocked(toast.dismiss)).toHaveBeenCalled();
  });

  it('shows connected toast when reconnecting->connected', () => {
    wsToastOnStatus(vi.fn() as unknown as TFunction, 'reconnecting', 'connected');
    expect(vi.mocked(toast.success)).toHaveBeenCalled();
  });

  it('dismisses when connected from non-reconnecting', () => {
    wsToastOnStatus(vi.fn() as unknown as TFunction, 'off', 'connected');
    expect(vi.mocked(toast.success)).not.toHaveBeenCalled();
  });

  it('shows reconnecting toast', () => {
    wsToastOnStatus(vi.fn() as unknown as TFunction, 'connected', 'reconnecting', 1, 10);
    expect(vi.mocked(toast.loading)).toHaveBeenCalled();
  });

  it('shows failed toast', () => {
    wsToastOnStatus(vi.fn() as unknown as TFunction, 'connected', 'failed');
    expect(vi.mocked(toast.error)).toHaveBeenCalled();
  });
});
