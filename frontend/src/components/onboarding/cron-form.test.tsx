import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { cronSchema, type CronFormValues, toCronFormValues } from './onboarding-schema';
import { CronForm } from './cron-form';
import type { Settings } from '@/api/api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `translated:${key}`,
    i18n: { language: 'en' },
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

function CronFormTestWrapper({ settings }: { settings: Settings }) {
  const form = useForm<CronFormValues>({
    resolver: zodResolver(cronSchema),
    defaultValues: toCronFormValues(settings),
  });
  return <CronForm form={form} />;
}

describe('CronForm', () => {
  it('renders four predefined cron toggle buttons', () => {
    render(<CronFormTestWrapper settings={mockSettings} />);
    expect(
      screen.getAllByRole('button', { name: /translated:ONBOARDING\.FORM\.CRON_/ }),
    ).toHaveLength(4);
  });

  it('highlights the toggle matching the current cron value', () => {
    const { container } = render(<CronFormTestWrapper settings={mockSettings} />);
    const toggles = container.querySelectorAll('[data-slot="toggle"]');
    expect(toggles[1].getAttribute('aria-pressed')).toBe('true');
  });

  it('renders a cron text input field', () => {
    const { container } = render(<CronFormTestWrapper settings={mockSettings} />);
    expect(container.querySelectorAll('[data-slot="input"]')).toHaveLength(1);
  });

  it('sets cron value when a different toggle is clicked', async () => {
    render(<CronFormTestWrapper settings={mockSettings} />);
    const hourlyToggle = screen.getByRole('button', {
      name: 'translated:ONBOARDING.FORM.CRON_HOURLY',
    });
    fireEvent.click(hourlyToggle);
    await waitFor(() => {
      const toggles = screen.getAllByRole('button', {
        name: /translated:ONBOARDING\.FORM\.CRON_/,
      });
      expect(toggles[2].getAttribute('aria-pressed')).toBe('true');
      expect(toggles[1].getAttribute('aria-pressed')).toBe('false');
    });
  });

  it('clears toggle highlight when custom cron value is typed', async () => {
    const { container } = render(<CronFormTestWrapper settings={mockSettings} />);
    const cronInput = container.querySelector('[data-slot="input"]') as HTMLInputElement;
    fireEvent.change(cronInput, { target: { value: '*/5 * * * *' } });
    await waitFor(() => {
      const toggles = container.querySelectorAll('[data-slot="toggle"]');
      expect(toggles[1].getAttribute('aria-pressed')).toBe('false');
    });
  });
});
