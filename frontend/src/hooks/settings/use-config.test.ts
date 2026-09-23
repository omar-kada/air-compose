const mockFn = vi.hoisted(() => vi.fn());

vi.mock("@/api/api", async (importOriginal) => {
  const original = await importOriginal<typeof import("@/api/api")>();
  return { ...original, getConfigAPIGetQueryOptions: mockFn };
});

import { getConfigQueryOptions } from "./use-config";

describe("getConfigQueryOptions", () => {
  beforeEach(() => mockFn.mockClear());

  it("passes select, gcTime, and enabled to API", () => {
    getConfigQueryOptions({ enabled: true });
    expect(mockFn).toHaveBeenCalledWith({
      query: {
        select: expect.any(Function),
        gcTime: 600000,
        enabled: true,
      },
    });
  });

  it("passes enabled=false through", () => {
    getConfigQueryOptions({ enabled: false });
    expect(mockFn).toHaveBeenCalledWith(
      expect.objectContaining({
        query: expect.objectContaining({ enabled: false }),
      }),
    );
  });

  it("select extracts data.data from response", () => {
    getConfigQueryOptions({ enabled: true });
    const select = mockFn.mock.calls[0][0].query.select;
    expect(select({ data: { repo: "test" } })).toEqual({ repo: "test" });
  });
});
