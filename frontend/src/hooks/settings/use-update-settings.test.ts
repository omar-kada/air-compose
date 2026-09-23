const mockToastPromise = vi.hoisted(() => vi.fn());
const mockFn = vi.hoisted(() =>
  vi.fn((opts: any) => ({
    mutationFn: vi.fn().mockResolvedValue({ data: { success: true } }),
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

vi.mock("@/api/api", async (importOriginal) => {
  const original = await importOriginal<typeof import("@/api/api")>();
  return {
    ...original,
    getSettingsAPISetMutationOptions: mockFn,
    getSettingsAPIGetQueryKey: vi.fn(() => ["settings-key"]),
  };
});

import { useUpdateSettings } from "./use-update-settings";
import { renderHookWithQuery } from "@/tests/test-utils";
import type { AnyFunction } from "@/tests/test-utils";

describe("useUpdateSettings", () => {
  it("returns updateSettings function", () => {
    const { result } = renderHookWithQuery(() => useUpdateSettings());
    expect(result.current.updateSettings).toBeInstanceOf(Function);
  });

  it("updateSettings calls toast.promise", () => {
    const { result } = renderHookWithQuery(() => useUpdateSettings());
    (result.current.updateSettings as AnyFunction)({ theme: "dark" });
    expect(mockToastPromise).toHaveBeenCalled();
  });
});
