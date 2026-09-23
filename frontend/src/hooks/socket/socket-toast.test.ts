vi.mock('sonner', async () => {
  const { createSonnerMock } = await import('@/tests/mock-factories');
  const m = createSonnerMock();
  m.toast.success = vi.fn(() => 'toast-id');
  m.toast.error = vi.fn(() => 'toast-id');
  m.toast.loading = vi.fn(() => 'toast-id');
  (m.toast as any).dismiss = vi.fn();
  return m;
});

import { toast } from 'sonner';
import { wsToast, wsToastOnStatus } from './socket-toast';

describe('wsToast', () => {
  beforeEach(() => {
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
    vi.mocked(toast.loading).mockClear();
    vi.mocked(toast.dismiss as any).mockClear?.();
  });

  it('connected calls toast.success', () => {
    wsToast.connected(vi.fn() as any);
    expect(vi.mocked(toast.success)).toHaveBeenCalled();
  });

  it('reconnecting calls toast.loading', () => {
    wsToast.reconnecting(vi.fn() as any, 1, 10);
    expect(vi.mocked(toast.loading)).toHaveBeenCalled();
  });

  it('failed calls toast.error', () => {
    wsToast.failed(vi.fn() as any);
    expect(vi.mocked(toast.error)).toHaveBeenCalled();
  });

  it('dismiss calls toast.dismiss', () => {
    wsToast.connected(vi.fn() as any);
    wsToast.dismiss();
    expect(vi.mocked(toast.dismiss as any)).toHaveBeenCalled();
  });
});

describe('wsToastOnStatus', () => {
  beforeEach(() => {
    wsToast.connected(vi.fn() as any);
    wsToast.dismiss();
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
    vi.mocked(toast.loading).mockClear();
    vi.mocked(toast.dismiss as any).mockClear?.();
  });

  it('dismisses on off status', () => {
    wsToast.connected(vi.fn() as any);
    wsToastOnStatus(vi.fn() as any, 'connected', 'off');
    expect(vi.mocked(toast.dismiss as any)).toHaveBeenCalled();
  });

  it('shows connected toast when reconnecting->connected', () => {
    wsToastOnStatus(vi.fn() as any, 'reconnecting', 'connected');
    expect(vi.mocked(toast.success)).toHaveBeenCalled();
  });

  it('dismisses when connected from non-reconnecting', () => {
    wsToastOnStatus(vi.fn() as any, 'off', 'connected');
    expect(vi.mocked(toast.success)).not.toHaveBeenCalled();
  });

  it('shows reconnecting toast', () => {
    wsToastOnStatus(vi.fn() as any, 'connected', 'reconnecting', 1, 10);
    expect(vi.mocked(toast.loading)).toHaveBeenCalled();
  });

  it('shows failed toast', () => {
    wsToastOnStatus(vi.fn() as any, 'connected', 'failed');
    expect(vi.mocked(toast.error)).toHaveBeenCalled();
  });
});
