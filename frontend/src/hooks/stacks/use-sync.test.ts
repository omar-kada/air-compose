const mockToastPromise = vi.hoisted(() => vi.fn());
const mockFn = vi.hoisted(() =>
  vi.fn((opts: any) => ({
    mutationFn: vi.fn().mockResolvedValue({ data: { id: "1" } }),
    ...(opts?.mutation ?? {}),
  })),
);

vi.mock("sonner", async () => {
  const { createSonnerMock } = await import("@/tests/mock-factories");
  const m = createSonnerMock();
  m.toast.promise = mockToastPromise;
  return m;
});

vi.mock("react-i18next", async () => {
  const { createI18nMock } = await import("@/tests/mock-factories");
  return createI18nMock();
});

vi.mock("@/lib", async (importOriginal) => {
  const original = await importOriginal<typeof import("@/lib")>();
  return { ...original, useDeploymentNavigate: vi.fn(() => vi.fn()) };
});

vi.mock("@/api/api", async (importOriginal) => {
  const original = await importOriginal<typeof import("@/api/api")>();
  return {
    ...original,
    getDeployementAPISyncMutationOptions: mockFn,
  };
});

import { useSync } from "./use-sync";
import { renderHookWithQuery } from "@/tests/test-utils";
import type { AnyFunction } from "@/tests/test-utils";

describe("useSync", () => {
  it("returns sync function", () => {
    const { result } = renderHookWithQuery(() => useSync());
    expect(result.current.sync).toBeInstanceOf(Function);
  });

  it("sync calls toast.promise", () => {
    const { result } = renderHookWithQuery(() => useSync());
    (result.current.sync as AnyFunction)();
    expect(mockToastPromise).toHaveBeenCalled();
  });
});
