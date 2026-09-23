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
    getAuthAPIRegisterMutationOptions: mockFn,
    getAuthAPIRegisteredQueryOptions: vi.fn(() => ({
      queryKey: ["registered"],
    })),
    getUserAPIGetQueryOptions: vi.fn(() => ({ queryKey: ["user"] })),
  };
});

import { useRegister } from "./use-register";
import { renderHookWithQuery } from "@/tests/test-utils";
import type { AnyFunction } from "@/tests/test-utils";

describe("useRegister", () => {
  it("returns register function", () => {
    const { result } = renderHookWithQuery(() => useRegister());
    expect(result.current.register).toBeInstanceOf(Function);
  });

  it("register calls toast.promise", () => {
    const { result } = renderHookWithQuery(() => useRegister());
    (result.current.register as AnyFunction)({
      username: "test",
      password: "test",
    });
    expect(mockToastPromise).toHaveBeenCalled();
  });
});
