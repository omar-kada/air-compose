const mockList = vi.hoisted(() =>
  vi.fn(() => ({ data: { items: [], pageInfo: { endCursor: "" } } })),
);
const mockKey = vi.hoisted(() => vi.fn(() => ["notifications-key"]));

import type { AnyFunction } from "@/tests/test-utils";
import type { QueryClient } from "@tanstack/react-query";
vi.mock("@/api/api", async (importOriginal) => {
  const original = await importOriginal<typeof import("@/api/api")>();
  return {
    ...original,
    notificationsAPIList: mockList,
    getNotificationsAPIListQueryKey: mockKey,
  };
});

import {
  getNotificationsQueryOptions,
  refetchNotifications,
} from "./use-notifications";

describe("getNotificationsQueryOptions", () => {
  it("returns query options with correct structure", () => {
    const options = getNotificationsQueryOptions();
    expect(options.queryKey).toEqual(["notifications-key"]);
    expect(options.queryFn).toBeInstanceOf(Function);
    expect(options.select).toBeInstanceOf(Function);
    expect(options.getNextPageParam).toBeInstanceOf(Function);
    expect(options.initialPageParam).toEqual({ limit: 10, offset: "" });
    expect(options.gcTime).toBe(600000);
    expect(options.refetchOnMount).toBe(true);
  });

  it("queryFn calls notificationsAPIList with pageParam", async () => {
    const options = getNotificationsQueryOptions();
    const result = await (options.queryFn! as AnyFunction)({
      pageParam: { limit: 10, offset: "abc" },
    });
    expect(mockList).toHaveBeenCalledWith({ limit: 10, offset: "abc" });
    expect(result).toEqual({
      data: { items: [], pageInfo: { endCursor: "" } },
    });
  });

  it("select flattens pages into Event[]", () => {
    const options = getNotificationsQueryOptions();
    const data = {
      pageParams: [],
      pages: [
        { data: { items: [{ id: 1 }, { id: 2 }] } },
        { data: { items: [{ id: 3 }] } },
      ],
    };
    expect((options.select! as AnyFunction)(data)).toEqual([
      { id: 1 },
      { id: 2 },
      { id: 3 },
    ]);
  });

  it("getNextPageParam returns undefined when endCursor is empty", () => {
    const options = getNotificationsQueryOptions();
    const lastPage = { data: { items: [], pageInfo: { endCursor: "" } } };
    expect(
      (options.getNextPageParam! as AnyFunction)(lastPage),
    ).toBeUndefined();
  });

  it("getNextPageParam returns offset object when endCursor is present", () => {
    const options = getNotificationsQueryOptions();
    const lastPage = { data: { items: [], pageInfo: { endCursor: "xyz" } } };
    expect((options.getNextPageParam! as AnyFunction)(lastPage)).toEqual({
      limit: 10,
      offset: "xyz",
    });
  });

  it("refetchNotifications calls refetchQueries", () => {
    const mockClient = { refetchQueries: vi.fn() } as unknown as QueryClient;
    refetchNotifications(mockClient);
    expect(mockClient.refetchQueries).toHaveBeenCalled();
  });
});
