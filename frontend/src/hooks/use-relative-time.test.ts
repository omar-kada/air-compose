import { renderHook } from "@testing-library/react";
import { useRelativeTime } from "./use-relative-time";

vi.mock("react-i18next", async () => {
  const { createI18nMock } = await import("@/tests/mock-factories");
  return createI18nMock();
});

describe("useRelativeTime", () => {
  const realDateNow = Date.now;
  const realSetInterval = globalThis.setInterval;
  const realClearInterval = globalThis.clearInterval;

  beforeEach(() => {
    const fixedNow = new Date("2024-01-15T12:00:00.000Z").getTime();
    Date.now = vi.fn(() => fixedNow);
    globalThis.setInterval = vi.fn() as unknown as typeof setInterval;
    globalThis.clearInterval = vi.fn() as unknown as typeof clearInterval;
  });

  afterEach(() => {
    Date.now = realDateNow;
    globalThis.setInterval = realSetInterval;
    globalThis.clearInterval = realClearInterval;
  });

  it("returns null for empty target", () => {
    const { result } = renderHook(() => useRelativeTime(""));
    expect(result.current).toBeNull();
  });

  it("returns null for pre-1970 date", () => {
    const { result } = renderHook(() => useRelativeTime("1960-01-01"));
    expect(result.current).toBeNull();
  });

  it('returns "just now" for recent past (within 60s)', () => {
    const fixedNow = new Date("2024-01-15T12:00:00.000Z").getTime();
    const target = new Date(fixedNow - 30_000).toISOString();
    const { result } = renderHook(() => useRelativeTime(target));
    expect(result.current).toBe("translated:TIME.JUST_NOW");
  });

  it('returns "in a few seconds" for recent future (within 60s)', () => {
    const fixedNow = new Date("2024-01-15T12:00:00.000Z").getTime();
    const target = new Date(fixedNow + 30_000).toISOString();
    const { result } = renderHook(() => useRelativeTime(target));
    expect(result.current).toBe("translated:TIME.IN_FEW_SECONDS");
  });

  it("returns humanized duration for dates >60s in the past", () => {
    const fixedNow = new Date("2024-01-15T12:00:00.000Z").getTime();
    const target = new Date(fixedNow - 5 * 60_000).toISOString();
    const { result } = renderHook(() => useRelativeTime(target));
    expect(result.current).not.toBeNull();
  });

  it("returns humanized duration for dates >60s in the future", () => {
    const fixedNow = new Date("2024-01-15T12:00:00.000Z").getTime();
    const target = new Date(fixedNow + 5 * 60_000).toISOString();
    const { result } = renderHook(() => useRelativeTime(target));
    expect(result.current).not.toBeNull();
  });

  it("sets up interval for periodic updates", () => {
    const fixedNow = new Date("2024-01-15T12:00:00.000Z").getTime();
    const target = new Date(fixedNow + 5 * 60_000).toISOString();
    renderHook(() => useRelativeTime(target));
    expect(globalThis.setInterval).toHaveBeenCalled();
  });

  it("cleans up interval on unmount", () => {
    const fixedNow = new Date("2024-01-15T12:00:00.000Z").getTime();
    const target = new Date(fixedNow + 5 * 60_000).toISOString();
    const { unmount } = renderHook(() => useRelativeTime(target));
    unmount();
    expect(globalThis.clearInterval).toHaveBeenCalled();
  });
});
