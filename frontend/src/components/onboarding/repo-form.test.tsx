import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { repoSchema, type RepoFormValues, toRepoFormValues } from './onboarding-schema';
import { RepoForm } from './repo-form';
import type { Settings } from '@/api/api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `translated:${key}`,
    i18n: { language: 'en' },
  }),
}));

vi.mock('@/hooks', () => ({
  useTestConnection: () => ({
    testConnection: vi.fn(),
    isPending: false,
  }),
}));

const mockSettings: Settings = {
  repo: 'owner/repo',
  branch: 'main',
  cron: '*/10 * * * *',
  notificationURL: undefined,
  notificationTypes: [],
  retriesOnUnhealthy: 3,
  retryDelay: 60000,
};

function RepoFormTestWrapper({ settings }: { settings: Settings }) {
  const form = useForm<RepoFormValues>({
    resolver: zodResolver(repoSchema),
    defaultValues: toRepoFormValues(settings),
  });
  return <RepoForm form={form} />;
}

describe('RepoForm', () => {
  it('renders repo and branch input fields', () => {
    render(<RepoFormTestWrapper settings={mockSettings} />);
    expect(screen.getAllByRole('textbox')).toHaveLength(2);
  });

  it('renders an enabled Test Connection button when repo is provided', () => {
    render(<RepoFormTestWrapper settings={mockSettings} />);
    const button = screen.getByRole('button', { name: /TEST_CONNECTION/ });
    expect(button.hasAttribute('disabled')).toBe(false);
  });

  it('disables Test Connection button when repo is empty', () => {
    render(<RepoFormTestWrapper settings={{ ...mockSettings, repo: '' }} />);
    const button = screen.getByRole('button', { name: /TEST_CONNECTION/ });
    expect(button.hasAttribute('disabled')).toBe(true);
  });

  it('shows username and token fields when private repo is enabled', async () => {
    render(<RepoFormTestWrapper settings={mockSettings} />);
    const toggle = screen.getByRole('switch');
    fireEvent.click(toggle);
    await waitFor(() => {
      expect(screen.getAllByRole('textbox')).toHaveLength(4);
    });
  });

  it('hides username and token fields when private repo is disabled', async () => {
    render(
      <RepoFormTestWrapper settings={{ ...mockSettings, token: 'secret', username: 'user' }} />,
    );
    const toggle = screen.getByRole('switch');
    expect(screen.getAllByRole('textbox')).toHaveLength(4);
    fireEvent.click(toggle);
    await waitFor(() => {
      expect(screen.getAllByRole('textbox')).toHaveLength(2);
    });
  });
});
