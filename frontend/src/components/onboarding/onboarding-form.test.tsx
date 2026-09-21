import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { OnboardingForm } from './onboarding-form';
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

describe('OnboardingForm', () => {
  it('renders three step indicators', () => {
    const { container } = render(<OnboardingForm settings={mockSettings} onSubmit={() => {}} />);
    expect(container.querySelectorAll('[data-active]')).toHaveLength(3);
  });

  it('disables Previous button on first step', () => {
    render(<OnboardingForm settings={mockSettings} onSubmit={() => {}} />);
    const prevButton = screen.getByRole('button', { name: /PREVIOUS/ });
    expect(prevButton.hasAttribute('disabled')).toBe(true);
  });

  it('enables Next button after forms are initialized', async () => {
    render(<OnboardingForm settings={mockSettings} onSubmit={() => {}} />);
    const nextButton = screen.getByRole('button', { name: /NEXT/ });
    await waitFor(() => {
      expect(nextButton.hasAttribute('disabled')).toBe(false);
    });
  });

  it('navigates to step 1 when Next is clicked', async () => {
    render(<OnboardingForm settings={mockSettings} onSubmit={() => {}} />);
    const nextButton = screen.getByRole('button', { name: /NEXT/ });
    await waitFor(() => {
      expect(nextButton.hasAttribute('disabled')).toBe(false);
    });
    fireEvent.click(nextButton);
    await waitFor(() => {
      expect(
        screen.getAllByRole('button', { name: /translated:ONBOARDING\.FORM\.CRON_/ }),
      ).toHaveLength(4);
    });
  });

  it('calls onSubmit when Submit is clicked on final step', async () => {
    const onSubmit = vi.fn();
    const { container } = render(<OnboardingForm settings={mockSettings} onSubmit={onSubmit} />);

    // Step 0 → 1
    let nextButton = screen.getByRole('button', { name: /NEXT/ });
    await waitFor(() => {
      expect(nextButton.hasAttribute('disabled')).toBe(false);
    });
    fireEvent.click(nextButton);

    // Step 1 → 2
    await waitFor(() => {
      expect(container.querySelectorAll('[data-slot="toggle"]')).toHaveLength(4);
    });
    nextButton = screen.getByRole('button', { name: /NEXT/ });
    await waitFor(() => {
      expect(nextButton.hasAttribute('disabled')).toBe(false);
    });
    fireEvent.click(nextButton);

    // Step 2 → Submit
    await waitFor(() => {
      expect(screen.getByRole('switch')).toBeTruthy();
    });
    const submitButton = screen.getByRole('button', { name: /SUBMIT/ });
    await waitFor(() => {
      expect(submitButton.hasAttribute('disabled')).toBe(false);
    });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          repo: 'owner/repo',
          cron: '*/10 * * * *',
        }),
      );
    });
  });
});
