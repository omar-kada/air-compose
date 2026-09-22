import { render, screen } from '@testing-library/react';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import type { ReactNode } from 'react';
import type { Settings } from '@/api/api';
import { formSchema, fromSettings, type FormValues } from './settings-form-schema';
import { SettingsForm } from './settings-form';

const mockSettings: Settings = {
  repo: 'owner/repo',
  branch: 'main',
  username: 'user',
  token: 'token',
  cron: '0 */10 * * *',
  notificationURL: 'https://example.com',
  notificationTypes: [],
  retriesOnUnhealthy: 3,
  retryDelay: 60000,
};

const mockUserType = { value: 'LOCAL' as string };

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks', () => ({
  useDeleteAccount: () => ({
    deleteAccount: vi.fn(),
    isPending: false,
  }),
  useUser: () => ({
    data: { type: mockUserType.value, username: 'testuser' },
    isPending: false,
  }),
}));

vi.mock('../view', () => ({
  ConfirmationDialog: ({ children }: { children: ReactNode }) => children,
  NotificationMultiSelect: () => <div data-slot="notification-multiselect" />,
}));

vi.mock('./change-password-dialog', () => ({
  ChangePasswordDialog: ({ children }: { children: ReactNode }) => (
    <div data-slot="change-password-dialog">{children}</div>
  ),
}));

function SettingsFormTestWrapper({ settings }: { settings?: Settings }) {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: fromSettings(settings),
  });
  return <SettingsForm form={form} />;
}

describe('SettingsForm', () => {
  beforeEach(() => {
    mockUserType.value = 'LOCAL';
  });

  it('renders git section input fields', () => {
    render(<SettingsFormTestWrapper settings={mockSettings} />);
    expect(screen.getAllByRole('textbox')).toHaveLength(6);
  });

  it('renders auto-sync section with slider fields', () => {
    const { container } = render(<SettingsFormTestWrapper settings={mockSettings} />);
    expect(container.querySelectorAll('[data-slot="slider"]')).toHaveLength(2);
  });

  it('renders notification section with multi-select', () => {
    const { container } = render(<SettingsFormTestWrapper settings={mockSettings} />);
    expect(container.querySelector('[data-slot="notification-multiselect"]')).not.toBeNull();
  });

  it('shows change password button for local users', () => {
    render(<SettingsFormTestWrapper settings={mockSettings} />);
    expect(screen.getByRole('button', { name: /CHANGE_PASSWORD/ })).toBeTruthy();
  });

  it('hides change password button for non-local users', () => {
    mockUserType.value = 'OIDC';
    render(<SettingsFormTestWrapper settings={mockSettings} />);
    expect(screen.queryByRole('button', { name: /CHANGE_PASSWORD/ })).toBeNull();
  });

  it('renders delete account button', () => {
    render(<SettingsFormTestWrapper settings={mockSettings} />);
    expect(screen.getByRole('button', { name: /DELETE_ACCOUNT/ })).toBeTruthy();
  });
});
