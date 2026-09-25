import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Topbar } from "./topbar";
import { vi } from "vitest";

const mockLogout = vi.fn();
const mockSetTheme = vi.fn();
const mockResetUnreadCount = vi.fn();

const mockUser = { username: "testuser", email: "test@test.com" };

const translations: Record<string, string> = {
  APP_NAME: "Air",
  "MENU.LOGGED_AS": "Logged in as:",
  "MENU.DARK_MODE": "Dark mode",
  "MENU.SETTINGS": "Settings",
  "ACTION.SIGN_OUT": "Sign out",
  "NOTIFICATIONS.NOTIFICATIONS": "Notifications",
  "NOTIFICATIONS.DESCRIPTION": "See your notifications below",
  "NOTIFICATIONS.EMPTY": "No notifications",
};

vi.mock("react-intersection-observer", () => ({
  useInView: () => ({ ref: vi.fn(), inView: false }),
}));

vi.mock("@/lib/navigation", () => ({
  useDeploymentNavigate: () => vi.fn(),
}));

vi.mock("react-router-dom", () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock("react-i18next", () => ({
  useTranslation: () => ({
    t: (key: string) => translations[key] ?? key,
    i18n: { language: "en" },
  }),
}));

vi.mock("@/hooks", () => ({
  useLogout: () => ({ logout: mockLogout }),
  useUnreadNotificationCount: () => 0,
  useUser: () => ({ data: mockUser, isPending: false }),
  useResetUnreadCount: () => ({ mutate: mockResetUnreadCount }),
  useFilteredQuery: () => ({ data: undefined, error: null, isPending: false }),
  getNotificationsQueryOptions: () => ({
    queryKey: ["notifications"],
    queryFn: () => ({ pages: [] }),
  }),
  getFeaturesQueryOptions: () => ({
    queryKey: ["features"],
    queryFn: () => ({}),
  }),
  getSettingsQueryOptions: () => ({
    queryKey: ["settings"],
    queryFn: () => ({}),
  }),
  useUpdateSettings: () => ({ updateSettings: mockLogout }),
}));

vi.mock("@/hooks/theme-provider", () => ({
  useTheme: () => ({ theme: "light", setTheme: mockSetTheme }),
}));

vi.mock("@tanstack/react-query", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@tanstack/react-query")>();
  return {
    ...actual,
    useInfiniteQuery: () => ({
      data: [],
      fetchNextPage: vi.fn(),
      hasNextPage: false,
      isFetchingNextPage: false,
      isLoading: false,
      isError: false,
      isPending: false,
      error: null,
    }),
  };
});

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false, staleTime: 0 } },
});

const renderTopbar = () =>
  render(
    <QueryClientProvider client={queryClient}>
      <Topbar />
    </QueryClientProvider>,
  );

const openUserDropDown = async () => {
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", { name: /^T$/ }));
};

const findMenuItem = (label: string) => {
  const needle = label.toLowerCase().replace(/\s+/g, "");
  return screen.getAllByRole("menuitem").find((item) => {
    const hay = item.textContent?.toLowerCase().replace(/\s+/g, "") ?? "";
    return hay.includes(needle);
  });
};

describe("Topbar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    queryClient.clear();
  });

  it("T1: renders the app logo", () => {
    const { container } = renderTopbar();
    expect(container.textContent).toContain("AirCompose");
  });

  it("T2: shows notifications sheet trigger (bell button)", () => {
    renderTopbar();
    // The bell button is the first button in the header (NotificationSheet trigger)
    const bellButton = screen.getAllByRole("button")[0];
    expect(bellButton.hasAttribute("aria-haspopup")).toBe(true);
  });

  it("T3: shows logged in user", async () => {
    renderTopbar();
    await openUserDropDown();
    expect(
      screen.getByText((content) => content.includes("testuser")),
    ).toBeTruthy();
  });

  it("T4: renders settings menu item", async () => {
    renderTopbar();
    await openUserDropDown();
    expect(findMenuItem("settings")).toBeTruthy();
  });

  it("T5: renders dark mode toggle", async () => {
    renderTopbar();
    await openUserDropDown();
    expect(findMenuItem("dark mode")).toBeTruthy();
  });

  it("T6: toggles theme when clicking dark mode", async () => {
    renderTopbar();
    await openUserDropDown();
    const darkModeItem = findMenuItem("dark mode");
    if (!darkModeItem) throw new Error("dark mode menu item not found");
    await userEvent.click(darkModeItem);
    expect(mockSetTheme).toHaveBeenCalledWith("dark");
  });

  it("T7: notifications sheet opens on click", async () => {
    renderTopbar();
    const bellButton = screen.getAllByRole("button")[0];
    await userEvent.click(bellButton);
    expect(screen.getByText("Notifications")).toBeTruthy();
  });

  it("T8: signs out when clicking sign out", async () => {
    renderTopbar();
    await openUserDropDown();
    const signOutItem = findMenuItem("sign out");
    if (!signOutItem) throw new Error("sign out menu item not found");
    await userEvent.click(signOutItem);
    expect(mockLogout).toHaveBeenCalled();
  });
});
