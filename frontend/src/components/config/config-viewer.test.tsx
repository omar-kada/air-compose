import { render, screen, fireEvent, act } from "@testing-library/react";
import { ConfigViewer } from "./config-viewer";

vi.mock("react-i18next", () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock("@/hooks/theme-provider", () => ({
  useTheme: () => ({ theme: "light" as const, setTheme: vi.fn() }),
}));
vi.mock("react-syntax-highlighter", () => ({
  Light: ({ children }: { children: string }) => (
    <div data-testid="syntax-highlighter">{children}</div>
  ),
}));
vi.mock("react-syntax-highlighter/dist/esm/styles/hljs", () => ({
  tomorrow: {},
  tomorrowNightBlue: {},
}));

describe("ConfigViewer", () => {
  const mockWriteText = vi.fn();
  let originalClipboard: typeof navigator.clipboard;

  beforeEach(() => {
    vi.useFakeTimers();
    mockWriteText.mockResolvedValue(undefined);
    originalClipboard = navigator.clipboard;
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText: mockWriteText },
      configurable: true,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    Object.defineProperty(navigator, "clipboard", {
      value: originalClipboard,
      configurable: true,
    });
  });

  describe("rendering", () => {
    it("renders the yaml text in the syntax highlighter", () => {
      render(<ConfigViewer text="globalVariables: {}" />);
      const highlighter = screen.getByTestId("syntax-highlighter");
      expect(highlighter.textContent).toContain("globalVariables: {}");
    });

    it("renders the config.yaml title", () => {
      render(<ConfigViewer text="test" />);
      expect(screen.getByText("config.yaml")).not.toBeNull();
    });

    it("renders only the copy button when onClose is not provided", () => {
      render(<ConfigViewer text="test" />);
      expect(screen.getAllByRole("button")).toHaveLength(1);
    });

    it("renders both close and copy buttons when onClose is provided", () => {
      render(<ConfigViewer text="test" onClose={vi.fn()} />);
      expect(screen.getAllByRole("button")).toHaveLength(2);
    });
  });

  describe("copy functionality", () => {
    it("calls navigator.clipboard.writeText with the text", () => {
      render(<ConfigViewer text="yaml-content" />);
      fireEvent.click(screen.getByRole("button"));
      expect(mockWriteText).toHaveBeenCalledWith("yaml-content");
    });

    it("shows COPIED text immediately after clicking copy", () => {
      render(<ConfigViewer text="yaml-content" />);
      fireEvent.click(screen.getByRole("button"));
      expect(screen.getByText("ALERT.COPIED")).not.toBeNull();
    });

    it("hides COPIED text after 3 seconds", () => {
      render(<ConfigViewer text="yaml-content" />);
      fireEvent.click(screen.getByRole("button"));
      expect(screen.getByText("ALERT.COPIED")).not.toBeNull();
      act(() => {
        vi.advanceTimersByTime(3000);
      });
      expect(screen.queryByText("ALERT.COPIED")).toBeNull();
    });
  });

  describe("close button", () => {
    it("calls onClose when clicked", () => {
      const onClose = vi.fn();
      render(<ConfigViewer text="test" onClose={onClose} />);
      const buttons = screen.getAllByRole("button");
      // close button is the first button rendered in DOM order
      fireEvent.click(buttons[0]);
      expect(onClose).toHaveBeenCalledTimes(1);
    });

    it("does not render the close button when onClose is not provided", () => {
      render(<ConfigViewer text="test" />);
      expect(screen.queryAllByRole("button")).toHaveLength(1);
    });
  });
});
