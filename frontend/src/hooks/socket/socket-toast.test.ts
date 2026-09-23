vi.mock("sonner", async () => {
  const { createSonnerMock } = await import("@/tests/mock-factories");
  const m = createSonnerMock();
  m.toast.success = vi.fn(() => "toast-id");
  m.toast.error = vi.fn(() => "toast-id");
  m.toast.loading = vi.fn(() => "toast-id");
  m.toast.dismiss = vi.fn();
  return m;
});

import { toast } from "sonner";
import { wsToast, wsToastOnStatus } from "./socket-toast";

import type { AnyFunction } from "@/tests/test-utils";
describe("wsToast", () => {
  beforeEach(() => {
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
    vi.mocked(toast.loading).mockClear();
    vi.mocked(toast.dismiss).mockClear?.();
  });

  it("connected calls toast.success", () => {
    wsToast.connected(vi.fn() as AnyFunction);
    expect(vi.mocked(toast.success)).toHaveBeenCalled();
  });

  it("reconnecting calls toast.loading", () => {
    wsToast.reconnecting(vi.fn() as AnyFunction, 1, 10);
    expect(vi.mocked(toast.loading)).toHaveBeenCalled();
  });

  it("failed calls toast.error", () => {
    wsToast.failed(vi.fn() as AnyFunction);
    expect(vi.mocked(toast.error)).toHaveBeenCalled();
  });

  it("dismiss calls toast.dismiss", () => {
    wsToast.connected(vi.fn() as AnyFunction);
    wsToast.dismiss();
    expect(vi.mocked(toast.dismiss)).toHaveBeenCalled();
  });
});

describe("wsToastOnStatus", () => {
  beforeEach(() => {
    wsToast.connected(vi.fn() as AnyFunction);
    wsToast.dismiss();
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
    vi.mocked(toast.loading).mockClear();
    vi.mocked(toast.dismiss).mockClear?.();
  });

  it("dismisses on off status", () => {
    wsToast.connected(vi.fn() as AnyFunction);
    wsToastOnStatus(vi.fn() as AnyFunction, "connected", "off");
    expect(vi.mocked(toast.dismiss)).toHaveBeenCalled();
  });

  it("shows connected toast when reconnecting->connected", () => {
    wsToastOnStatus(vi.fn() as AnyFunction, "reconnecting", "connected");
    expect(vi.mocked(toast.success)).toHaveBeenCalled();
  });

  it("dismisses when connected from non-reconnecting", () => {
    wsToastOnStatus(vi.fn() as AnyFunction, "off", "connected");
    expect(vi.mocked(toast.success)).not.toHaveBeenCalled();
  });

  it("shows reconnecting toast", () => {
    wsToastOnStatus(vi.fn() as AnyFunction, "connected", "reconnecting", 1, 10);
    expect(vi.mocked(toast.loading)).toHaveBeenCalled();
  });

  it("shows failed toast", () => {
    wsToastOnStatus(vi.fn() as AnyFunction, "connected", "failed");
    expect(vi.mocked(toast.error)).toHaveBeenCalled();
  });
});
