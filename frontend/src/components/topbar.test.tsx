import { render, screen, fireEvent } from '@testing-library/react';
import { Topbar } from './topbar';

const { mockUseUser, mockUseUnreadCount, mockLogout, mockSetTheme } = vi.hoisted(() => ({
  mockUseUser: vi.fn(),
  mockUseUnreadCount: vi.fn(),
  mockLogout: vi.fn(),
  mockSetTheme: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `t:${key}`,
  }),
}));

vi.mock('@/hooks', () => ({
  useLogout: () => ({ logout: mockLogout }),
  useUnreadNotificationCount: mockUseUnreadCount,
  useUser: mockUseUser,
}));

vi.mock('@/hooks/theme-provider', () => ({
  useTheme: () => ({ theme: 'light', setTheme: mockSetTheme }),
}));

vi.mock('@/lib', () => ({
  cn: (...args: unknown[]) => args.filter(Boolean).join(' '),
}));

vi.mock('./notifications', () => ({
  NotificationSheet: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="notification-sheet">{children}</div>
  ),
}));

vi.mock('./settings', () => ({
  SettingsSheet: () => <div data-slot="settings-sheet" />,
}));

vi.mock('./ui/dropdown-menu', () => {
  const dc = ({ children }: { children: React.ReactNode }) => (
    <div data-slot="dropdown-content">{children}</div>
  );
  return {
    DropdownMenu: ({ children }: { children: React.ReactNode }) => (
      <div data-slot="dropdown-menu">{children}</div>
    ),
    DropdownMenuTrigger: ({ children }: { children: React.ReactNode }) => (
      <div data-slot="dropdown-trigger">{children}</div>
    ),
    DropdownMenuContent: dc,
    DropdownMenuGroup: ({ children }: { children: React.ReactNode }) => (
      <div data-slot="dropdown-group">{children}</div>
    ),
    DropdownMenuItem: ({
      children,
      onClick,
      onSelect,
    }: {
      children: React.ReactNode;
      onClick?: () => void;
      onSelect?: () => void;
    }) => (
      <div data-slot="dropdown-item" onClick={onClick || onSelect}>
        {children}
      </div>
    ),
    DropdownMenuLabel: ({ children }: { children: React.ReactNode }) => (
      <div data-slot="dropdown-label">{children}</div>
    ),
    DropdownMenuSeparator: () => <div data-slot="dropdown-separator" />,
  };
});

vi.mock('./ui/switch', () => ({
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

describe('Topbar', () => {
  beforeEach(() => {
    mockUseUser.mockReturnValue({ data: { username: 'testuser', email: 'test@test.com' } });
    mockUseUnreadCount.mockReturnValue(0);
    mockLogout.mockClear();
    mockSetTheme.mockClear();
  });

  it('renders the logo', () => {
    const { container } = render(<Topbar />);
    expect(container.textContent).toContain('AirCompose');
  });

  it('shows unread count badge when count is greater than zero', () => {
    mockUseUnreadCount.mockReturnValue(5);
    render(<Topbar />);
    expect(screen.getByText('5')).toBeTruthy();
  });

  it('hides unread count badge when count is zero', () => {
    mockUseUnreadCount.mockReturnValue(0);
    render(<Topbar />);
    expect(screen.queryByText('0')).toBeNull();
  });

  it('renders user avatar when user is logged in', () => {
    render(<Topbar />);
    expect(document.querySelector('[data-slot="avatar"]')).not.toBeNull();
  });

  it('hides user section when user is not logged in', () => {
    mockUseUser.mockReturnValueOnce({ data: null });
    const { container } = render(<Topbar />);
    expect(container.querySelector('[data-slot="avatar"]')).toBeNull();
  });

  it('calls setTheme when dark mode toggle is clicked', () => {
    render(<Topbar />);
    fireEvent.click(
      screen.getByText('t:MENU.DARK_MODE').closest('[data-slot="dropdown-item"]') as HTMLElement,
    );
    expect(mockSetTheme).toHaveBeenCalledWith('dark');
  });
});
