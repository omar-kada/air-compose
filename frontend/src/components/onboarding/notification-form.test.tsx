import { render } from "@testing-library/react";
import { useForm } from "react-hook-form";
import type { NotificationFormValues } from "./onboarding-schema";
import { NotificationForm } from "./notification-form";

vi.mock("react-i18next", () => ({
  useTranslation: () => ({
    t: (key: string) => `t:${key}`,
  }),
}));

vi.mock("../ui/field", () => ({
  Field: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="field">{children}</div>
  ),
  FieldDescription: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="field-description">{children}</div>
  ),
  FieldError: () => null,
  FieldGroup: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="field-group">{children}</div>
  ),
  FieldTitle: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="field-title">{children}</div>
  ),
}));

vi.mock("../ui/input", () => ({
  Input: (props: React.ComponentProps<"input">) => (
    <input data-slot="input" {...props} />
  ),
}));

vi.mock("../ui/switch", () => ({
  Switch: ({
    checked,
    onCheckedChange,
  }: {
    checked?: boolean;
    onCheckedChange?: (checked: boolean) => void;
  }) => (
    <input
      type="checkbox"
      data-slot="switch"
      checked={checked}
      onChange={(e) => onCheckedChange?.(e.target.checked)}
    />
  ),
}));

vi.mock("../view", () => ({
  NotificationMultiSelect: () => <div data-slot="notification-multi-select" />,
}));

function TestWrapper({
  enableNotifications = false,
}: {
  enableNotifications?: boolean;
}) {
  const form = useForm<NotificationFormValues>({
    defaultValues: {
      enableNotifications,
      notificationURL: "",
      notificationTypes: [],
    },
  });
  return <NotificationForm form={form} />;
}

describe("NotificationForm", () => {
  it("renders form description text", () => {
    const { container } = render(<TestWrapper />);
    expect(container.textContent).toContain(
      "t:ONBOARDING.FORM.NOTIFICATION_FORM_DESCRIPTION",
    );
  });

  it("renders enableNotifications switch", () => {
    const { container } = render(<TestWrapper />);
    expect(container.querySelector('[data-slot="switch"]')).not.toBeNull();
  });

  it("renders notification fields when notifications are enabled", () => {
    const { container } = render(<TestWrapper enableNotifications />);
    expect(container.querySelector('[data-slot="input"]')).not.toBeNull();
    expect(
      container.querySelector('[data-slot="notification-multi-select"]'),
    ).not.toBeNull();
  });

  it("does not render notification fields when notifications are disabled", () => {
    const { container } = render(<TestWrapper enableNotifications={false} />);
    expect(container.querySelector('[data-slot="input"]')).toBeNull();
    expect(
      container.querySelector('[data-slot="notification-multi-select"]'),
    ).toBeNull();
  });
});
