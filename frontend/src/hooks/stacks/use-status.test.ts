const mockFn = vi.hoisted(() => vi.fn());

vi.mock("@/api/api", async (importOriginal) => {
  const original = await importOriginal<typeof import("@/api/api")>();
  return { ...original, getStatusAPIGetQueryOptions: mockFn };
});

import { getStatusQueryOptions } from "./use-status";

describe("getStatusQueryOptions", () => {
  beforeEach(() => mockFn.mockClear());

  it("passes select, staleTime=10000, gcTime to API", () => {
    getStatusQueryOptions();
    expect(mockFn).toHaveBeenCalledWith({
      query: {
        select: expect.any(Function),
        staleTime: 10000,
        gcTime: 600000,
      },
    });
  });

  it("select extracts data.data from response", () => {
    getStatusQueryOptions();
    const select = mockFn.mock.calls[0][0].query.select;
    expect(select({ data: { status: "ok" } })).toEqual({ status: "ok" });
  });
});
