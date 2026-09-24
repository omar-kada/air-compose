import { act } from "@testing-library/react";
import type { QueryClient } from "@tanstack/react-query";
import {
  useUnreadNotificationCount,
  incrementUnreadCount,
  useResetUnreadCount,
  UNREAD_NOTIFICATION_COUNT_KEY,
} from "./use-unread-notifications";
import { renderHookWithQuery } from "@/tests/test-utils";

describe("useUnreadNotificationCount", () => {
  it("returns initial data of 0", () => {
    const { result } = renderHookWithQuery(() => useUnreadNotificationCount());
    expect(result.current).toBe(0);
  });
});

describe("incrementUnreadCount", () => {
  it("increments the unread count", () => {
    const setQueryData = vi.fn();
    incrementUnreadCount({ setQueryData } as unknown as QueryClient);
    expect(setQueryData).toHaveBeenCalledWith(
      UNREAD_NOTIFICATION_COUNT_KEY,
      expect.any(Function),
    );
    const updater = setQueryData.mock.calls[0][1];
    expect(updater(5)).toBe(6);
    expect(updater(undefined)).toBe(1);
  });
});

describe("useResetUnreadCount", () => {
  it("returns a function that resets count to 0", () => {
    const { result, queryClient } = renderHookWithQuery(() =>
      useResetUnreadCount(),
    );
    const spy = vi.spyOn(queryClient, "setQueryData");
    act(() => {
      result.current();
    });
    expect(spy).toHaveBeenCalledWith(UNREAD_NOTIFICATION_COUNT_KEY, 0);
  });
});
