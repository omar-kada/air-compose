const mockFn = vi.hoisted(() =>
  vi.fn((opts: any) => ({
    queryKey: ["user"],
    queryFn: vi.fn().mockResolvedValue({ data: { username: "test" } }),
    ...(opts?.query ?? {}),
  })),
);

vi.mock("@/api/api", async (importOriginal) => {
  const original = await importOriginal<typeof import("@/api/api")>();
  return { ...original, getUserAPIGetQueryOptions: mockFn };
});

import { useUser } from "./use-user";
import { renderHookWithQuery } from "@/tests/test-utils";

describe("useUser", () => {
  it("passes select that returns data.data when username is truthy", () => {
    renderHookWithQuery(() => useUser());
    const select = mockFn.mock.calls[0][0].query.select;
    expect(select({ data: { username: "test" } })).toEqual({
      username: "test",
    });
    expect(select({ data: null })).toBeUndefined();
  });
});
